// ABOUTME: This file contains tests for MicroChess types and game state.
// ABOUTME: It verifies piece setup, board initialization, and piece finding logic.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	var buf bytes.Buffer
	game := NewGame(&buf)
	assert.NotNil(t, game, "NewGame() should not return nil")

	// NewGame() should start with empty board (matching original assembly behavior)
	// Only one black pawn at 00 (the "garbage" state)
	assert.Equal(t, board.Square(0x00), game.BK[8], "Black pawn 1 should be at position 0x00")

	// All other pieces should be off-board (0xFF)
	for i := Piece(0); i < 16; i++ {
		if i == 8 {
			continue // Skip the one black pawn we set
		}
		assert.Equal(t, board.Square(0xFF), game.Board[i], "White piece %d should be off-board (0xFF)", i)
		if i != 8 {
			assert.Equal(t, board.Square(0xFF), game.BK[i], "Black piece %d should be off-board (0xFF)", i)
		}
	}

	assert.False(t, game.Reversed, "New game should not be reversed")
}

func TestSetupBoard(t *testing.T) {
	game := &GameState{}
	game.SetupBoard()

	// Check a few key pieces
	expectedWhite := map[Piece]board.Square{
		PieceKing:  0x03,
		PieceQueen: 0x04,
		PieceRook1: 0x00,
		PieceRook2: 0x07,
		PiecePawn1: 0x10,
		PiecePawn8: 0x13,
	}

	for piece, expected := range expectedWhite {
		assert.Equal(t, expected, game.Board[piece], "White piece %d position", piece)
	}

	expectedBlack := map[Piece]board.Square{
		PieceKing:  0x73,
		PieceQueen: 0x74,
		PieceRook1: 0x70,
		PieceRook2: 0x77,
	}

	for piece, expected := range expectedBlack {
		assert.Equal(t, expected, game.BK[piece], "Black piece %d position", piece)
	}
}

func TestFindPieceAt(t *testing.T) {
	var buf bytes.Buffer
	game := NewGame(&buf)
	game.SetupBoard() // Setup the board first

	tests := []struct {
		name    string
		square  board.Square
		piece   Piece
		found   bool
		isWhite bool
	}{
		{"white king on d1", 0x03, PieceKing, true, true},
		{"white queen on e1", 0x04, PieceQueen, true, true},
		{"black king on d8", 0x73, PieceKing, true, false},
		{"empty square e4", 0x34, NoPiece, false, false},
		{"empty square e5", 0x44, NoPiece, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			piece, found, isWhite := game.FindPieceAt(tt.square)
			assert.Equal(t, tt.found, found, "FindPieceAt(0x%02X) found", tt.square)
			if found {
				assert.Equal(t, tt.piece, piece, "FindPieceAt(0x%02X) piece", tt.square)
				assert.Equal(t, tt.isWhite, isWhite, "FindPieceAt(0x%02X) isWhite", tt.square)
			}
		})
	}
}

func TestGetPieceChar(t *testing.T) {
	tests := []struct {
		name    string
		piece   Piece
		isWhite bool
		want    string
	}{
		{"white king", PieceKing, true, "WK"},
		{"black king", PieceKing, false, "BK"},
		{"white queen", PieceQueen, true, "WQ"},
		{"white rook", PieceRook1, true, "WR"},
		{"black bishop", PieceBishop2, false, "BB"},
		{"white knight", PieceKnight1, true, "WN"},
		{"black pawn", PiecePawn1, false, "BP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPieceChar(tt.piece, tt.isWhite)
			assert.Equal(t, tt.want, got, "GetPieceChar(%d, %v)", tt.piece, tt.isWhite)
		})
	}
}

func TestReverse(t *testing.T) {
	var buf bytes.Buffer

	t.Run("coordinate transformation", func(t *testing.T) {
		// Test that 0x77 - coordinate transformation works correctly
		// Key principle: square + (0x77 - square) = 0x77
		testSquares := []board.Square{0x00, 0x07, 0x34, 0x70, 0x77}
		for _, sq := range testSquares {
			reversed := 0x77 - sq
			// Verify that reversing twice gets back to original
			doubleReversed := 0x77 - reversed
			assert.Equal(t, sq, doubleReversed, "0x77 - (0x77 - 0x%02X) should equal 0x%02X", sq, sq)
		}
	})

	t.Run("single reverse swaps boards and transforms coordinates", func(t *testing.T) {
		game := NewGame(&buf)
		game.SetupBoard()

		// Save original positions
		originalWhiteKing := game.Board[PieceKing]
		originalBlackKing := game.BK[PieceKing]

		// Perform reverse
		game.Reverse()

		// After reverse:
		// - Board and BK should be swapped
		// - Coordinates should be transformed (0x77 - original)
		expectedWhiteKingPos := 0x77 - originalBlackKing
		expectedBlackKingPos := 0x77 - originalWhiteKing

		assert.Equal(t, expectedWhiteKingPos, game.Board[PieceKing],
			"After reverse, white king should be at transformed black king position")
		assert.Equal(t, expectedBlackKingPos, game.BK[PieceKing],
			"After reverse, black king should be at transformed white king position")

		// Verify Reversed flag
		assert.True(t, game.Reversed, "Reversed flag should be true after first reverse")
	})

	t.Run("double reverse restores original position", func(t *testing.T) {
		game := NewGame(&buf)
		game.SetupBoard()

		// Save all original positions
		originalBoard := game.Board
		originalBK := game.BK

		// Reverse twice
		game.Reverse()
		game.Reverse()

		// Should be back to original
		assert.Equal(t, originalBoard, game.Board, "Double reverse should restore Board")
		assert.Equal(t, originalBK, game.BK, "Double reverse should restore BK")
		assert.False(t, game.Reversed, "Reversed flag should be false after double reverse")
	})

	t.Run("reverse with all pieces", func(t *testing.T) {
		game := NewGame(&buf)
		game.SetupBoard()

		// Verify that all 16 pieces in each board are handled
		for i := Piece(0); i < 16; i++ {
			originalWhite := game.Board[i]
			originalBlack := game.BK[i]

			game.Reverse()

			expectedWhite := 0x77 - originalBlack
			expectedBlack := 0x77 - originalWhite

			assert.Equal(t, expectedWhite, game.Board[i],
				"Piece %d: Board[%d] should be transformed", i, i)
			assert.Equal(t, expectedBlack, game.BK[i],
				"Piece %d: BK[%d] should be transformed", i, i)

			// Reverse back for next iteration
			game.Reverse()
		}
	})

	t.Run("specific coordinate examples from original", func(t *testing.T) {
		// Test key transformations from the original assembly
		// White pieces start on rank 0-1, black on rank 6-7
		// After reverse, they should be flipped
		game := NewGame(&buf)
		game.SetupBoard()

		// White king starts at 0x03 (d1)
		assert.Equal(t, board.Square(0x03), game.Board[PieceKing])
		// Black king starts at 0x73 (d8)
		assert.Equal(t, board.Square(0x73), game.BK[PieceKing])

		game.Reverse()

		// After reverse:
		// White king should be at 0x77 - 0x73 = 0x04 (e1)
		assert.Equal(t, board.Square(0x04), game.Board[PieceKing])
		// Black king should be at 0x77 - 0x03 = 0x74 (e8)
		assert.Equal(t, board.Square(0x74), game.BK[PieceKing])
	})
}

func TestHandleCommandE(t *testing.T) {
	t.Run("E command calls Reverse", func(t *testing.T) {
		var buf bytes.Buffer
		game := NewGame(&buf)
		game.SetupBoard()

		originalWhiteKing := game.Board[PieceKing]
		originalBlackKing := game.BK[PieceKing]

		// Execute E command
		result := game.HandleCharacter('E')

		// Should return true (continue)
		assert.True(t, result, "HandleCommand('E') should return true")

		// Should have reversed the board
		assert.NotEqual(t, originalWhiteKing, game.Board[PieceKing],
			"E command should change piece positions")
		assert.Equal(t, 0x77-originalBlackKing, game.Board[PieceKing],
			"E command should transform coordinates correctly")

		// LED display should show EE EE EE
		assert.Equal(t, uint8(0xEE), game.DIS1, "DIS1 should be 0xEE after E command")
		assert.Equal(t, uint8(0xEE), game.DIS2, "DIS2 should be 0xEE after E command")
		assert.Equal(t, uint8(0xEE), game.DIS3, "DIS3 should be 0xEE after E command")

		// Reversed flag should be set
		assert.True(t, game.Reversed, "Reversed flag should be true after E command")
	})

	t.Run("double E command restores position", func(t *testing.T) {
		var buf bytes.Buffer
		game := NewGame(&buf)
		game.SetupBoard()

		originalBoard := game.Board
		originalBK := game.BK

		// Execute E command twice
		game.HandleCharacter('E')
		game.HandleCharacter('E')

		// Should restore original position
		assert.Equal(t, originalBoard, game.Board, "Double E should restore Board")
		assert.Equal(t, originalBK, game.BK, "Double E should restore BK")
		assert.False(t, game.Reversed, "Reversed flag should be false after double E")
	})
}
