// ABOUTME: This file implements the GNM (Generate New Move) routine from MicroChess.
// ABOUTME: GNM generates all pseudo-legal moves for the current side by iterating through pieces.

package microchess

import (
	"github.com/matteo/microchess-go/pkg/board"
)

// Move represents a chess move with from/to squares and the piece being moved.
type Move struct {
	From  board.Square
	To    board.Square
	Piece Piece
}

// MoveCallback is called for each generated move.
// This allows GNM to either collect moves (for 'L' command) or call JANUS (for computer play).
type MoveCallback func(from, to board.Square, piece Piece)

// janusCheckDetection implements the JANUS check detection logic (assembly lines 223-234).
// When STATE == -7 (0xF9), it checks if the current move captures the king at BK[0].
// If so, it sets InChek = 0x00 to indicate the king is in check.
//
// Assembly reference:
//
//	NOCOUNT CPX     #$F9          ; Is STATE = -7?
//	        BNE     TREE
//	        LDA     BK            ; Get king position
//	        CMP     SQUARE        ; Does this move attack the king?
//	        BNE     RETJ
//	        LDA     #$00          ; YES - set INCHEK = 0
//	        STA     INCHEK
//	RETJ    RTS
//
// Returns true if check detection is active and we should skip calling the normal callback.
func (g *GameState) janusCheckDetection() bool {
	// Check if we're in check detection mode (STATE == -7 / 0xF9)
	if g.State != -7 {
		return false // Not in check detection mode
	}

	// Check if this move captures the king at BK[0]
	// Assembly: LDA BK / CMP SQUARE
	if g.MoveSquare == g.BK[0] {
		// King can be captured! Set InChek = 0x00
		// Assembly: LDA #$00 / STA INCHEK
		g.InChek = 0x00
	}

	// Return true to indicate we're in check detection mode
	// The caller should skip the normal callback (JANUS just detects, doesn't evaluate)
	return true
}

// Reset restores MoveSquare to the current piece's board position.
// This implements the RESET routine from assembly line 473.
//
// Assembly reference:
//
//	RESET    LDX PIECE      ; Get piece index
//	         LDA BOARD,X    ; Load piece's current position
//	         STA SQUARE     ; Store in working square
//	         RTS
func (g *GameState) Reset() {
	g.MoveSquare = g.Board[g.MovePiece]
}

// singleMove generates a single move in the current direction (MOVEN).
// This implements the SNGMV routine from assembly line 357.
//
// Assembly reference:
//
//	SNGMV    JSR CMOVE      ; Calculate move
//	         BMI ILL1       ; If illegal, skip
//	         JSR JANUS      ; Evaluate move
//	ILL1     JSR RESET      ; Restore position
//	         DEC MOVEN      ; Next direction
//	         RTS
//
// Parameters:
//   - callback: Function to call if move is legal
//
// Returns:
//   - true if MOVEN should continue decrementing (matching assembly flow)
func (g *GameState) singleMove(callback MoveCallback) bool {
	// Calculate move using CMOVE
	fromSquare := g.MoveSquare
	result := g.CMOVE(fromSquare, g.MoveN)

	// If legal (not illegal and not in check), process the move
	if !result.Illegal && !result.InCheck {
		// JANUS check detection: If STATE==-7, check if this move attacks the king
		// If so, set InChek=0x00 and skip callback (assembly lines 223-234)
		if !g.janusCheckDetection() {
			// Not in check detection mode, call normal callback (if provided)
			if callback != nil {
				callback(fromSquare, g.MoveSquare, g.MovePiece)
			}
		}
	}

	// Restore piece position
	g.Reset()

	// Decrement MOVEN
	g.MoveN--

	return true
}

// slidingLine generates all moves in one direction until blocked or captures.
// This implements the LINE routine from assembly line 367.
//
// Assembly reference:
//
//	LINE     JSR CMOVE      ; Calculate next square
//	         BCC OVL        ; No check? Continue
//	         BVC LINE       ; No capture? Keep sliding
//	OVL      BMI ILL        ; Illegal? Stop
//	         PHP            ; Save flags
//	         JSR JANUS      ; Evaluate move
//	         PLP            ; Restore flags
//	         BVC LINE       ; Not a capture? Continue sliding
//	ILL      JSR RESET      ; Restore position
//	         DEC MOVEN      ; Next direction
//	         RTS
//
// Parameters:
//   - callback: Function to call for each legal move in this line
//
// Returns:
//   - true if MOVEN should continue decrementing
func (g *GameState) slidingLine(callback MoveCallback) bool {
	fromSquare := g.Board[g.MovePiece]

	for {
		// Calculate next square in this direction
		result := g.CMOVE(g.MoveSquare, g.MoveN)

		// If check without capture, continue sliding (assembly: BCC OVL / BVC LINE)
		if result.InCheck && !result.Capture {
			continue
		}

		// If illegal (off board or own piece), stop this line
		if result.Illegal {
			break
		}

		// Legal move - process it
		// JANUS check detection: If STATE==-7, check if this move attacks the king
		if !g.janusCheckDetection() {
			// Not in check detection mode, call normal callback (if provided)
			if callback != nil {
				callback(fromSquare, g.MoveSquare, g.MovePiece)
			}
		}

		// If capture, stop sliding (assembly: BVC LINE - branch if V clear)
		if result.Capture {
			break
		}

		// Empty square - continue sliding in this direction
		// MoveSquare has been updated by CMOVE, so next iteration continues from there
	}

	// Restore piece to original position
	g.Reset()

	// Decrement MOVEN for next direction
	g.MoveN--

	return true
}

// generateKingMoves generates all 8 possible king moves (single steps in all directions).
// This implements the KING section from assembly line 306.
//
// Assembly reference:
//
//	KING     JSR SNGMV      ; Single move
//	         BNE KING       ; Loop while MOVEN != 0
//	         BEQ NEWP       ; Done, next piece
func (g *GameState) generateKingMoves(callback MoveCallback) {
	// King uses moves 8 down to 1 (all 8 directions)
	g.MoveN = 8

	for g.MoveN != 0 {
		g.singleMove(callback)
	}
}

// generateQueenMoves generates all queen moves (8 sliding directions).
// This implements the QUEEN section from assembly line 309.
//
// Assembly reference:
//
//	QUEEN    JSR LINE       ; Sliding line
//	         BNE QUEEN      ; Loop while MOVEN != 0
//	         BEQ NEWP       ; Done, next piece
func (g *GameState) generateQueenMoves(callback MoveCallback) {
	// Queen uses moves 8 down to 1 (all 8 rays)
	g.MoveN = 8

	for g.MoveN != 0 {
		g.slidingLine(callback)
	}
}

// generateRookMoves generates all rook moves (4 orthogonal directions).
// This implements the ROOK section from assembly line 313.
//
// Assembly reference:
//
//	ROOK     LDX #$04       ; 4 directions
//	         STX MOVEN
//	AGNR     JSR LINE       ; Sliding line
//	         BNE AGNR       ; Loop while MOVEN != 0
//	         BEQ NEWP       ; Done, next piece
func (g *GameState) generateRookMoves(callback MoveCallback) {
	// Rook uses moves 4 down to 1 (orthogonal only)
	g.MoveN = 4

	for g.MoveN != 0 {
		g.slidingLine(callback)
	}
}

// generateBishopMoves generates all bishop moves (4 diagonal directions).
// This implements the BISHOP section from assembly line 319.
//
// Assembly reference:
//
//	BISHOP   JSR LINE       ; Sliding line
//	         LDA MOVEN
//	         CMP #$04       ; Stop at move 4
//	         BNE BISHOP     ; Loop while MOVEN != 4
//	         BEQ NEWP       ; Done, next piece
func (g *GameState) generateBishopMoves(callback MoveCallback) {
	// Bishop starts at move 8, generates moves 8,7,6,5 (diagonals only)
	g.MoveN = 8

	for g.MoveN != 4 {
		g.slidingLine(callback)
	}
}

// generateKnightMoves generates all 8 knight moves (L-shaped jumps).
// This implements the KNIGHT section from assembly line 325.
//
// Assembly reference:
//
//	KNIGHT   LDX #$10       ; 16 (uses moves 16-9)
//	         STX MOVEN
//	AGNN     JSR SNGMV      ; Single move
//	         LDA MOVEN
//	         CMP #$08       ; Stop at move 8
//	         BNE AGNN       ; Loop while MOVEN != 8
//	         BEQ NEWP       ; Done, next piece
func (g *GameState) generateKnightMoves(callback MoveCallback) {
	// Knight uses moves 16 down to 9 (8 L-shaped moves)
	g.MoveN = 16

	for g.MoveN != 8 {
		g.singleMove(callback)
	}
}

// generatePawnMoves generates pawn moves (forward non-capture, diagonal captures).
// This implements the PAWN section from assembly line 333.
//
// Assembly reference:
//
//	PAWN     LDX #$06       ; Start with right capture
//	         STX MOVEN
//	P1       JSR CMOVE      ; Try right diagonal
//	         BVC P2         ; Not a capture? Skip
//	         BMI P2         ; Illegal? Skip
//	         JSR JANUS      ; Legal capture
//	P2       JSR RESET
//	         DEC MOVEN      ; Try left capture
//	         LDA MOVEN
//	         CMP #$05
//	         BEQ P1         ; Loop for left capture
//	P3       JSR CMOVE      ; Forward move
//	         BVS NEWP       ; Capture? Illegal for pawn forward
//	         BMI NEWP       ; Off board? Done
//	         JSR JANUS      ; Legal forward move
//	         LDA SQUARE
//	         AND #$F0       ; Check rank
//	         CMP #$20       ; On rank 2?
//	         BEQ P3         ; Yes, can do double move
//	         JMP NEWP       ; Done
func (g *GameState) generatePawnMoves(callback MoveCallback) {
	fromSquare := g.Board[g.MovePiece]

	// Try right diagonal capture (MOVEN=6)
	g.MoveN = 6
	result := g.CMOVE(g.MoveSquare, g.MoveN)
	if result.Capture && !result.Illegal && !result.InCheck {
		// JANUS check detection
		if !g.janusCheckDetection() {
			if callback != nil {
				callback(fromSquare, g.MoveSquare, g.MovePiece)
			}
		}
	}

	// Try left diagonal capture (MOVEN=5)
	g.Reset()
	g.MoveN = 5
	result = g.CMOVE(g.MoveSquare, g.MoveN)
	if result.Capture && !result.Illegal && !result.InCheck {
		// JANUS check detection
		if !g.janusCheckDetection() {
			if callback != nil {
				callback(fromSquare, g.MoveSquare, g.MovePiece)
			}
		}
	}

	// Try forward move(s) (MOVEN=4)
	g.Reset()
	g.MoveN = 4

	for {
		result = g.CMOVE(g.MoveSquare, g.MoveN)

		// If capture, illegal, or leaves king in check, pawn can't move forward
		if result.Capture || result.Illegal || result.InCheck {
			break
		}

		// Legal forward move
		// JANUS check detection
		if !g.janusCheckDetection() {
			if callback != nil {
				callback(fromSquare, g.MoveSquare, g.MovePiece)
			}
		}

		// Check if on rank 2 (can do double move)
		// Assembly: AND #$F0 / CMP #$20
		if (g.MoveSquare & 0xF0) != 0x20 {
			break // Not on rank 2, done
		}

		// On rank 2, can try another forward move (no Reset - keep moving forward)
	}
}

// GNM (Generate New Move) generates all pseudo-legal moves for the current side.
// This implements the main GNM routine from assembly line 286.
//
// INPUTS (GameState fields read):
//   - Board[0..15]: Current positions of all pieces (set by RESET for each piece)
//   - State: Controls behavior (-7 = check detection mode, other values = normal move generation)
//   - BK[0]: King position (used when State == -7 for check detection)
//
// OUTPUTS (GameState fields modified):
//   - MovePiece: Iterates from 15 down to 0 (all pieces for current side)
//   - MoveSquare: Working square updated by CMOVE as moves are generated
//   - MoveN: Move direction index (varies by piece type: 1-8 for sliding, 9-16 for knights, etc.)
//   - InChek: Set to 0x00 if State == -7 and a move attacks the king (check detection)
//
// CALLBACK INVOCATION:
//   - callback(from, to, piece) is called for each legal move generated
//   - NOT called when State == -7 (check detection mode only sets InChek)
//   - If callback is nil, moves are still generated (useful for side effects like check detection)
//
// ALGORITHM:
//  1. Iterates through pieces 15 down to 0 (white's pieces)
//  2. For each piece, dispatches to appropriate move generator based on piece type
//  3. Each move generator uses CMOVE to test moves and calls callback for legal ones
//  4. When State == -7, janusCheckDetection sets InChek if king can be captured
//
// Assembly reference:
//
//	GNM      LDA #$10       ; Start with piece 16
//	         STA PIECE
//	NEWP     DEC PIECE      ; Next piece
//	         BPL NEX        ; Continue if >= 0
//	         RTS            ; All done
//	NEX      JSR RESET      ; Load piece position
//	         LDY PIECE
//	         LDX #$08       ; Default MOVEN
//	         STX MOVEN
//	         CPY #$08       ; Pawn?
//	         BPL PAWN
//	         CPY #$06       ; Knight?
//	         BPL KNIGHT
//	         CPY #$04       ; Bishop?
//	         BPL BISHOP
//	         CPY #$01       ; Queen?
//	         BEQ QUEEN
//	         BPL ROOK       ; Rook
//	         ; Must be King (piece 0)
//	KING     ...
//
// Parameters:
//   - callback: Function called for each generated move
//     If nil, moves are generated but not reported (useful for counting)
func (g *GameState) GNM(callback MoveCallback) {
	// Start with piece index 16, will decrement to 15 first
	g.MovePiece = 16

	for {
		// Decrement to next piece (assembly: DEC PIECE)
		g.MovePiece--

		// Check if done (assembly: BPL NEX - branch if positive)
		// Piece is uint8, so we check for wraparound
		if g.MovePiece > 15 {
			return // All pieces processed
		}

		// Load piece's current position (assembly: JSR RESET)
		g.Reset()

		// Default MOVEN (assembly: LDX #$08 / STX MOVEN)
		g.MoveN = 8

		// Dispatch based on piece type
		// Assembly compares piece index to determine type
		switch {
		case g.MovePiece >= 8: // Pawns (8-15)
			g.generatePawnMoves(callback)

		case g.MovePiece >= 6: // Knights (6-7)
			g.generateKnightMoves(callback)

		case g.MovePiece >= 4: // Bishops (4-5)
			g.generateBishopMoves(callback)

		case g.MovePiece == 1: // Queen (1)
			g.generateQueenMoves(callback)

		case g.MovePiece >= 1: // Rooks (2-3, but already handled queen so just 2-3)
			g.generateRookMoves(callback)

		default: // King (0)
			g.generateKingMoves(callback)
		}
	}
}
