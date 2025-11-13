// ABOUTME: This file implements the 'L' command to list all legal moves.
// ABOUTME: This is a NEW command not in the original 6502 code, used for debugging and learning.

package microchess

import (
	"fmt"
	"sort"

	"github.com/matteo/microchess-go/pkg/board"
)

// ListLegalMoves generates and displays all legal moves for the current position.
// This is the handler for the 'L' command (NEW - not in original).
//
// Output format matches original's LED display style: hex coordinates (14 34 for e2-e4)
func (g *GameState) ListLegalMoves() {
	// Collect all moves
	var moves []Move

	// Use GNM with a callback that collects moves
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{
			From:  from,
			To:    to,
			Piece: piece,
		})
	})

	// Sort moves for consistent output (by from square, then to square)
	sort.Slice(moves, func(i, j int) bool {
		if moves[i].From != moves[j].From {
			return moves[i].From < moves[j].From
		}
		return moves[i].To < moves[j].To
	})

	// Display moves
	_, _ = fmt.Fprintf(g.out, "\r\nLegal moves (%d):\r\n", len(moves))

	for _, move := range moves {
		// Show in algebraic notation for clarity (e2e4)
		// TODO: Fix hex formatting issue and use 0x88 format instead
		from := move.From.String()
		to := move.To.String()
		pieceChar := getPieceTypeChar(move.Piece)
		_, _ = fmt.Fprintf(g.out, "  %s%s  (%c)\n", from, to, pieceChar)
	}

	_, _ = fmt.Fprintf(g.out, "\r\n")
}

// getPieceTypeChar returns a single character representing the piece type.
// K=King, Q=Queen, R=Rook, B=Bishop, N=Knight, P=Pawn
func getPieceTypeChar(piece Piece) rune {
	switch piece {
	case PieceKing:
		return 'K'
	case PieceQueen:
		return 'Q'
	case PieceRook1, PieceRook2:
		return 'R'
	case PieceBishop1, PieceBishop2:
		return 'B'
	case PieceKnight1, PieceKnight2:
		return 'N'
	default: // Pawns
		return 'P'
	}
}
