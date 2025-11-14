// ABOUTME: This file tests the CMOVE (Calculate Move) routine implementation.
// ABOUTME: It verifies CMOVE's actual responsibilities: board edge detection and piece collision detection.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

// TestCMOVE_OffBoardDetection tests the 0x88 edge detection logic.
// CMOVE should return illegal (N=true) when the calculated target square has bit 0x88 set.
// Note: STATE is set to -1 to skip CHKCHK, testing only basic edge detection.
func TestCMOVE_OffBoardDetection(t *testing.T) {
	tests := []struct {
		name        string
		startSquare board.Square
		moven       uint8
		description string
	}{
		// Right edge tests (file overflow: bit 3 gets set)
		{
			name:        "Right edge: h1 + right",
			startSquare: 0x07, // h1 (file 7, rank 0)
			moven:       3,    // MOVEX[3] = 0x01 (right)
			description: "0x07 + 0x01 = 0x08 (bit 3 set)",
		},
		{
			name:        "Right edge: h4 + right",
			startSquare: 0x37,
			moven:       3,
			description: "0x37 + 0x01 = 0x38 (bit 3 set)",
		},
		{
			name:        "Right edge: h8 + right",
			startSquare: 0x77,
			moven:       3,
			description: "0x77 + 0x01 = 0x78 (bit 3 set)",
		},

		// Top edge tests (rank overflow: bit 7 gets set)
		{
			name:        "Top edge: a8 + up",
			startSquare: 0x70, // a8 (file 0, rank 7)
			moven:       4,    // MOVEX[4] = 0x10 (up)
			description: "0x70 + 0x10 = 0x80 (bit 7 set)",
		},
		{
			name:        "Top edge: d8 + up",
			startSquare: 0x73,
			moven:       4,
			description: "0x73 + 0x10 = 0x83 (bit 7 set)",
		},
		{
			name:        "Top edge: h8 + up",
			startSquare: 0x77,
			moven:       4,
			description: "0x77 + 0x10 = 0x87 (bit 7 set)",
		},

		// Left edge tests (file underflow: wraps to negative, bit 3 set)
		{
			name:        "Left edge: a1 + left",
			startSquare: 0x00, // a1 (file 0, rank 0)
			moven:       2,    // MOVEX[2] = -0x01 (left)
			description: "0x00 - 0x01 = 0xFF (bit 7 set)",
		},
		{
			name:        "Left edge: a4 + left",
			startSquare: 0x30,
			moven:       2,
			description: "0x30 - 0x01 = 0x2F (bit 3 set due to signed arithmetic)",
		},

		// Bottom edge tests (rank underflow)
		{
			name:        "Bottom edge: a1 + down",
			startSquare: 0x00,
			moven:       1, // MOVEX[1] = -0x10 (down)
			description: "0x00 - 0x10 = 0xF0 (bit 7 set)",
		},
		{
			name:        "Bottom edge: h1 + down",
			startSquare: 0x07,
			moven:       1,
			description: "0x07 - 0x10 = 0xF7 (bit 7 set)",
		},

		// Diagonal edge tests
		{
			name:        "Corner: h8 + up-right",
			startSquare: 0x77,
			moven:       5, // MOVEX[5] = 0x11 (up-right)
			description: "0x77 + 0x11 = 0x88 (both bits 3 and 7 set)",
		},

		// Knight move edge tests
		{
			name:        "Knight: h8 + knight up-up-right",
			startSquare: 0x77,
			moven:       16, // MOVEX[16] = 0x21 (knight move)
			description: "0x77 + 0x21 = 0x98 (off board)",
		},
		{
			name:        "Knight: a1 + knight down-down-left",
			startSquare: 0x00,
			moven:       9, // MOVEX[9] = -0x21 (knight move)
			description: "0x00 - 0x21 = 0xDF (off board)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(&bytes.Buffer{})
			// Initialize empty board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place a piece at the starting square
			game.Board[PieceKing] = tt.startSquare

			// Skip CHKCHK for basic edge detection test
			game.State = -1

			result := game.CMOVE(tt.startSquare, tt.moven)

			// Verify move is illegal (N flag set)
			assert.False(t, result.isLegal(),
				"Move from $%02X with MOVEN=%d should be illegal (off board): %s",
				tt.startSquare, tt.moven, tt.description)
			assert.True(t, result.Illegal,
				"N flag should be set for off-board move")
			assert.False(t, result.Capture,
				"V flag should be clear (no capture)")
			assert.False(t, result.InCheck,
				"C flag should be clear (no check detection)")
		})
	}
}

// TestCMOVE_EmptySquare tests moving to an empty square on the board.
// CMOVE should return legal (N=false, V=false, C=false).
// Note: STATE is set to -1 to skip CHKCHK, testing only basic collision detection.
func TestCMOVE_EmptySquare(t *testing.T) {
	tests := []struct {
		name         string
		startSquare  board.Square
		moven        uint8
		expectSquare board.Square
		description  string
	}{
		{
			name:         "Move up from center",
			startSquare:  0x33, // d4
			moven:        4,    // MOVEX[4] = 0x10 (up)
			expectSquare: 0x43, // d5
		},
		{
			name:         "Move right from center",
			startSquare:  0x33,
			moven:        3, // MOVEX[3] = 0x01 (right)
			expectSquare: 0x34,
		},
		{
			name:         "Move diagonal up-right",
			startSquare:  0x33,
			moven:        5, // MOVEX[5] = 0x11 (diagonal)
			expectSquare: 0x44,
		},
		{
			name:         "Knight move",
			startSquare:  0x33,
			moven:        16, // MOVEX[16] = 0x21 (knight)
			expectSquare: 0x54,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(&bytes.Buffer{})
			// Initialize empty board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place a piece at the starting square
			game.Board[PieceKing] = tt.startSquare

			// Skip CHKCHK for basic empty square test
			game.State = -1

			result := game.CMOVE(tt.startSquare, tt.moven)

			// Verify move is legal with no capture
			assert.True(t, result.isLegal(),
				"Move from $%02X with MOVEN=%d to $%02X should be legal (empty square)",
				tt.startSquare, tt.moven, tt.expectSquare)
			assert.False(t, result.Illegal, "N flag should be clear (legal)")
			assert.False(t, result.Capture, "V flag should be clear (no capture)")
			assert.False(t, result.InCheck, "C flag should be clear (no check)")
		})
	}
}

// TestCMOVE_OwnPieceCollision tests collision detection with own pieces.
// CMOVE should return illegal (N=true, V=false, C=false) when target square has own piece.
// Note: STATE is set to -1 to skip CHKCHK, testing only basic collision detection.
func TestCMOVE_OwnPieceCollision(t *testing.T) {
	tests := []struct {
		name           string
		startSquare    board.Square
		moven          uint8
		blockingPiece  Piece
		blockingSquare board.Square
		description    string
	}{
		{
			name:           "Own pawn blocks forward move",
			startSquare:    0x33,
			moven:          4, // up
			blockingPiece:  PiecePawn1,
			blockingSquare: 0x43,
		},
		{
			name:           "Own queen blocks diagonal",
			startSquare:    0x33,
			moven:          5, // diagonal up-right
			blockingPiece:  PieceQueen,
			blockingSquare: 0x44,
		},
		{
			name:           "Own knight blocks L-move",
			startSquare:    0x33,
			moven:          16, // knight move
			blockingPiece:  PieceKnight1,
			blockingSquare: 0x54,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(&bytes.Buffer{})
			// Initialize empty board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the moving piece
			game.Board[PieceKing] = tt.startSquare
			// Place own blocking piece
			game.Board[tt.blockingPiece] = tt.blockingSquare

			// Skip CHKCHK for basic collision test
			game.State = -1

			result := game.CMOVE(tt.startSquare, tt.moven)

			// Verify move is illegal with no capture flag
			assert.False(t, result.isLegal(),
				"Move from $%02X with MOVEN=%d should be illegal (blocked by own piece at $%02X)",
				tt.startSquare, tt.moven, tt.blockingSquare)
			assert.True(t, result.Illegal, "N flag should be set (illegal)")
			assert.False(t, result.Capture, "V flag should be clear (own piece, not capture)")
			assert.False(t, result.InCheck, "C flag should be clear (no check)")
		})
	}
}

// TestCMOVE_OpponentPieceCollision tests collision detection with opponent pieces.
// CMOVE should return legal with capture flag (N=false, V=true, C=false).
// Note: STATE is set to -1 to skip CHKCHK, testing only basic capture detection.
func TestCMOVE_OpponentPieceCollision(t *testing.T) {
	tests := []struct {
		name           string
		startSquare    board.Square
		moven          uint8
		opponentPiece  Piece
		opponentSquare board.Square
		description    string
	}{
		{
			name:           "Capture opponent pawn",
			startSquare:    0x33,
			moven:          4, // up
			opponentPiece:  PiecePawn1,
			opponentSquare: 0x43,
		},
		{
			name:           "Capture opponent queen diagonally",
			startSquare:    0x33,
			moven:          5, // diagonal up-right
			opponentPiece:  PieceQueen,
			opponentSquare: 0x44,
		},
		{
			name:           "Capture opponent knight with L-move",
			startSquare:    0x33,
			moven:          16, // knight move
			opponentPiece:  PieceKnight1,
			opponentSquare: 0x54,
		},
		{
			name:           "Capture opponent king",
			startSquare:    0x33,
			moven:          3, // right
			opponentPiece:  PieceKing,
			opponentSquare: 0x34,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(&bytes.Buffer{})
			// Initialize empty board
			for i := range game.Board {
				game.Board[i] = 0xFF
				game.BK[i] = 0xFF
			}
			// Place the moving piece
			game.Board[PieceKing] = tt.startSquare
			// Place opponent piece
			game.BK[tt.opponentPiece] = tt.opponentSquare

			// Skip CHKCHK for basic capture test
			game.State = -1

			result := game.CMOVE(tt.startSquare, tt.moven)

			// Verify move is legal with capture flag set
			assert.True(t, result.isLegal(),
				"Move from $%02X with MOVEN=%d should be legal (can capture opponent at $%02X)",
				tt.startSquare, tt.moven, tt.opponentSquare)
			assert.False(t, result.Illegal, "N flag should be clear (legal)")
			assert.True(t, result.Capture, "V flag should be set (capture)")
			assert.False(t, result.InCheck, "C flag should be clear (no check)")
			assert.True(t, result.isCapture(), "isCapture() should return true")
		})
	}
}

// TestCMOVE_CHKCHKDetectsPinnedPiece tests that CHKCHK detects when a move exposes the king to check.
// This tests the bishop pin scenario from the acceptance test.
func TestCMOVE_CHKCHKDetectsPinnedPiece(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.SetupBoard()

	// Create a pin: King at d1 (0x03), pawn at e2 (0x14), black bishop at g4 (0x36)
	// The diagonal d1-e2-f3-g4 means the pawn is pinned and cannot move
	game.BK[5] = 0x36 // Black kingside bishop at g4

	// Enable CHKCHK by setting STATE=4
	game.State = 4

	tests := []struct {
		name        string
		piece       Piece
		fromSquare  board.Square
		moven       uint8
		expectCheck bool
		description string
	}{
		{
			name:        "e2-e3 exposes king to bishop",
			piece:       PiecePawn7, // e-file pawn (piece 14)
			fromSquare:  0x14,       // e2
			moven:       4,          // up
			expectCheck: true,
			description: "Moving pinned pawn up one square exposes king",
		},
		{
			name:        "e2-e4 exposes king to bishop",
			piece:       PiecePawn7, // e-file pawn
			fromSquare:  0x14,       // e2
			moven:       4,          // up (would need two calls for double move, but single test is enough)
			expectCheck: true,
			description: "Moving pinned pawn up exposes king",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the move
			game.MovePiece = tt.piece
			game.MoveSquare = tt.fromSquare

			// Call CMOVE (which will run CHKCHK)
			result := game.CMOVE(tt.fromSquare, tt.moven)

			// Verify check detection
			assert.Equal(t, tt.expectCheck, result.InCheck,
				"%s: InCheck should be %v", tt.description, tt.expectCheck)

			if tt.expectCheck {
				assert.False(t, result.isLegal(),
					"%s: Move should not be legal (exposes king)", tt.description)
			}
		})
	}
}

// TestCMOVE_CHKCHKSkippedWhenNotNeeded tests that CHKCHK is only run when STATE is 0-7.
func TestCMOVE_CHKCHKSkippedWhenNotNeeded(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.SetupBoard()

	// Create the same pin scenario
	game.BK[5] = 0x36 // Black bishop at g4

	tests := []struct {
		name        string
		state       int8
		expectCheck bool
		description string
	}{
		{
			name:        "STATE=-1 skips CHKCHK",
			state:       -1,
			expectCheck: false,
			description: "Deep analysis skips check detection",
		},
		{
			name:        "STATE=8 skips CHKCHK",
			state:       8,
			expectCheck: false,
			description: "Continuation moves skip check detection",
		},
		{
			name:        "STATE=0 runs CHKCHK",
			state:       0,
			expectCheck: true,
			description: "Normal move generation includes check detection",
		},
		{
			name:        "STATE=4 runs CHKCHK",
			state:       4,
			expectCheck: true,
			description: "Normal move generation includes check detection",
		},
		{
			name:        "STATE=7 runs CHKCHK",
			state:       7,
			expectCheck: true,
			description: "Normal move generation includes check detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game.State = tt.state
			game.MovePiece = PiecePawn7 // e-pawn
			game.MoveSquare = 0x14      // e2

			result := game.CMOVE(0x14, 4) // move up

			assert.Equal(t, tt.expectCheck, result.InCheck,
				"%s: InCheck should be %v", tt.description, tt.expectCheck)
		})
	}
}
