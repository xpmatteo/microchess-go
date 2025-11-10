// ABOUTME: This file contains tests for the Square type and 0x88 board representation.
// ABOUTME: It verifies coordinate conversion, validation, and string formatting.

package board

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSquareIsValid(t *testing.T) {
	tests := []struct {
		name   string
		square Square
		valid  bool
	}{
		{"a1", 0x00, true},
		{"h1", 0x07, true},
		{"a8", 0x70, true},
		{"h8", 0x77, true},
		{"e4", 0x34, true},
		{"off board file 8", 0x08, false},
		{"off board rank 8", 0x80, false},
		{"off board both", 0x88, false},
		{"invalid", 0xFF, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.square.IsValid()
			assert.Equal(t, tt.valid, got, "Square(0x%02X).IsValid()", tt.square)
		})
	}
}

func TestSquareRankFile(t *testing.T) {
	tests := []struct {
		name   string
		square Square
		rank   int
		file   int
	}{
		{"a1", 0x00, 0, 0},
		{"h1", 0x07, 0, 7},
		{"a8", 0x70, 7, 0},
		{"h8", 0x77, 7, 7},
		{"e4", 0x34, 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.rank, tt.square.Rank(), "Square(0x%02X).Rank()", tt.square)
			assert.Equal(t, tt.file, tt.square.File(), "Square(0x%02X).File()", tt.square)
		})
	}
}

func TestSquareString(t *testing.T) {
	tests := []struct {
		name   string
		square Square
		str    string
	}{
		{"a1", 0x00, "a1"},
		{"h1", 0x07, "h1"},
		{"a8", 0x70, "a8"},
		{"h8", 0x77, "h8"},
		{"e4", 0x34, "e4"},
		{"invalid", 0x08, "??"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.square.String()
			assert.Equal(t, tt.str, got, "Square(0x%02X).String()", tt.square)
		})
	}
}

func TestParseSquare(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		square  Square
		wantErr bool
	}{
		{"a1", "a1", 0x00, false},
		{"h1", "h1", 0x07, false},
		{"a8", "a8", 0x70, false},
		{"h8", "h8", 0x77, false},
		{"e4", "e4", 0x34, false},
		{"invalid file", "i1", 0, true},
		{"invalid rank", "a9", 0, true},
		{"too short", "e", 0, true},
		{"too long", "e44", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSquare(tt.str)
			if tt.wantErr {
				assert.Error(t, err, "ParseSquare(%q) should return error", tt.str)
			} else {
				assert.NoError(t, err, "ParseSquare(%q) should not return error", tt.str)
				assert.Equal(t, tt.square, got, "ParseSquare(%q)", tt.str)
			}
		})
	}
}
