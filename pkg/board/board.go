// ABOUTME: This file defines the Square type and 0x88 board representation.
// ABOUTME: It provides helper functions for working with chess board coordinates.

package board

import "fmt"

// Square represents a position on the chess board using 0x88 representation.
// In 0x88 encoding: square = (rank << 4) | file
// Where rank and file are each 0-7.
// This allows fast edge detection: (square & 0x88) != 0 means off-board.
//
// Reference: MicroChess assembly lines 50-60 (board encoding)
type Square uint8

// Board square encoding constants for reference positions
const (
	SquareA1 Square = 0x00
	SquareH1 Square = 0x07
	SquareA8 Square = 0x70
	SquareH8 Square = 0x77
)

// IsValid checks if a square is on the board using 0x88 trick.
// Assembly equivalent: AND #$88 (line 407+)
func (s Square) IsValid() bool {
	return (s & 0x88) == 0
}

// Rank returns the rank (0-7) of the square, where 0 is rank 1 and 7 is rank 8.
func (s Square) Rank() int {
	return int(s >> 4)
}

// File returns the file (0-7) of the square, where 0 is file 'a' and 7 is file 'h'.
func (s Square) File() int {
	return int(s & 0x07)
}

// String converts a square to algebraic notation (e.g., "e4").
func (s Square) String() string {
	if !s.IsValid() {
		return "??"
	}
	file := 'a' + rune(s.File())
	rank := '1' + rune(s.Rank())
	return fmt.Sprintf("%c%c", file, rank)
}

// ParseSquare converts algebraic notation (e.g., "e4") to a Square.
func ParseSquare(str string) (Square, error) {
	if len(str) != 2 {
		return 0, fmt.Errorf("invalid square format: %s", str)
	}

	file := str[0] - 'a'
	rank := str[1] - '1'

	if file < 0 || file > 7 || rank < 0 || rank > 7 { // nolint:staticcheck
		return 0, fmt.Errorf("square out of range: %s", str)
	}

	return Square((rank << 4) | file), nil
}
