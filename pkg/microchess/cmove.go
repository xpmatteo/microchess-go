// ABOUTME: This file implements the CMOVE (Calculate Move) routine from MicroChess.
// ABOUTME: It validates moves and returns processor-style flags indicating legality and capture status.

package microchess

import (
	"github.com/matteo/microchess-go/pkg/board"
)

// CMoveResult represents the processor flags returned by CMOVE.
// These flags mirror the 6502 processor status register flags used in the assembly.
type CMoveResult struct {
	N bool // Negative flag: true if move is illegal (off board or blocked by own piece)
	V bool // oVerflow flag: true if capture possible (opponent piece on target)
	C bool // Carry flag: true if move leaves king in check
}

// isLegal returns true if the move is legal (N=false and C=false)
func (r CMoveResult) isLegal() bool {
	return !r.N && !r.C
}

// isCapture returns true if the move captures an opponent piece (V=true)
func (r CMoveResult) isCapture() bool {
	return r.V
}

// CMOVE calculates whether a move is legal and sets processor-style flags.
// This is a stub implementation that will be replaced with the actual CMOVE logic.
//
// Assembly reference: Lines 407-469 in Microchess6502.txt
//
// Parameters:
//   - piece: The piece being moved (0-15)
//   - from: Starting square in 0x88 format
//   - to: Target square in 0x88 format
//
// Returns CMoveResult with flags:
//   - N (Negative): true if move is illegal (off board or blocked by own piece)
//   - V (oVerflow): true if capture possible (opponent piece on target)
//   - C (Carry): true if move leaves king in check
func (g *GameState) CMOVE(piece Piece, from board.Square, to board.Square) CMoveResult {
	// TODO: Implement actual CMOVE logic
	// This stub always returns "legal, no capture, no check"
	return CMoveResult{
		N: false, // Legal move
		V: false, // No capture
		C: false, // No check
	}
}
