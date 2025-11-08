// ABOUTME: This file defines core MicroChess types including pieces and game state.
// ABOUTME: It maintains the same piece indexing and data structures as the original assembly.

package microchess

import "github.com/matteo/microchess-go/pkg/board"

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

// NewGame creates a new game with the initial position set up.
func NewGame() *GameState {
	g := &GameState{}
	g.SetupBoard()
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
func (g *GameState) FindPieceAt(sq board.Square) (Piece, bool, bool) {
	// Check white pieces (current board)
	for i := Piece(0); i < 16; i++ {
		if g.Board[i] == sq {
			return i, true, true // piece, found, isWhite
		}
	}
	// Check black pieces (alternate board)
	for i := Piece(0); i < 16; i++ {
		if g.BK[i] == sq {
			return i, true, false // piece, found, isBlack
		}
	}
	return NoPiece, false, false
}
