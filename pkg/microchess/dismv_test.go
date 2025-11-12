// ABOUTME: This file tests the DISMV (digit rotation) routine implementation.
// ABOUTME: It verifies the bit rotation logic matches the 6502 assembly behavior.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestRotateDigitIntoMove_FromZero(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.DIS1 = 0x00
	game.DIS2 = 0x00
	game.DIS3 = 0x00

	// Enter digit sequence: 1, 4, 3, 4 (e2 to e4)
	// Digits are entered as: from_rank, from_file, to_rank, to_file
	// This should build: from=0x14 (rank 1, file 4 = e2), to=0x34 (rank 3, file 4 = e4)

	// After digit '1' (from rank = rank 2)
	game.RotateDigitIntoMove(1)
	assert.Equal(t, uint8(0x00), game.DIS2, "DIS2 after 1st digit")
	assert.Equal(t, uint8(0x01), game.DIS3, "DIS3 after 1st digit")

	// After digit '4' (from file = e)
	game.RotateDigitIntoMove(4)
	assert.Equal(t, uint8(0x00), game.DIS2, "DIS2 after 2nd digit")
	assert.Equal(t, uint8(0x14), game.DIS3, "DIS3 after 2nd digit - should be 'from' square 0x14")

	// After digit '3' (to rank = rank 4)
	game.RotateDigitIntoMove(3)
	assert.Equal(t, uint8(0x01), game.DIS2, "DIS2 after 3rd digit - should have rotated from DIS3")
	assert.Equal(t, uint8(0x43), game.DIS3, "DIS3 after 3rd digit")

	// After digit '4' (to file = e)
	game.RotateDigitIntoMove(4)
	assert.Equal(t, uint8(0x14), game.DIS2, "DIS2 after 4th digit - should be 'from' square")
	assert.Equal(t, uint8(0x34), game.DIS3, "DIS3 after 4th digit - should be 'to' square")
}

func TestRotateDigitIntoMove_FromCC(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	// Start from CC CC CC (after 'C' command)
	game.DIS1 = 0xCC
	game.DIS2 = 0xCC
	game.DIS3 = 0xCC

	// Enter digit sequence: 1, 4, 3, 4 (e2 to e4)
	// Starting from CC CC CC

	// Enter digit '1'
	game.RotateDigitIntoMove(1)
	// The high nibble of CC (C) rotates into DIS2
	// DIS3 becomes C0, then OR 1 = C1
	assert.Equal(t, uint8(0xCC), game.DIS2, "DIS2 after 1st digit from CC")
	assert.Equal(t, uint8(0xC1), game.DIS3, "DIS3 after 1st digit from CC")

	// Enter digit '4'
	game.RotateDigitIntoMove(4)
	assert.Equal(t, uint8(0xCC), game.DIS2, "DIS2 after 2nd digit from CC")
	assert.Equal(t, uint8(0x14), game.DIS3, "DIS3 after 2nd digit from CC - should be from square")

	// Enter digit '3'
	game.RotateDigitIntoMove(3)
	assert.Equal(t, uint8(0xC1), game.DIS2, "DIS2 after 3rd digit from CC")
	assert.Equal(t, uint8(0x43), game.DIS3, "DIS3 after 3rd digit from CC")

	// Enter digit '4'
	game.RotateDigitIntoMove(4)
	assert.Equal(t, uint8(0x14), game.DIS2, "DIS2 after 4th digit from CC - should be from square")
	assert.Equal(t, uint8(0x34), game.DIS3, "DIS3 after 4th digit from CC - should be to square")
}

func TestFindPieceAtSquare(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.SetupBoard()

	// White pawn at 0x14 (e2) should be piece 14 (PiecePawn7 - e-file pawn)
	piece := game.FindPieceAtSquare(0x14)
	assert.Equal(t, Piece(0x0E), piece, "Should find piece 0x0E (14 decimal) at square 0x14")

	// White king at 0x03 should be piece 0 (PieceKing)
	piece = game.FindPieceAtSquare(0x03)
	assert.Equal(t, PieceKing, piece, "Should find king at square 0x03")

	// Empty square should return NoPiece
	piece = game.FindPieceAtSquare(0x24)
	assert.Equal(t, NoPiece, piece, "Should return NoPiece for empty square")
}

func TestExecuteMove_SimplePawnMove(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.SetupBoard()

	// Set up move: e2-e4 (14 -> 34)
	game.DIS2 = 0x14          // from square
	game.DIS3 = 0x34          // to square
	game.SelectedPiece = 0x0E // piece 14 (e-file pawn)
	game.DigitCount = 4
	game.DIS1 = 0x0E

	// Execute move
	game.ExecuteMove()

	// Verify pawn moved
	assert.Equal(t, board.Square(0x34), game.Board[0x0E], "Pawn should be at 0x34")

	// Verify from square is empty
	piece := game.FindPieceAtSquare(0x14)
	assert.Equal(t, NoPiece, piece, "Square 0x14 should be empty")

	// Verify DIS1 reset to 0xFF
	assert.Equal(t, uint8(0xFF), game.DIS1, "DIS1 should reset to 0xFF")

	// Verify digit count reset
	assert.Equal(t, uint8(0), game.DigitCount, "DigitCount should reset to 0")
}

func TestExecuteMove_Capture(t *testing.T) {
	game := NewGame(&bytes.Buffer{})
	game.SetupBoard()

	// First move white pawn e2->e4 (14->34) - need to set it up manually
	game.Board[0x0E] = 0x34 // Move e-pawn to 34

	// Now try to capture: move white pawn e4->e5 (34->54) to capture black pawn
	// First we need to move a black pawn to e5
	// Black e-file pawn is at index 0x0E in BK array, currently at 0x64
	// We'll pretend the black pawn is at 0x54 (e5) for this test
	// Actually, let's set up white pawn at e4 capturing something at e5

	// Set up: white pawn at 0x14, "move" to 0x54 where there's a black pawn
	game.Board[0x0E] = 0x14 // Reset pawn to e2
	// Manually place a black pawn at 0x54 by putting it in BK array
	// But for simplicity, let's just place a white piece there that we'll capture
	game.Board[0x0D] = 0x54 // Put pawn 13 (f-file) at 0x54 as target

	game.DIS2 = 0x14          // from square
	game.DIS3 = 0x54          // to square
	game.SelectedPiece = 0x0E // e-file pawn
	game.DigitCount = 4
	game.DIS1 = 0x0E

	// Execute move (will capture piece 13)
	game.ExecuteMove()

	// Verify moving piece at new location
	assert.Equal(t, board.Square(0x54), game.Board[0x0E], "Pawn should be at 0x54")

	// Verify captured piece marked as captured (0xCC)
	assert.Equal(t, board.Square(0xCC), game.Board[0x0D], "Captured piece should be at 0xCC")
}
