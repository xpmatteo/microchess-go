// ABOUTME: This file contains tests for the Square type and 0x88 board representation.
// ABOUTME: It verifies coordinate conversion, validation, and string formatting.

package board

import "testing"

func TestSquareIsValid(t *testing.T) {
	tests := []struct {
		square Square
		valid  bool
	}{
		{0x00, true},  // a1
		{0x07, true},  // h1
		{0x70, true},  // a8
		{0x77, true},  // h8
		{0x34, true},  // e4
		{0x08, false}, // off board (file 8)
		{0x80, false}, // off board (rank 8)
		{0x88, false}, // off board (both)
		{0xFF, false}, // invalid
	}

	for _, tt := range tests {
		if got := tt.square.IsValid(); got != tt.valid {
			t.Errorf("Square(0x%02X).IsValid() = %v, want %v", tt.square, got, tt.valid)
		}
	}
}

func TestSquareRankFile(t *testing.T) {
	tests := []struct {
		square Square
		rank   int
		file   int
	}{
		{0x00, 0, 0}, // a1
		{0x07, 0, 7}, // h1
		{0x70, 7, 0}, // a8
		{0x77, 7, 7}, // h8
		{0x34, 3, 4}, // e4
	}

	for _, tt := range tests {
		if got := tt.square.Rank(); got != tt.rank {
			t.Errorf("Square(0x%02X).Rank() = %d, want %d", tt.square, got, tt.rank)
		}
		if got := tt.square.File(); got != tt.file {
			t.Errorf("Square(0x%02X).File() = %d, want %d", tt.square, got, tt.file)
		}
	}
}

func TestSquareString(t *testing.T) {
	tests := []struct {
		square Square
		str    string
	}{
		{0x00, "a1"},
		{0x07, "h1"},
		{0x70, "a8"},
		{0x77, "h8"},
		{0x34, "e4"},
		{0x08, "??"}, // invalid
	}

	for _, tt := range tests {
		if got := tt.square.String(); got != tt.str {
			t.Errorf("Square(0x%02X).String() = %q, want %q", tt.square, got, tt.str)
		}
	}
}

func TestParseSquare(t *testing.T) {
	tests := []struct {
		str     string
		square  Square
		wantErr bool
	}{
		{"a1", 0x00, false},
		{"h1", 0x07, false},
		{"a8", 0x70, false},
		{"h8", 0x77, false},
		{"e4", 0x34, false},
		{"i1", 0, true},  // invalid file
		{"a9", 0, true},  // invalid rank
		{"e", 0, true},   // too short
		{"e44", 0, true}, // too long
	}

	for _, tt := range tests {
		got, err := ParseSquare(tt.str)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseSquare(%q) expected error, got nil", tt.str)
			}
		} else {
			if err != nil {
				t.Errorf("ParseSquare(%q) unexpected error: %v", tt.str, err)
			}
			if got != tt.square {
				t.Errorf("ParseSquare(%q) = 0x%02X, want 0x%02X", tt.str, got, tt.square)
			}
		}
	}
}
