// ABOUTME: This file implements MOVE, UMOVE, and RUM routines for making and unmaking moves.
// ABOUTME: These replace the assembly's dual-stack mechanism (SP1/SP2) used during CHKCHK.

package microchess

import (
	"github.com/matteo/microchess-go/pkg/board"
)

// MOVE makes a move on the board and saves it to the move history stack.
// This implements the MOVE routine from assembly lines 511-539.
//
// The assembly uses a dual-stack mechanism (SP1 for call stack, SP2 for game state).
// In Go, we use MoveHistory slice as the equivalent of SP2.
//
// Algorithm:
//  1. Save move state to MoveHistory (equivalent to pushing to SP2)
//  2. Find piece at target square (SQUARE), mark as captured if found
//  3. Update Board[MovePiece] = MoveSquare
//
// Assembly reference: Lines 511-539
//
// MOVE            TSX
//
//	STX     SP1             ; Switch
//	LDX     SP2             ; Stacks
//	TXS
//	LDA     SQUARE
//	PHA                     ; TO SQUARE
//	TAY
//	LDX     #$1F
//
// CHECK           CMP     BOARD,X         ; CHECK FOR
//
//	BEQ     TAKE            ; CAPTURE
//	DEX
//	BPL     CHECK
//
// TAKE            LDA     #$CC
//
//	STA     BOARD,X
//	TXA                     ; CAPTURED
//	PHA                     ; PIECE
//	LDX     PIECE
//	LDA     BOARD,X
//	STY     BOARD,X         ; FROM
//	PHA                     ; SQUARE
//	TXA
//	PHA                     ; PIECE
//	LDA     MOVEN
//	PHA                     ; MOVEN
//
// STRV            TSX
//
//	STX     SP2             ; SWITCH
//	LDX     SP1             ; STACKS
//	TXS                     ; BACK
//	RTS
func (g *GameState) MOVE() {
	// Get from square (where piece currently is)
	fromSquare := g.Board[g.MovePiece]

	// Find piece at target square (if any)
	capturedPiece := NoPiece
	capturedSquare := board.Square(0xCC) // Default to off-board

	// Scan all 32 pieces (BOARD[0-15] + BK[0-15])
	// Assembly lines 519-522: Loop X from $1F down to $00
	for pieceIndex := 31; pieceIndex >= 0; pieceIndex-- {
		var pieceSquare board.Square

		if pieceIndex < 16 {
			// Board array pieces (indices 0-15)
			pieceSquare = g.Board[pieceIndex]
		} else {
			// BK array pieces (indices 16-31, stored in BK[0-15])
			pieceSquare = g.BK[pieceIndex-16]
		}

		// Check if this piece occupies the target square
		if pieceSquare == g.MoveSquare {
			// Found captured piece
			capturedPiece = Piece(pieceIndex)
			capturedSquare = pieceSquare

			// Mark as captured by setting position to 0xCC (off-board)
			// Assembly line 523-524: LDA #$CC / STA BOARD,X
			if pieceIndex < 16 {
				g.Board[pieceIndex] = 0xCC
			} else {
				g.BK[pieceIndex-16] = 0xCC
			}
			break
		}
	}

	// Save move state to history (equivalent to pushing to SP2)
	// Assembly pushes (in order): MOVEN, PIECE, from square, captured piece, to square
	// We store all in a MoveRecord struct
	record := MoveRecord{
		FromSquare:     fromSquare,
		ToSquare:       g.MoveSquare,
		MovingPiece:    g.MovePiece,
		CapturedPiece:  capturedPiece,
		CapturedSquare: capturedSquare,
		MoveN:          g.MoveN,
	}
	g.MoveHistory = append(g.MoveHistory, record)

	// Move the piece to target square
	// Assembly line 529: STY BOARD,X (stores SQUARE into PIECE's position)
	g.Board[g.MovePiece] = g.MoveSquare
}

// UMOVE unmakes the last move by popping from the move history stack.
// This implements the UMOVE routine from assembly lines 488-504.
//
// Algorithm:
//  1. Pop last MoveRecord from MoveHistory
//  2. Restore Board[MovingPiece] to FromSquare
//  3. Restore captured piece (if any) to its original square
//  4. Restore MOVEN
//
// Assembly reference: Lines 488-504
//
// UMOVE           TSX                     ; UNMAKE MOVE
//
//	STX     SP1
//	LDX     SP2             ; EXCHANGE
//	TXS                     ; STACKS
//	PLA                     ; MOVEN
//	STA     MOVEN
//	PLA                     ; CAPTURED
//	STA     PIECE           ; PIECE
//	TAX
//	PLA                     ; FROM SQUARE
//	STA     BOARD,X
//	PLA                     ; PIECE
//	TAX
//	PLA                     ; TO SOUARE
//	STA     SQUARE
//	STA     BOARD,X
//	JMP     STRV
func (g *GameState) UMOVE() {
	// Pop last move from history
	if len(g.MoveHistory) == 0 {
		// No moves to undo (shouldn't happen in normal flow)
		return
	}

	record := g.MoveHistory[len(g.MoveHistory)-1]
	g.MoveHistory = g.MoveHistory[:len(g.MoveHistory)-1]

	// Restore MOVEN
	// Assembly line 492-493: PLA / STA MOVEN
	g.MoveN = record.MoveN

	// Restore PIECE (moving piece index)
	// Assembly line 494-496: PLA / STA PIECE / TAX
	// This is CRITICAL - without this, MovePiece becomes corrupted!
	g.MovePiece = record.MovingPiece

	// Restore moving piece to its original square
	// Assembly line 497-498: PLA / STA BOARD,X
	g.Board[record.MovingPiece] = record.FromSquare

	// Restore captured piece (if any)
	if record.CapturedPiece != NoPiece {
		if record.CapturedPiece < 16 {
			// Board array piece
			g.Board[record.CapturedPiece] = record.CapturedSquare
		} else {
			// BK array piece
			g.BK[record.CapturedPiece-16] = record.CapturedSquare
		}
	}

	// Restore SQUARE (working square) to the destination
	// Assembly line 501-503: PLA / STA SQUARE / STA BOARD,X
	g.MoveSquare = record.ToSquare
}

// RUM reverses the board and unmakes the last move.
// This implements the RUM routine from assembly line 483.
//
// Assembly reference: Line 483
//
//	RUM             JSR     REVERSE         ; REVERSE BACK
//	                                        ; (falls through to UMOVE)
//
// In the assembly, RUM calls REVERSE and then falls through to UMOVE
// (no explicit JSR to UMOVE). In Go, we call both explicitly.
func (g *GameState) RUM() {
	g.Reverse()
	g.UMOVE()
}
