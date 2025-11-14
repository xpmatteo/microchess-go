package microchess

import "github.com/matteo/microchess-go/pkg/board"

// evalInputs bundles every byte STRATGY and CKMATE need to evaluate a move.
// Keeping this as a plain data struct lets us unit-test the pure 8-bit math
// separately from the rest of the engine.
type evalInputs struct {
	WMOB, WMAXC, WCC     uint8
	WCAP0, WCAP1, WCAP2  uint8
	PMOB, PMAXC, PCC     uint8
	BMOB, BMAXC, BMCC    uint8
	BCAP0, BCAP1, BCAP2  uint8
	WMAXP                Piece
	MovingPiece          Piece
	FromSquare, ToSquare board.Square
}

// eval1 replicates the 6502 STRATGY + CKMATE logic using the counters above.
// Implementation TBD.
func eval1(in evalInputs) uint8 {
	panic("eval1 not implemented yet")
}
