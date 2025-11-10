// ABOUTME: This file contains tests for MicroChess types and game state.
// ABOUTME: It verifies piece setup, board initialization, and piece finding logic.

package microchess

import (
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	game := NewGame()
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
		PieceKing:   0x03,
		PieceQueen:  0x04,
		PieceRook1:  0x00,
		PieceRook2:  0x07,
		PiecePawn1:  0x10,
		PiecePawn8:  0x13,
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
	game := NewGame()
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
