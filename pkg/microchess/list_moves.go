// ABOUTME: This file implements the 'L' command to list all legal moves.
// ABOUTME: This is a NEW command not in the original 6502 code, used for debugging and learning.

package microchess

import (
	"fmt"

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
	// Moves are collected in GNM's natural generation order (piece 15 -> 0)
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{
			From:  from,
			To:    to,
			Piece: piece,
		})
	})

	// Display moves in hex format matching LED display style
	// Format: "- FF TT" where FF is from square, TT is to square (both in hex)
	// Note: Explicit uint8() cast needed for fmt.Fprintf variadic arguments
	for _, move := range moves {
		_, _ = fmt.Fprintf(g.out, "- %02X %02X\r\n", uint8(move.From), uint8(move.To))
	}
}
