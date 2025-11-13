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

// CMOVE calculates whether a move is legal, results in capture, or leaves own king in check
// This implements the core CMOVE routine from assembly lines 407-469, excluding CHKCHK.
//
// Assembly reference: Lines 407-469 in Microchess6502.txt
//
// Parameters:
//   - from: Starting square in 0x88 format
//   - moven: Index into MOVEX table (0-16) indicating move direction
//
// Returns CMoveResult with flags:
//   - N (Negative): true if move is illegal (off board or blocked by own piece)
//   - V (oVerflow): true if capture possible (opponent piece on target)
//   - C (Carry): false (always - CHKCHK not implemented yet)
//
// Algorithm (from CMOVE_PSEUDOCODE.md lines 76-110):
//  1. Calculate newSquare = square + MOVEX[moven]
//  2. Check if off board using 0x88 trick: (newSquare & 0x88) != 0
//  3. Scan all 32 pieces (Board[0-15] + BK[0-15]) for collision
//  4. If collision with own piece (index < 16): return ILLEGAL
//  5. If collision with opponent (index >= 16): set V flag (capture)
//  6. Return result (CHKCHK skipped for now)
func (g *GameState) CMOVE(from board.Square, moven uint8) CMoveResult {
	// Step 1: Calculate new from position
	// Assembly line 407-408: LDA SQUARE / CLC / ADC MOVEX,X / STA SQUARE
	newSquare := int16(from) + int16(MOVEX[moven])

	// Step 2: Check if off board using 0x88 trick
	// Assembly line 410-411: AND #$88 / BNE ILLEGAL
	// Any off-board from has bit $08 or $80 set
	if (newSquare & 0x88) != 0 {
		// ILLEGAL: off board
		// Assembly line 461-464
		return CMoveResult{
			N: true,  // Negative flag set (illegal)
			V: false, // No capture
			C: false, // No check
		}
	}

	// Step 3: Scan all 32 pieces to check for collision
	// Assembly line 413-421: Loop X from $1F down to $00
	// Memory layout: BOARD[0-15] at $50-$5F, BK[0-15] at $60-$6F
	// The assembly uses "LDA BOARD,X" which accesses:
	//   - BOARD[0-15] when X=0-15
	//   - BK[0-15] when X=16-31 (due to continuous memory)

	captureFlag := false

	// Scan from index 31 down to 0 (matching assembly's DEX loop)
	for pieceIndex := 31; pieceIndex >= 0; pieceIndex-- {
		var pieceSquare board.Square

		if pieceIndex < 16 {
			// Own pieces (BOARD array)
			pieceSquare = g.Board[pieceIndex]
		} else {
			// Opponent pieces (BK array)
			pieceSquare = g.BK[pieceIndex-16]
		}

		// Check if this piece occupies the target from
		// Assembly line 414: CMP BOARD,X
		if pieceSquare == board.Square(newSquare) {
			// Square is occupied!

			// Assembly line 415-416: CPX #$10 / BCC ILLEGAL
			if pieceIndex < 16 {
				// Blocked by own piece
				return CMoveResult{
					N: true,  // Negative flag set (illegal)
					V: false, // No capture
					C: false, // No check
				}
			}

			// Opponent piece - this is a capture
			// Assembly line 418-421: Set V flag using signed overflow trick
			// LDA #$7F / ADC #$01 / BVC SPX
			// The assembly does: $7F + 1 = $80, which causes signed overflow
			// In Go, we just set the flag directly
			captureFlag = true
			break // Found the collision, no need to continue scanning
		}
	}

	// Step 4: Return result
	// Assembly line 424 (SPX): Check if CHKCHK needed
	// We skip CHKCHK for now (lines 426-458), so go directly to RETL

	// Assembly line 459 (RETL): Legal move
	return CMoveResult{
		N: false,       // Not illegal
		V: captureFlag, // Capture if opponent piece found
		C: false,       // No check (CHKCHK not implemented)
	}
}
