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
	// Indices 0-15 represent pieces in Board array (initially at bottom, ranks 0-1)
	Board [16]board.Square

	// BK is the alternate board used during REVERSE
	// Indices 0-15 represent pieces in BK array (initially at top, ranks 6-7)
	// Assembly: $60-$6F
	BK [16]board.Square

	// REV tracks if board is reversed (0 = bottom's view, non-zero = top's view)
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

	// Move generation state (used by GNM)
	// Assembly: $B0, $B1, $B6
	MovePiece  Piece        // Current piece being processed by GNM (assembly: PIECE at $B0)
	MoveSquare board.Square // Working square for move calculation (assembly: SQUARE at $B1)
	MoveN      uint8        // Move direction index into MOVEX (assembly: MOVEN at $B6)

	// STATE machine and check detection
	// Assembly: $B5, $B4
	State  int8  // STATE machine value for analysis depth control (assembly: STATE at $B5)
	InChek uint8 // Check detection flag: 0xF9=safe, 0x00=king capturable (assembly: INCHEK at $B4)

	// Move history stack (replaces assembly's SP2 dual-stack mechanism)
	// Used by MOVE/UMOVE to make and unmake trial moves during CHKCHK
	MoveHistory []MoveRecord

	// Evaluation counters (assembly: COUNT array at $DE-$EE)
	// These track mobility, captures, and threats for position evaluation.
	// Indexed by STATE value offset (e.g., STATE=4 uses index 4).
	// Assembly reference: doc/DATA_STRUCTURES.md lines 140-173
	Mobility      [16]uint8 // MOB: Legal move count per state
	MaxCapture    [16]uint8 // MAXC: Highest value piece capturable per state
	CaptureCount  [16]uint8 // CC: Total capture count per state
	PieceCaptured [16]Piece // PCAP: Index of best capturable piece per state

	// Named counter instances for key positions (assembly: $EB-$EE, $E3-$E6, $EF-$F2)
	// These are aliases into the arrays above for specific STATE values
	// Assembly reference: doc/DATA_STRUCTURES.md lines 143-158
	WMOB, WMAXC, WCC uint8 // White's mobility/captures/count (STATE=11)
	WMAXP            Piece // White's best capturable piece
	BMOB, BMAXC, BMCC uint8 // Black's mobility/captures/count (STATE=3)
	BMAXP             Piece // Black's best capturable piece
	PMOB, PMAXC, PCC uint8 // Position mobility/captures/count (STATE=15)
	PCP              Piece // Position captured piece

	// Capture depth counters (assembly: $DD-$E2)
	// Track captured piece values at different search depths
	// Assembly reference: doc/DATA_STRUCTURES.md lines 160-171
	WCAP0, WCAP1, WCAP2 uint8 // White's captures at depths 0,1,2
	BCAP0, BCAP1, BCAP2 uint8 // Black's captures at depths 0,1,2
	XMAXC               uint8 // Saved maximum capture value (assembly: $E8)

	// Best move tracking (assembly: BESTP/$FB, BESTV/$FA, BESTM/$F9)
	// Note: These overlap with DIS1/DIS2/DIS3 in assembly to save page zero space
	// Assembly reference: doc/DATA_STRUCTURES.md lines 176-189
	BestPiece  Piece        // Best piece to move (piece index 0-15)
	BestValue  uint8        // Best move evaluation score
	BestSquare board.Square // Best destination square

	// I/O for display and input
	out io.Writer
}

// MoveRecord stores the state of a move for undo purposes.
// This replaces the assembly's SP2 stack mechanism (dual stack for game state).
// Assembly reference: MOVE routine (lines 511-539) pushes to SP2 stack
type MoveRecord struct {
	FromSquare     board.Square // Original square of moving piece
	ToSquare       board.Square // Destination square
	MovingPiece    Piece        // Index of piece being moved
	CapturedPiece  Piece        // Index of captured piece (NoPiece if none)
	CapturedSquare board.Square // Original square of captured piece
	MoveN          uint8        // MOVEN value at time of move
}

// MOVEX is the direction offset table used for move generation and validation.
// This matches the MOVEX table from assembly (line 875-876).
//
// Each entry represents a directional offset to add to a square in 0x88 format:
//
//	Index 0: $00 (null move - unused)
//	Index 1: $F0 (-16 decimal, one rank down)
//	Index 2: $FF (-1 horizontal in 0x88, one file left)
//	Index 3: $01 (+1 horizontal, one file right)
//	Index 4: $10 (+16 decimal, one rank up)
//	Index 5-8: Diagonal moves (±1 rank, ±1 file)
//	Index 9-16: Knight moves (±2 ranks/±1 file and ±1 rank/±2 files)
//
// These offsets work with 0x88 board representation where:
//   - Ranks are in bits 4-6 (0x00, 0x10, 0x20, ..., 0x70)
//   - Files are in bits 0-2 (0x00-0x07)
//   - Adding $10 moves up one rank, $01 moves right one file
//
// Assembly reference: line 875-876
var MOVEX = [17]int8{
	0x00,  // 0: null
	-0x10, // 1: down ($F0 in unsigned byte)
	-0x01, // 2: left ($FF in unsigned byte)
	0x01,  // 3: right
	0x10,  // 4: up
	0x11,  // 5: up-right
	0x0F,  // 6: up-left
	-0x11, // 7: down-left ($EF in unsigned byte)
	-0x0F, // 8: down-right ($F1 in unsigned byte)
	-0x21, // 9: knight down-down-left ($DF in unsigned byte)
	-0x1F, // 10: knight down-down-right ($E1 in unsigned byte)
	-0x12, // 11: knight down-left-left ($EE in unsigned byte)
	-0x0E, // 12: knight down-right-right ($F2 in unsigned byte)
	0x12,  // 13: knight up-left-left
	0x0E,  // 14: knight up-right-right
	0x1F,  // 15: knight up-up-left
	0x21,  // 16: knight up-up-right
}

// POINTS is the piece value table used for capture evaluation.
// This matches the POINTS table from assembly (line 878-879).
//
// Values represent the relative worth of each piece:
//   - King: 11 (special value, effectively infinite)
//   - Queen: 10
//   - Rook: 6
//   - Bishop: 4
//   - Knight: 4
//   - Pawn: 2
//
// These values are used by COUNTS to evaluate captures and by STRATGY
// for material balance calculation.
//
// Assembly reference: line 878-879 (POINTS table)
var POINTS = [16]uint8{
	11,              // 0: King (special - should never be captured)
	10,              // 1: Queen
	6, 6,            // 2-3: Rooks
	4, 4,            // 4-5: Bishops
	4, 4,            // 6-7: Knights
	2, 2, 2, 2, 2, 2, 2, 2, // 8-15: Pawns
}

// InitialSetup contains the starting positions for all pieces.
// This matches the SETW/SETB tables from the assembly (lines 672-687).
//
// Indices 0-15: Board array pieces (initially at bottom, ranks 0-1)
// Indices 16-31: BK array pieces (initially at top, ranks 6-7)
// In actual play, this is copied to both arrays with coordinate transformation for top pieces.
var InitialSetup = [32]board.Square{
	// Board array pieces (indices 0-7): K, Q, R, R, B, B, N, N
	0x03, 0x04, 0x00, 0x07, 0x02, 0x05, 0x01, 0x06,
	// Board array pawns (indices 8-15): specific file ordering from assembly
	0x10, 0x17, 0x11, 0x16, 0x12, 0x15, 0x14, 0x13,
	// BK array pieces (indices 16-23): K, Q, R, R, B, B, N, N
	0x73, 0x74, 0x70, 0x77, 0x72, 0x75, 0x71, 0x76,
	// BK array pawns (indices 24-31)
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
	// However, to match the original's display, put one BK pawn at 00
	// This is the "garbage" state the original shows
	g.BK[8] = 0x00 // BK array pawn at position 00

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
	// Copy Board array pieces (bottom of board, ranks 0-1)
	for i := 0; i < 16; i++ {
		g.Board[i] = InitialSetup[i]
	}
	// Copy BK array pieces (top of board, ranks 6-7)
	for i := 0; i < 16; i++ {
		g.BK[i] = InitialSetup[i+16]
	}
	// NOTE: The Reversed flag is NOT reset here. The original assembly SETUP routine
	// (line 116-126) does not modify the REV flag. Only the REVERSE routine toggles it.
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
// The isWhite flag indicates the COLOR TO DISPLAY (white/black) which depends on Reversed flag.
//
// In the original assembly (POUT line 750), when REV flag is set, it uses different
// color characters (cpl+16 vs cpl) to flip the white/black display.
func (g *GameState) FindPieceAt(sq board.Square) (piece Piece, found bool, isWhite bool) {
	// Check pieces in Board array (bottom of board initially)
	for i := Piece(0); i < 16; i++ {
		if g.Board[i] == sq {
			// Board pieces display as white when REV=0, black when REV!=0
			// This matches assembly line 750: REV flag determines color character used
			return i, true, !g.Reversed // piece, found, isWhite=(REV==0)
		}
	}
	// Check pieces in BK array (top of board initially)
	for i := Piece(0); i < 16; i++ {
		if g.BK[i] == sq {
			// BK pieces display as black when REV=0, white when REV!=0
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

	case 'L':
		// List legal moves (NEW command - not in original)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed 'L'
		g.ListLegalMoves()
		return true

	case 'S':
		// Show position evaluation (NEW command - not in original)
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed 'S'
		g.ShowEvaluation()
		return true

	case '\r', '\n':
		// Enter/Return key
		_, _ = fmt.Fprintln(g.out, "\r") // Clean newline after echoed char

		// If we have 4 or more digits entered, execute the move using the last 4 digits
		// The last 4 digits are stored in DIS2 (from square) and DIS3 (to square)
		// This allows users to enter extra digits (5, 6, 7, 8, ...) and still execute
		if g.DigitCount >= 4 {
			g.ExecuteMove()
		}
		// Always display board after carriage return (even if no move executed)
		// This matches 6502 behavior
		g.Display()
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

		// If we've completed 4 digits previously (DigitCount >= 4),
		// reset DIS1 to FF but keep the digit rotation going (don't reset DigitCount)
		// This allows entering 5+ digits - they keep rotating through DIS2/DIS3
		if g.DigitCount >= 4 {
			g.DIS1 = 0xFF
			// Don't reset DigitCount - just let it keep incrementing
		}

		// Rotate digit into move display (DISMV routine, line 625)
		g.RotateDigitIntoMove(digit)
		g.DigitCount++

		// After 4 or more digits, always find piece at from square (DIS2)
		// The last 4 digits in the rolling buffer define the current move
		// Assembly lines 266-272 (SEARCH loop in INPUT)
		if g.DigitCount >= 4 {
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
		if capturedPiece < 16 {
			// Board array piece (indices 0-15)
			g.Board[capturedPiece] = 0xCC
		} else {
			// BK array piece (indices 16-31, stored in BK[0-15])
			g.BK[capturedPiece-16] = 0xCC
		}
	}

	// Move the selected piece to target square
	// Assembly line 535: STA BOARD,X (stores SQUARE into PIECE's position)
	if g.SelectedPiece < 16 {
		// Board array piece (indices 0-15)
		g.Board[g.SelectedPiece] = targetSquare
	} else {
		// BK array piece (indices 16-31, stored in BK[0-15])
		g.BK[g.SelectedPiece-16] = targetSquare
	}

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
	// Search Board array (indices 0-15) - pieces at bottom initially
	for piece := Piece(0); piece < 16; piece++ {
		if g.Board[piece] == sq {
			return piece
		}
	}
	// Search BK array (indices 0-15) - pieces at top initially
	// Return piece index + 16 to match assembly convention (BK pieces are 0x10-0x1F)
	for piece := Piece(0); piece < 16; piece++ {
		if g.BK[piece] == sq {
			return piece + 16
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
