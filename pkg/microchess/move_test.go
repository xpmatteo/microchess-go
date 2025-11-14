// ABOUTME: This file contains unit tests for MOVE, UMOVE, and RUM routines.
// ABOUTME: These tests verify that moves can be made and unmade correctly.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

// TestMoveUmoveRoundtrip verifies that MOVE followed by UMOVE restores the board state.
func TestMoveUmoveRoundtrip(t *testing.T) {
	tests := []struct {
		name          string
		setupBoard    func(*GameState)
		movePiece     Piece
		targetSquare  board.Square
		expectCapture bool
	}{
		{
			name: "simple non-capture move",
			setupBoard: func(g *GameState) {
				g.SetupBoard()
			},
			movePiece:     PiecePawn7, // e2 pawn (piece 14)
			targetSquare:  0x24,       // e3
			expectCapture: false,
		},
		{
			name: "capture move",
			setupBoard: func(g *GameState) {
				g.SetupBoard()
				// Move white pawn to e5
				g.Board[PiecePawn7] = 0x54 // e5
				// Move black pawn to d6
				g.BK[PiecePawn5] = 0x53 // d6
			},
			movePiece:     PiecePawn7, // e5 pawn
			targetSquare:  0x53,       // d6 (capture black pawn)
			expectCapture: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(&bytes.Buffer{})
			tt.setupBoard(g)

			// Save original board state
			origBoard := g.Board
			origBK := g.BK

			// Set up move state
			g.MovePiece = tt.movePiece
			g.MoveSquare = tt.targetSquare
			g.MoveN = 4 // arbitrary MOVEN value

			// Make the move
			g.MOVE()

			// Verify move was made
			assert.Equal(t, tt.targetSquare, g.Board[tt.movePiece], "piece should be at target square")

			// If capture expected, verify captured piece is marked
			if tt.expectCapture {
				found := false
				for i := 0; i < 16; i++ {
					if g.BK[i] == 0xCC {
						found = true
						break
					}
				}
				assert.True(t, found, "captured piece should be marked as 0xCC")
			}

			// Verify move was added to history
			assert.Len(t, g.MoveHistory, 1, "move should be in history")

			// Unmake the move
			g.UMOVE()

			// Verify board state restored
			assert.Equal(t, origBoard, g.Board, "Board should be restored")
			assert.Equal(t, origBK, g.BK, "BK should be restored")
			assert.Len(t, g.MoveHistory, 0, "move history should be empty")
		})
	}
}

// TestRUM is tested implicitly through CHKCHK functionality in cmove_test.go.
// RUM (REVERSE + UMOVE) is specifically used in CHKCHK to restore state after trial moves.

// These tests are superseded by the more comprehensive CHKCHK tests in cmove_test.go
// (TestCMOVE_CHKCHKDetectsPinnedPiece and TestCMOVE_CHKCHKSkippedWhenNotNeeded)
