// ABOUTME: This file contains unit tests for the GNM (Generate New Move) routine.
// ABOUTME: Tests verify move counts and correctness for known chess positions.

package microchess

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
)

// TestGNM_StartingPosition verifies that the starting position generates exactly 20 legal moves.
// This is a well-known chess fact: white has 20 legal moves from the initial position.
func TestGNM_StartingPosition(t *testing.T) {
	// Create game and set up starting position
	var buf bytes.Buffer
	g := NewGame(&buf)
	g.SetupBoard()

	// Collect all moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// White should have exactly 20 legal moves from starting position:
	// - 8 pawns × 2 moves each (one square or two squares forward) = 16 moves
	// - 2 knights × 2 moves each (knights can jump) = 4 moves
	// - Rooks, bishops, queen, king are all blocked
	expectedCount := 20
	if len(moves) != expectedCount {
		t.Errorf("Starting position should have %d moves, got %d", expectedCount, len(moves))
		t.Logf("Moves generated:")
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_PawnMoves verifies pawn move generation.
func TestGNM_PawnMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Set up a position with just one pawn on e2 (0x14)
	// Piece 14 (PiecePawn7) is the e-file pawn
	g.Board[PiecePawn7] = 0x14 // e2
	// Mark all other pieces as captured/off-board
	for i := Piece(0); i < 16; i++ {
		if i != PiecePawn7 {
			g.Board[i] = 0xCC
		}
	}

	// Clear opponent pieces
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Pawn on e2 should have 2 moves: e3 (0x24) and e4 (0x34)
	expectedCount := 2
	if len(moves) != expectedCount {
		t.Errorf("Pawn on e2 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}

	// Verify the moves are correct
	expectedMoves := map[board.Square]bool{
		0x24: true, // e3
		0x34: true, // e4
	}
	for _, m := range moves {
		if !expectedMoves[m.To] {
			t.Errorf("Unexpected pawn move to %02X", m.To)
		}
	}
}

// TestGNM_KnightMoves verifies knight move generation.
func TestGNM_KnightMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place knight on e4 (0x34) - center of board
	g.Board[PieceKnight1] = 0x34
	// Mark all other pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PieceKnight1 {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Knight on e4 (center) should have 8 legal moves
	expectedCount := 8
	if len(moves) != expectedCount {
		t.Errorf("Knight on e4 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_RookMoves verifies rook move generation.
func TestGNM_RookMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place rook on d4 (0x33)
	g.Board[PieceRook1] = 0x33
	// Mark all other pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PieceRook1 {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Rook on d4 should have 14 moves:
	// - 3 squares left (c4, b4, a4)
	// - 4 squares right (e4, f4, g4, h4)
	// - 3 squares up (d5, d6, d7)
	// - 4 squares down (d3, d2, d1)
	expectedCount := 14
	if len(moves) != expectedCount {
		t.Errorf("Rook on d4 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_BishopMoves verifies bishop move generation.
func TestGNM_BishopMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place bishop on d4 (0x33)
	g.Board[PieceBishop1] = 0x33
	// Mark all other pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PieceBishop1 {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Bishop on d4 should have 13 moves:
	// - Up-right diagonal: e5, f6, g7 (3)
	// - Up-left diagonal: c5, b6, a7 (3)
	// - Down-right diagonal: e3, f2, g1 (3)
	// - Down-left diagonal: c3, b2, a1 (3)
	// Total: 13 moves
	expectedCount := 13
	if len(moves) != expectedCount {
		t.Errorf("Bishop on d4 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_QueenMoves verifies queen move generation.
func TestGNM_QueenMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place queen on d4 (0x33)
	g.Board[PieceQueen] = 0x33
	// Mark all other pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PieceQueen {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Queen on d4 combines rook + bishop moves: 14 + 13 = 27 moves
	expectedCount := 27
	if len(moves) != expectedCount {
		t.Errorf("Queen on d4 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_KingMoves verifies king move generation.
func TestGNM_KingMoves(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place king on e4 (0x34) - center of board
	g.Board[PieceKing] = 0x34
	// Mark all other pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PieceKing {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// King on e4 (not at edge) should have 8 moves
	expectedCount := 8
	if len(moves) != expectedCount {
		t.Errorf("King on e4 should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_PawnCapture verifies pawn diagonal capture.
func TestGNM_PawnCapture(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place white pawn on e4 (0x34)
	g.Board[PiecePawn7] = 0x34
	// Mark all other white pieces as off-board
	for i := Piece(0); i < 16; i++ {
		if i != PiecePawn7 {
			g.Board[i] = 0xCC
		}
	}

	// Place black pawns on d5 (0x43) and f5 (0x45) - capturable by white pawn
	// Mark all black pieces as off-board first
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}
	// Then place the two pawns
	g.BK[8] = 0x43 // d5
	g.BK[9] = 0x45 // f5

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		moves = append(moves, Move{From: from, To: to, Piece: piece})
	})

	// Pawn should have 3 moves: e5 (forward), d5 (capture), f5 (capture)
	expectedCount := 3
	if len(moves) != expectedCount {
		t.Errorf("Pawn on e4 with captures should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}

// TestGNM_BlockedPieces verifies that blocked pieces generate no moves.
func TestGNM_BlockedPieces(t *testing.T) {
	var buf bytes.Buffer
	g := NewGame(&buf)

	// Place king surrounded by own pieces (completely blocked)
	g.Board[PieceKing] = 0x33  // d4
	g.Board[PiecePawn1] = 0x23 // d3
	g.Board[PiecePawn2] = 0x43 // d5
	g.Board[PiecePawn3] = 0x32 // c4
	g.Board[PiecePawn4] = 0x34 // e4
	g.Board[PiecePawn5] = 0x22 // c3
	g.Board[PiecePawn6] = 0x24 // e3
	g.Board[PiecePawn7] = 0x42 // c5
	g.Board[PiecePawn8] = 0x44 // e5

	// Mark other pieces as off-board
	for i := Piece(1); i < 16; i++ {
		if i < 8 {
			g.Board[i] = 0xCC
		}
	}
	for i := 0; i < 16; i++ {
		g.BK[i] = 0xCC
	}

	// Collect moves
	var moves []Move
	g.GNM(func(from, to board.Square, piece Piece) {
		if piece == PieceKing {
			moves = append(moves, Move{From: from, To: to, Piece: piece})
		}
	})

	// King should have 0 legal moves (completely surrounded)
	expectedCount := 0
	if len(moves) != expectedCount {
		t.Errorf("Blocked king should have %d moves, got %d", expectedCount, len(moves))
		for _, m := range moves {
			t.Logf("  %02X -> %02X (piece %d)", m.From, m.To, m.Piece)
		}
	}
}
