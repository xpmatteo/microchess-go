// ABOUTME: This file defines core MicroChess types including pieces and game state.
// ABOUTME: It maintains the same piece indexing and data structures as the original assembly.

package microchess

import (
	"fmt"
	"io"

	"github.com/matteo/microchess-go/pkg/board"
)

// Piece represents a chess piece by its index (0-15 for one side).
// The original assembly tracks pieces by index into the board array.
//
// Reference: Assembly lines 50-80 (piece organization)
type Piece uint8

const (
	PieceKing    Piece = 0  // King
	PieceQueen   Piece = 1  // Queen
	PieceRook1   Piece = 2  // Rook (queenside)
	PieceRook2   Piece = 3  // Rook (kingside)
	PieceBishop1 Piece = 4  // Bishop (queenside)
	PieceBishop2 Piece = 5  // Bishop (kingside)
	PieceKnight1 Piece = 6  // Knight (queenside)
	PieceKnight2 Piece = 7  // Knight (kingside)
	PiecePawn1   Piece = 8  // Pawn (a-file in initial setup)
	PiecePawn2   Piece = 9  // Pawn (h-file)
	PiecePawn3   Piece = 10 // Pawn (b-file)
	PiecePawn4   Piece = 11 // Pawn (g-file)
	PiecePawn5   Piece = 12 // Pawn (c-file)
	PiecePawn6   Piece = 13 // Pawn (f-file)
	PiecePawn7   Piece = 14 // Pawn (e-file)
	PiecePawn8   Piece = 15 // Pawn (d-file)

	NoPiece Piece = 0xFF // Indicates empty square or no piece
)

// GameState represents the complete state of a MicroChess game.
// This replaces the page zero variables ($50-$FF) from the assembly.
//
// Reference: doc/DATA_STRUCTURES.md for complete variable map
type GameState struct {
	// Board representation - piece positions
	// Board[i] = square where piece i is located (0x88 format)
	// Indices 0-15 represent white pieces (or current side)
	Board [16]board.Square

	// BK is the alternate board used during REVERSE
	// Assembly: $60-$6F
	BK [16]board.Square

	// REV tracks if board is reversed (0 = white's view, non-zero = black's view)
	// Assembly: $70 (REV flag)
	Reversed bool

	// LED display values (shown at bottom of board)
	// Assembly: $F9, $FA, $FB (DIS1, DIS2, DIS3)
	DIS1, DIS2, DIS3 uint8

	// Move entry state
	// PIECE tracks which piece is being moved (set after 4th digit entered)
	// Assembly: $FB (PIECE - overlaps with DIS3 in assembly, but used differently)
	SelectedPiece Piece

	// DigitCount tracks how many digits (0-7) have been entered (0-4)
	// Used to know when we have a complete move (4 digits)
	DigitCount uint8

	// I/O for display and input
	out io.Writer
}

// InitialSetup contains the starting positions for all pieces.
// This matches the SETW/SETB tables from the assembly (lines 672-687).
//
// Indices 0-15: White pieces (current side in assembly terminology)
// In actual play, this is copied to both Board arrays with coordinate transformation for black.
var InitialSetup = [32]board.Square{
	// White pieces (indices 0-7): K, Q, R, R, B, B, N, N
	0x03, 0x04, 0x00, 0x07, 0x02, 0x05, 0x01, 0x06,
	// White pawns (indices 8-15): specific file ordering from assembly
	0x10, 0x17, 0x11, 0x16, 0x12, 0x15, 0x14, 0x13,
	// Black pieces (indices 16-23): K, Q, R, R, B, B, N, N
	0x73, 0x74, 0x70, 0x77, 0x72, 0x75, 0x71, 0x76,
	// Black pawns (indices 24-31)
	0x60, 0x67, 0x61, 0x66, 0x62, 0x65, 0x64, 0x63,
}

// NewGame creates a new game with an empty board state.
// The board must be initialized with SetupBoard() (via 'C' command) before play.
// This matches the original assembly behavior where the board is uninitialized at startup.
//
// The out Writer is used for all display output (board, messages, etc.)
func NewGame(out io.Writer) *GameState {
	g := &GameState{
		out: out,
	}
	// Initialize boards with off-board sentinel values (0xFF)
	// This simulates the uninitialized state of the original
	for i := 0; i < 16; i++ {
		g.Board[i] = 0xFF // Off-board position
		g.BK[i] = 0xFF
	}
	// However, to match the original's display, put one black pawn at 00
	// This is the "garbage" state the original shows
	g.BK[8] = 0x00 // Black pawn at position 00

	// DIS1, DIS2, DIS3 start at 0x00 (uninitialized state)
	// DIS1 will be set to 0xFF after 4th digit when piece is found
	g.DIS1 = 0x00
	g.DIS2 = 0x00
	g.DIS3 = 0x00

	return g
}

// SetupBoard initializes the board to the starting position.
// This is equivalent to the SETUP routine (assembly line 665).
func (g *GameState) SetupBoard() {
	// Copy white pieces
	for i := 0; i < 16; i++ {
		g.Board[i] = InitialSetup[i]
	}
	// Copy black pieces (with coordinate transformation)
	for i := 0; i < 16; i++ {
		g.BK[i] = InitialSetup[i+16]
	}
	g.Reversed = false
}

// GetPieceChar returns a character representation of a piece at a square.
// Used for board display.
func GetPieceChar(piece Piece, isWhite bool) string {
	var baseChar string
	switch piece {
	case PieceKing:
		baseChar = "K"
	case PieceQueen:
		baseChar = "Q"
	case PieceRook1, PieceRook2:
		baseChar = "R"
	case PieceBishop1, PieceBishop2:
		baseChar = "B"
	case PieceKnight1, PieceKnight2:
		baseChar = "N"
	case PiecePawn1, PiecePawn2, PiecePawn3, PiecePawn4,
		PiecePawn5, PiecePawn6, PiecePawn7, PiecePawn8:
		baseChar = "P"
	default:
		return " *"
	}

	if isWhite {
		return "W" + baseChar
	}
	return "B" + baseChar
}

// FindPieceAt returns which piece (if any) is at the given square.
// Returns the piece index and true if found, or NoPiece and false if empty.
// The isWhite flag indicates the COLOR TO DISPLAY (which depends on Reversed flag).
//
// In the original assembly (POUT line 750), when REV flag is set, it uses different
// color characters (cpl+16 vs cpl) to flip white<->black display.
func (g *GameState) FindPieceAt(sq board.Square) (piece Piece, found bool, isWhite bool) {
	// Check pieces in BOARD array
	for i := Piece(0); i < 16; i++ {
		if g.Board[i] == sq {
			// BOARD pieces are displayed as WHITE when REV=0, BLACK when REV!=0
			// This matches assembly line 750: REV flag determines color character used
			return i, true, !g.Reversed // piece, found, isWhite=(REV==0)
		}
	}
	// Check pieces in BK array
	for i := Piece(0); i < 16; i++ {
		if g.BK[i] == sq {
			// BK pieces are displayed as BLACK when REV=0, WHITE when REV!=0
			return i, true, g.Reversed // piece, found, isWhite=(REV!=0)
		}
	}
	return NoPiece, false, false
}

// Reverse flips the board perspective by swapping Board and BK arrays
// and transforming all coordinates: new = 0x77 - old.
// This allows the engine to analyze the position from the opponent's perspective.
//
// Reference: REVERSE routine (assembly line 382)
func (g *GameState) Reverse() {
	// The assembly performs this transformation for X = 0x0F down to 0x00:
	//   temp = BK[X]
	//   BK[X] = 0x77 - Board[X]
	//   Board[X] = 0x77 - temp
	// This simultaneously swaps the arrays and transforms coordinates

	for i := 15; i >= 0; i-- {
		// Save BK[i]
		temp := g.BK[i]
		// Transform Board[i] and store in BK[i]
		g.BK[i] = 0x77 - g.Board[i]
		// Transform old BK[i] and store in Board[i]
		g.Board[i] = 0x77 - temp
	}

	// Toggle the reversed flag
	g.Reversed = !g.Reversed
}

// HandleCharacter processes a single character input and returns true if the program should continue.
// Returns false if the program should quit.
// This matches the original assembly's KIN routine which processes one character at a time.
//
// Reference: Assembly lines 110-152 (main input loop), 812-816 (KIN routine)
func (g *GameState) HandleCharacter(char byte) bool {
	// Mask to handle both upper and lowercase (original: AND #$4F masks bits)
	// Convert lowercase to uppercase for simplicity
	if char >= 'a' && char <= 'z' {
		char = char - 'a' + 'A'
	}

	switch char {
	case 'Q':
		// Quit program (assembly line 148)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline
		return false

	case 'C':
		// Setup board (SETUP routine, line 665, called at line 116)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed 'C'
		g.SetupBoard()
		// Set LED display to "CC CC CC" to indicate setup
		g.DIS1 = 0xCC
		g.DIS2 = 0xCC
		g.DIS3 = 0xCC
		g.Display()
		return true

	case 'E':
		// Reverse board perspective (REVERSE routine, line 382, called at line 126)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed 'E'
		g.Reverse()
		// Set LED display to "EE EE EE" to indicate reversal
		g.DIS1 = 0xEE
		g.DIS2 = 0xEE
		g.DIS3 = 0xEE
		g.Display()
		return true

	case 'P':
		// Print board (POUT routine, line 702, called at line 140)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed 'P'
		g.Display()
		return true

	case '\r', '\n':
		// Enter/Return key
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed char

		// If we have 4 digits entered, execute the move
		if g.DigitCount == 4 {
			g.ExecuteMove()
			g.Display()
		}
		return true

	case '0', '1', '2', '3', '4', '5', '6', '7':
		// Digit input for move entry (INPUT routine, assembly line 262)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed char

		digit := uint8(char - '0')

		// On first digit, reset DIS1 to 0xFF (no piece selected yet)
		// This matches observed 6502 behavior where DIS1 shows FF during digit entry
		if g.DigitCount == 0 {
			g.DIS1 = 0xFF
		}

		// Rotate digit into move display (DISMV routine, line 625)
		g.RotateDigitIntoMove(digit)
		g.DigitCount++

		// After 4th digit, find piece at from square (DIS2)
		// Assembly lines 266-272 (SEARCH loop in INPUT)
		if g.DigitCount == 4 {
			fromSquare := board.Square(g.DIS2)
			piece := g.FindPieceAtSquare(fromSquare)

			// Store piece index in DIS1 and SelectedPiece
			// Assembly line 271-272: STX DIS1 / STX PIECE
			g.DIS1 = uint8(piece)
			g.SelectedPiece = piece
		}

		// Display board to show updated LED display
		g.Display()
		return true

	default:
		// Unknown command - print error
		_, _ = fmt.Fprintf(g.out, "\r\nUnknown command: %c\r\n", char)
		_, _ = fmt.Fprintln(g.out, "Available commands: C (setup), E (reverse), P (print), Q (quit), 0-7 (move)")
		return true
	}
}

// ExecuteMove executes the move stored in SelectedPiece and DIS3 (target square).
// This is a SIMPLIFIED version of the MOVE routine (assembly line 511) for Phase 4.
//
// Phase 4 behavior (no validation, basic capture):
//   - Move Board[SelectedPiece] to DIS3
//   - If piece at target square, mark as captured (set to 0xCC)
//   - Reset DIS1 to 0xFF
//   - Keep DIS2/DIS3 showing last move
//
// Full MOVE implementation (with undo stack) comes in Phase 6.
//
// Assembly MOVE routine (simplified for Phase 4):
//   - Switch to alternate stack (SP2)
//   - Search for piece at SQUARE (target), mark as captured if found
//   - Update BOARD[PIECE] = SQUARE
//   - Switch back to hardware stack
func (g *GameState) ExecuteMove() {
	// If no piece was found at the "from" square, can't execute move
	if g.SelectedPiece == NoPiece {
		// In Phase 4, we just skip execution silently (no validation messages)
		// Reset digit count for next move
		g.DigitCount = 0
		g.DIS1 = 0xFF
		return
	}

	targetSquare := board.Square(g.DIS3)

	// Check if there's a piece at the target square (capture)
	capturedPiece := g.FindPieceAtSquare(targetSquare)
	if capturedPiece != NoPiece {
		// Mark piece as captured by setting position to 0xCC (off-board sentinel)
		// Assembly line 527: STA BOARD,X (stores $CC into captured piece's position)
		g.Board[capturedPiece] = 0xCC
	}

	// Move the selected piece to target square
	// Assembly line 535: STA BOARD,X (stores SQUARE into PIECE's position)
	g.Board[g.SelectedPiece] = targetSquare

	// Reset DIS1 to 0xFF (no piece selected)
	// DIS2 and DIS3 keep showing the last move
	g.DIS1 = 0xFF

	// Reset digit count for next move
	g.DigitCount = 0
}

// FindPieceAtSquare searches the Board array for a piece at the given square.
// Returns the piece index (0-15) if found, or NoPiece (0xFF) if no piece at that square.
//
// This implements the SEARCH loop from INPUT routine (assembly lines 266-271):
//
//	SEARCH   LDA BOARD,X      ; Load piece position
//	         CMP DIS2         ; Compare with from square
//	         BEQ HERE         ; Found it!
//	         DEX              ; Next piece
//	         BPL SEARCH       ; Loop if X >= 0
//
// Assembly searches from X=$1F down to X=$00, checking both BOARD and BK arrays.
// For Phase 4, we only search the current player's Board array.
func (g *GameState) FindPieceAtSquare(sq board.Square) Piece {
	// Search Board array (indices 0-15)
	for piece := Piece(0); piece < 16; piece++ {
		if g.Board[piece] == sq {
			return piece
		}
	}
	// Not found
	return NoPiece
}

// RotateDigitIntoMove implements the DISMV routine from assembly (lines 625-633).
// This rotates a digit (0-7) into the move display registers DIS2/DIS3.
//
// The assembly shifts DIS3 left 4 bits, then shifts DIS2 left 4 bits with carry,
// then ORs the new digit into DIS3. After 4 digits:
//
//	DIS2 = from_square (first 2 digits)
//	DIS3 = to_square (last 2 digits)
//
// Assembly reference:
//
//	DISMV    LDX #$04          ; Loop 4 times (shift 4 bits)
//	DROL     ASL DIS3          ; Shift DIS3 left
//	         ROL DIS2          ; Shift DIS2 left with carry
//	         DEX
//	         BNE DROL
//	         ORA DIS3          ; OR digit into DIS3
//	         STA DIS3
func (g *GameState) RotateDigitIntoMove(digit uint8) {
	// Shift left 4 bits: DIS3 << 4, carry to DIS2, DIS2 << 4
	// In assembly this is done with 4 iterations of ASL/ROL
	// In Go we can do it directly with bit operations

	// Extract the high nibble of DIS3 (will become low nibble of DIS2)
	carry := (g.DIS3 & 0xF0) >> 4

	// Shift DIS3 left 4 bits
	g.DIS3 = g.DIS3 << 4

	// Shift DIS2 left 4 bits and add carry from DIS3
	g.DIS2 = (g.DIS2 << 4) | carry

	// OR the new digit into DIS3
	g.DIS3 = g.DIS3 | digit
}

// Display prints the chess board in the style of the original POUT routine (line 702).
// The display shows coordinates and piece positions using the 0x88 encoding.
func (g *GameState) Display() {
	_, _ = fmt.Fprintf(g.out, "MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com\r\n")
	_, _ = fmt.Fprintf(g.out, " 00 01 02 03 04 05 06 07\r\n")
	_, _ = fmt.Fprintf(g.out, "-------------------------\r\n")

	// Display ranks 0 to 7 (original displays 00-70)
	// The original scans Y from 0x00 to 0x77 in 0x88 format
	for rank := 0; rank <= 7; rank++ {
		_, _ = fmt.Fprint(g.out, "|")

		for file := 0; file < 8; file++ {
			sq := board.Square((rank << 4) | file)
			piece, found, isWhite := g.FindPieceAt(sq)

			if found {
				_, _ = fmt.Fprint(g.out, GetPieceChar(piece, isWhite))
			} else {
				// Checkerboard pattern for empty squares
				// Original: check if (file + rank) is odd for asterisk
				if (rank+file)%2 == 1 {
					_, _ = fmt.Fprint(g.out, "**")
				} else {
					_, _ = fmt.Fprint(g.out, "  ")
				}
			}

			_, _ = fmt.Fprint(g.out, "|")
		}

		// Print rank number in hex on the right (00, 10, 20, ...)
		_, _ = fmt.Fprintf(g.out, "%X0\r\n", rank)
	}

	_, _ = fmt.Fprintf(g.out, "-------------------------\r\n")
	_, _ = fmt.Fprintf(g.out, " 00 01 02 03 04 05 06 07\r\n")

	// Print LED display (DIS1 DIS2 DIS3)
	_, _ = fmt.Fprintf(g.out, "%02X %02X %02X\r\n", g.DIS1, g.DIS2, g.DIS3)
	_, _ = fmt.Fprintf(g.out, "\r\n")
}
