// ABOUTME: This file contains tests for MicroChess types and game state.
// ABOUTME: It verifies piece setup, board initialization, and piece finding logic.

package microchess

import (
	"testing"

	"github.com/matteo/microchess-go/pkg/board"
)

func TestNewGame(t *testing.T) {
	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	// Verify white pieces are in starting positions
	if game.Board[PieceKing] != 0x03 {
		t.Errorf("White king position = 0x%02X, want 0x03", game.Board[PieceKing])
	}
	if game.Board[PieceQueen] != 0x04 {
		t.Errorf("White queen position = 0x%02X, want 0x04", game.Board[PieceQueen])
	}

	// Verify black pieces are in starting positions
	if game.BK[PieceKing] != 0x73 {
		t.Errorf("Black king position = 0x%02X, want 0x73", game.BK[PieceKing])
	}
	if game.BK[PieceQueen] != 0x74 {
		t.Errorf("Black queen position = 0x%02X, want 0x74", game.BK[PieceQueen])
	}

	if game.Reversed {
		t.Error("New game should not be reversed")
	}
}

func TestSetupBoard(t *testing.T) {
	game := &GameState{}
	game.SetupBoard()

	// Check a few key pieces
	expectedWhite := map[Piece]board.Square{
		PieceKing:   0x03,
		PieceQueen:  0x04,
		PieceRook1:  0x00,
		PieceRook2:  0x07,
		PiecePawn1:  0x10,
		PiecePawn8:  0x13,
	}

	for piece, expected := range expectedWhite {
		if game.Board[piece] != expected {
			t.Errorf("White piece %d position = 0x%02X, want 0x%02X", piece, game.Board[piece], expected)
		}
	}

	expectedBlack := map[Piece]board.Square{
		PieceKing:  0x73,
		PieceQueen: 0x74,
		PieceRook1: 0x70,
		PieceRook2: 0x77,
	}

	for piece, expected := range expectedBlack {
		if game.BK[piece] != expected {
			t.Errorf("Black piece %d position = 0x%02X, want 0x%02X", piece, game.BK[piece], expected)
		}
	}
}

func TestFindPieceAt(t *testing.T) {
	game := NewGame()

	tests := []struct {
		square  board.Square
		piece   Piece
		found   bool
		isWhite bool
	}{
		{0x03, PieceKing, true, true},    // White king on d1
		{0x04, PieceQueen, true, true},   // White queen on e1
		{0x73, PieceKing, true, false},   // Black king on d8
		{0x34, NoPiece, false, false},    // Empty square e4
		{0x44, NoPiece, false, false},    // Empty square e5
	}

	for _, tt := range tests {
		piece, found, isWhite := game.FindPieceAt(tt.square)
		if found != tt.found {
			t.Errorf("FindPieceAt(0x%02X) found = %v, want %v", tt.square, found, tt.found)
		}
		if found && piece != tt.piece {
			t.Errorf("FindPieceAt(0x%02X) piece = %d, want %d", tt.square, piece, tt.piece)
		}
		if found && isWhite != tt.isWhite {
			t.Errorf("FindPieceAt(0x%02X) isWhite = %v, want %v", tt.square, isWhite, tt.isWhite)
		}
	}
}

func TestGetPieceChar(t *testing.T) {
	tests := []struct {
		piece   Piece
		isWhite bool
		want    string
	}{
		{PieceKing, true, "WK"},
		{PieceKing, false, "BK"},
		{PieceQueen, true, "WQ"},
		{PieceRook1, true, "WR"},
		{PieceBishop2, false, "BB"},
		{PieceKnight1, true, "WN"},
		{PiecePawn1, false, "BP"},
	}

	for _, tt := range tests {
		got := GetPieceChar(tt.piece, tt.isWhite)
		if got != tt.want {
			t.Errorf("GetPieceChar(%d, %v) = %q, want %q", tt.piece, tt.isWhite, got, tt.want)
		}
	}
}
