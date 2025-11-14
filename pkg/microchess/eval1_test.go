package microchess

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/matteo/microchess-go/pkg/board"
)

func TestEval1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   evalInputs
		want uint8
	}{
		{
			name: "neutral baseline",
			in: evalInputs{
				MovingPiece: PieceKnight1,
				FromSquare:  board.Square(0x20),
				ToSquare:    board.Square(0x30),
			},
			want: 0xD0,
		},
		{
			name: "clamped underflow",
			in: evalInputs{
				PMAXC:       200,
				MovingPiece: PieceBishop1,
				FromSquare:  board.Square(0x20),
				ToSquare:    board.Square(0x21),
			},
			want: 0xB0,
		},
		{
			name: "exchange emphasis",
			in: evalInputs{
				WMOB:        10,
				WMAXC:       8,
				WCC:         6,
				WCAP0:       9,
				WCAP1:       4,
				WCAP2:       2,
				PMAXC:       7,
				PCC:         3,
				PMOB:        5,
				BMOB:        6,
				BMAXC:       5,
				BMCC:        7,
				BCAP0:       4,
				BCAP1:       3,
				BCAP2:       2,
				MovingPiece: PieceBishop1,
				FromSquare:  board.Square(0x24),
				ToSquare:    board.Square(0x44),
			},
			want: 0xE1,
		},
		{
			name: "center bonus",
			in: evalInputs{
				MovingPiece: PieceKnight2,
				FromSquare:  board.Square(0x22),
				ToSquare:    board.Square(0x33),
			},
			want: 0xD2,
		},
		{
			name: "development bonus",
			in: evalInputs{
				MovingPiece: PieceBishop1,
				FromSquare:  board.Square(0x05),
				ToSquare:    board.Square(0x24),
			},
			want: 0xD2,
		},
		{
			name: "king capture override",
			in: evalInputs{
				BMAXC:       POINTS[PieceKing],
				MovingPiece: PieceQueen,
				FromSquare:  board.Square(0x40),
				ToSquare:    board.Square(0x50),
			},
			want: 0x00,
		},
		{
			name: "mate override",
			in: evalInputs{
				BMAXC:       10,
				WMOB:        5,
				WMAXP:       PieceKing,
				MovingPiece: PieceQueen,
				FromSquare:  board.Square(0x40),
				ToSquare:    board.Square(0x51),
			},
			want: 0xFF,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, eval1(tt.in))
		})
	}
}
