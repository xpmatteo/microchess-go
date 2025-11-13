// ABOUTME: This file tests the CMOVE (Calculate Move) routine implementation.
// ABOUTME: It verifies move validation logic matches the 6502 assembly behavior.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestCMOVE_ValidMoves(t *testing.T) {
	tests := []struct {
		name        string
		piece       Piece
		startSquare board.Square
		validMoves  []board.Square
		description string
	}{
		// King tests - central position
		{
			name:        "King from $33 (central)",
			piece:       PieceKing,
			startSquare: 0x33,
			validMoves: []board.Square{
				0x22, 0x23, 0x24, // rank below
				0x32, 0x34, // same rank (left/right)
				0x42, 0x43, 0x44, // rank above
			},
			description: "King can move one square in any direction",
		},
		{
			name:        "King from $00 (corner)",
			piece:       PieceKing,
			startSquare: 0x00,
			validMoves: []board.Square{
				0x01,       // right
				0x10, 0x11, // rank above
			},
			description: "King at corner has only 3 valid moves",
		},
		{
			name:        "King from $77 (opposite corner)",
			piece:       PieceKing,
			startSquare: 0x77,
			validMoves: []board.Square{
				0x76,       // left
				0x66, 0x67, // rank below
			},
			description: "King at opposite corner has only 3 valid moves",
		},

		// Queen tests - central position
		{
			name:        "Queen from $33 (central)",
			piece:       PieceQueen,
			startSquare: 0x33,
			validMoves: []board.Square{
				// Horizontal (rank 3)
				0x30, 0x31, 0x32, 0x34, 0x35, 0x36, 0x37,
				// Vertical (file 3)
				0x03, 0x13, 0x23, 0x43, 0x53, 0x63, 0x73,
				// Diagonal (NE-SW)
				0x00, 0x11, 0x22, 0x44, 0x55, 0x66, 0x77,
				// Diagonal (NW-SE)
				0x06, 0x15, 0x24, 0x42, 0x51, 0x60,
			},
			description: "Queen can move horizontally, vertically, or diagonally",
		},
		{
			name:        "Queen from $03 (edge)",
			piece:       PieceQueen,
			startSquare: 0x03,
			validMoves: []board.Square{
				// Horizontal (rank 0)
				0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07,
				// Vertical (file 3)
				0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73,
				// Diagonal (NE-SW) - limited at edge
				0x14, 0x25, 0x36, 0x47,
				// Diagonal (NW-SE) - limited at edge
				0x12, 0x21, 0x30,
			},
			description: "Queen at edge has limited diagonal moves",
		},

		// Rook tests - central position
		{
			name:        "Rook from $33 (central)",
			piece:       PieceRook1,
			startSquare: 0x33,
			validMoves: []board.Square{
				// Horizontal (rank 3)
				0x30, 0x31, 0x32, 0x34, 0x35, 0x36, 0x37,
				// Vertical (file 3)
				0x03, 0x13, 0x23, 0x43, 0x53, 0x63, 0x73,
			},
			description: "Rook can move horizontally or vertically",
		},
		{
			name:        "Rook from $00 (corner)",
			piece:       PieceRook1,
			startSquare: 0x00,
			validMoves: []board.Square{
				// Horizontal (rank 0)
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				// Vertical (file 0)
				0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70,
			},
			description: "Rook at corner can move along two edges",
		},

		// Bishop tests - central position
		{
			name:        "Bishop from $33 (central)",
			piece:       PieceBishop1,
			startSquare: 0x33,
			validMoves: []board.Square{
				// Diagonal (NE-SW)
				0x00, 0x11, 0x22, 0x44, 0x55, 0x66, 0x77,
				// Diagonal (NW-SE)
				0x06, 0x15, 0x24, 0x42, 0x51, 0x60,
			},
			description: "Bishop can move diagonally",
		},
		{
			name:        "Bishop from $07 (corner)",
			piece:       PieceBishop1,
			startSquare: 0x07,
			validMoves: []board.Square{
				// Diagonal (SW)
				0x16, 0x25, 0x34, 0x43, 0x52, 0x61, 0x70,
			},
			description: "Bishop at corner has only one diagonal",
		},

		// Knight tests - central position
		{
			name:        "Knight from $33 (central)",
			piece:       PieceKnight1,
			startSquare: 0x33,
			validMoves: []board.Square{
				0x11, 0x15, // two ranks down
				0x21, 0x25, // one rank down
				0x41, 0x45, // one rank up
				0x51, 0x55, // two ranks up
			},
			description: "Knight moves in L-shape",
		},
		{
			name:        "Knight from $00 (corner)",
			piece:       PieceKnight1,
			startSquare: 0x00,
			validMoves: []board.Square{
				0x12, 0x21,
			},
			description: "Knight at corner has only 2 valid moves",
		},
		{
			name:        "Knight from $03 (edge)",
			piece:       PieceKnight1,
			startSquare: 0x03,
			validMoves: []board.Square{
				0x11, 0x15,
				0x21, 0x25,
			},
			description: "Knight at edge has limited moves",
		},

		// Pawn tests - central position
		{
			name:        "Pawn from $33 (central, empty ahead)",
			piece:       PiecePawn1,
			startSquare: 0x33,
			validMoves: []board.Square{
				0x43, // one square forward
			},
			description: "Pawn can move one square forward",
		},
		{
			name:        "Pawn from $13 (second rank)",
			piece:       PiecePawn1,
			startSquare: 0x13,
			validMoves: []board.Square{
				0x23, // one square forward
				0x33, // two squares forward (from starting rank)
			},
			description: "Pawn on starting rank can move one or two squares",
		},
		{
			name:        "Pawn from $10 (starting rank, left edge)",
			piece:       PiecePawn1,
			startSquare: 0x10,
			validMoves: []board.Square{
				0x20, // one square forward
				0x30, // two squares forward
			},
			description: "Pawn at edge of starting rank can still double-step",
		},
		{
			name:        "Pawn from $17 (starting rank, right edge)",
			piece:       PiecePawn8,
			startSquare: 0x17,
			validMoves: []board.Square{
				0x27, // one square forward
				0x37, // two squares forward
			},
			description: "Pawn at right edge of starting rank can double-step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game state with only the specified piece at startSquare
			game := NewGame(&bytes.Buffer{})
			// Initialize all squares to off-board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the piece at its starting square
			game.Board[tt.piece] = tt.startSquare

			// Create a map of valid moves for quick lookup
			validMap := make(map[board.Square]bool)
			for _, sq := range tt.validMoves {
				validMap[sq] = true
			}

			// Test all squares on the board (0x00 to 0x77, skipping invalid 0x88 squares)
			for rank := uint8(0); rank < 8; rank++ {
				for file := uint8(0); file < 8; file++ {
					targetSquare := board.Square(rank<<4 | file)

					// Skip the starting square
					if targetSquare == tt.startSquare {
						continue
					}

					// Call CMOVE to check if move from startSquare to targetSquare is legal
					result := game.CMOVE(tt.piece, tt.startSquare, targetSquare)

					if validMap[targetSquare] {
						// This square should be a valid move
						assert.True(t, result.isLegal(),
							"Move from $%02X to $%02X should be legal (piece=%d)",
							tt.startSquare, targetSquare, tt.piece)
						assert.False(t, result.isCapture(),
							"Move from $%02X to $%02X should not be a capture (no opponent pieces)",
							tt.startSquare, targetSquare)
					} else {
						// This square should be an invalid move
						assert.False(t, result.isLegal(),
							"Move from $%02X to $%02X should be illegal (piece=%d)",
							tt.startSquare, targetSquare, tt.piece)
					}
				}
			}
		})
	}
}

func TestCMOVE_CaptureFlag(t *testing.T) {
	tests := []struct {
		name         string
		piece        Piece
		startSquare  board.Square
		targetSquare board.Square
		hasOpponent  bool
		expectV      bool
		description  string
	}{
		{
			name:         "King captures opponent piece",
			piece:        PieceKing,
			startSquare:  0x33,
			targetSquare: 0x43,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "King moves to empty square",
			piece:        PieceKing,
			startSquare:  0x33,
			targetSquare: 0x43,
			hasOpponent:  false,
			expectV:      false,
		},
		{
			name:         "Queen captures opponent on diagonal",
			piece:        PieceQueen,
			startSquare:  0x33,
			targetSquare: 0x44,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "Rook captures opponent horizontally",
			piece:        PieceRook1,
			startSquare:  0x33,
			targetSquare: 0x35,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "Bishop captures opponent on diagonal",
			piece:        PieceBishop1,
			startSquare:  0x33,
			targetSquare: 0x44,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "Knight captures opponent",
			piece:        PieceKnight1,
			startSquare:  0x33,
			targetSquare: 0x45,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "Pawn captures opponent diagonally",
			piece:        PiecePawn1,
			startSquare:  0x33,
			targetSquare: 0x42,
			hasOpponent:  true,
			expectV:      true,
		},
		{
			name:         "Pawn cannot move diagonally to empty square",
			piece:        PiecePawn1,
			startSquare:  0x33,
			targetSquare: 0x42,
			hasOpponent:  false,
			expectV:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game state with piece at startSquare
			game := NewGame(&bytes.Buffer{})
			// Initialize all squares to off-board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the current player's piece at startSquare
			game.Board[tt.piece] = tt.startSquare

			// If hasOpponent, place an opponent piece at targetSquare
			if tt.hasOpponent {
				game.BK[PieceKing] = tt.targetSquare // Use any opponent piece
			}

			// Call CMOVE to check the move
			result := game.CMOVE(tt.piece, tt.startSquare, tt.targetSquare)

			// Verify V flag matches expectation
			assert.Equal(t, tt.expectV, result.isCapture(),
				"V flag should be %v for move from $%02X to $%02X (hasOpponent=%v)",
				tt.expectV, tt.startSquare, tt.targetSquare, tt.hasOpponent)

			// If expectV is true, move should be legal
			if tt.expectV {
				assert.True(t, result.isLegal(),
					"Capture move from $%02X to $%02X should be legal",
					tt.startSquare, tt.targetSquare)
			}
		})
	}
}

func TestCMOVE_OwnPieceBlocking(t *testing.T) {
	tests := []struct {
		name         string
		piece        Piece
		startSquare  board.Square
		targetSquare board.Square
		blockingWith Piece
		description  string
	}{
		{
			name:         "King blocked by own pawn",
			piece:        PieceKing,
			startSquare:  0x33,
			targetSquare: 0x43,
			blockingWith: PiecePawn1,
		},
		{
			name:         "Queen blocked by own rook horizontally",
			piece:        PieceQueen,
			startSquare:  0x33,
			targetSquare: 0x35,
			blockingWith: PieceRook1,
		},
		{
			name:         "Rook blocked by own bishop",
			piece:        PieceRook1,
			startSquare:  0x33,
			targetSquare: 0x53,
			blockingWith: PieceBishop1,
		},
		{
			name:         "Bishop blocked by own knight",
			piece:        PieceBishop1,
			startSquare:  0x33,
			targetSquare: 0x44,
			blockingWith: PieceKnight1,
		},
		{
			name:         "Knight blocked by own pawn",
			piece:        PieceKnight1,
			startSquare:  0x33,
			targetSquare: 0x45,
			blockingWith: PiecePawn1,
		},
		{
			name:         "Pawn blocked by own piece forward",
			piece:        PiecePawn1,
			startSquare:  0x33,
			targetSquare: 0x43,
			blockingWith: PiecePawn2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game state with piece at startSquare
			game := NewGame(&bytes.Buffer{})
			// Initialize all squares to off-board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the moving piece at startSquare
			game.Board[tt.piece] = tt.startSquare
			// Place own blocking piece at targetSquare
			game.Board[tt.blockingWith] = tt.targetSquare

			// Call CMOVE to check the move
			result := game.CMOVE(tt.piece, tt.startSquare, tt.targetSquare)

			// Verify move is illegal (N flag set)
			assert.False(t, result.isLegal(),
				"Move from $%02X to $%02X should be illegal (blocked by own piece %d)",
				tt.startSquare, tt.targetSquare, tt.blockingWith)

			// Verify capture flag is NOT set (not a capture)
			assert.False(t, result.isCapture(),
				"Move from $%02X to $%02X should not have capture flag set (own piece, not opponent)",
				tt.startSquare, tt.targetSquare)
		})
	}
}

func TestCMOVE_OpponentPieceBlocking(t *testing.T) {
	tests := []struct {
		name          string
		piece         Piece
		startSquare   board.Square
		blockerSquare board.Square
		beyondSquare  board.Square
		description   string
	}{
		{
			name:          "Queen blocked by opponent horizontally",
			piece:         PieceQueen,
			startSquare:   0x33,
			blockerSquare: 0x35,
			beyondSquare:  0x36,
			description:   "Queen cannot move beyond opponent piece horizontally",
		},
		{
			name:          "Queen blocked by opponent vertically",
			piece:         PieceQueen,
			startSquare:   0x33,
			blockerSquare: 0x53,
			beyondSquare:  0x63,
			description:   "Queen cannot move beyond opponent piece vertically",
		},
		{
			name:          "Queen blocked by opponent diagonally",
			piece:         PieceQueen,
			startSquare:   0x33,
			blockerSquare: 0x44,
			beyondSquare:  0x55,
			description:   "Queen cannot move beyond opponent piece diagonally",
		},
		{
			name:          "Rook blocked by opponent horizontally",
			piece:         PieceRook1,
			startSquare:   0x33,
			blockerSquare: 0x35,
			beyondSquare:  0x37,
			description:   "Rook cannot jump over opponent piece",
		},
		{
			name:          "Rook blocked by opponent vertically",
			piece:         PieceRook1,
			startSquare:   0x33,
			blockerSquare: 0x43,
			beyondSquare:  0x53,
			description:   "Rook cannot move through opponent piece",
		},
		{
			name:          "Bishop blocked by opponent on diagonal",
			piece:         PieceBishop1,
			startSquare:   0x33,
			blockerSquare: 0x44,
			beyondSquare:  0x55,
			description:   "Bishop cannot jump over opponent piece",
		},
		{
			name:          "Bishop blocked by opponent on other diagonal",
			piece:         PieceBishop1,
			startSquare:   0x33,
			blockerSquare: 0x42,
			beyondSquare:  0x51,
			description:   "Bishop stops at opponent piece, cannot continue",
		},
		{
			name:          "Pawn blocked by opponent forward (one square)",
			piece:         PiecePawn1,
			startSquare:   0x33,
			blockerSquare: 0x43,
			beyondSquare:  0x43,
			description:   "Pawn cannot move forward onto opponent piece",
		},
		{
			name:          "Pawn blocked by opponent forward (double-step blocked at first square)",
			piece:         PiecePawn1,
			startSquare:   0x13,
			blockerSquare: 0x23,
			beyondSquare:  0x33,
			description:   "Pawn cannot double-step if first square occupied",
		},
		{
			name:          "Pawn blocked by opponent forward (double-step blocked at second square)",
			piece:         PiecePawn1,
			startSquare:   0x13,
			blockerSquare: 0x33,
			beyondSquare:  0x33,
			description:   "Pawn cannot double-step if second square occupied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game state with piece at startSquare
			game := NewGame(&bytes.Buffer{})
			// Initialize all squares to off-board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the moving piece at startSquare
			game.Board[tt.piece] = tt.startSquare
			// Place opponent piece at blockerSquare
			game.BK[PieceKing] = tt.blockerSquare

			// Test 1: Move to blockerSquare should be LEGAL with V flag set (capture)
			resultCapture := game.CMOVE(tt.piece, tt.startSquare, tt.blockerSquare)
			assert.True(t, resultCapture.isLegal(),
				"Move from $%02X to $%02X should be legal (can capture opponent)",
				tt.startSquare, tt.blockerSquare)
			assert.True(t, resultCapture.isCapture(),
				"Move from $%02X to $%02X should have capture flag set (capturing opponent)",
				tt.startSquare, tt.blockerSquare)

			// Test 2: Move beyond blockerSquare should be ILLEGAL (cannot jump over)
			// Only test if beyondSquare is different from blockerSquare (pawn case)
			if tt.beyondSquare != tt.blockerSquare {
				resultBeyond := game.CMOVE(tt.piece, tt.startSquare, tt.beyondSquare)
				assert.False(t, resultBeyond.isLegal(),
					"Move from $%02X to $%02X should be illegal (cannot jump over opponent at $%02X)",
					tt.startSquare, tt.beyondSquare, tt.blockerSquare)
			}
		})
	}
}
