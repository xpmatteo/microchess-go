// ABOUTME: This file implements position evaluation routines from MicroChess.
// ABOUTME: Includes GNMZ (clear counters), COUNTS (mobility tracking), and STRATGY (evaluation formula).

package microchess

import "fmt"

// GNMZ clears all evaluation counters and calls GNM (assembly line 280).
//
// This routine initializes all evaluation state before move generation.
// It clears the COUNT array ($DE-$EE, 17 bytes) which includes:
//   - Mobility counters (MOB)
//   - Maximum capture values (MAXC)
//   - Capture counts (CC)
//   - Piece captured indices (PCAP)
//
// After clearing, it calls GNM to generate all moves, which will populate
// the counters via the COUNTS routine (called from JANUS).
//
// Assembly reference:
//
//	GNMZ     LDX #$10       ; 16 counters
//	GNMZ2    LDA #$00
//	         STA COUNT,X    ; Clear COUNT array
//	         DEX
//	         BPL GNMZ2
//	         JSR GNM        ; Generate moves
//	         RTS
//
// Assembly line: 280-285
func (g *GameState) GNMZ() {
	// Clear all 16 counter array elements
	for i := 0; i < 16; i++ {
		g.Mobility[i] = 0
		g.MaxCapture[i] = 0
		g.CaptureCount[i] = 0
		g.PieceCaptured[i] = NoPiece
	}

	// Also clear the named counter instances
	g.WMOB, g.WMAXC, g.WCC = 0, 0, 0
	g.WMAXP = NoPiece
	g.BMOB, g.BMAXC, g.BMCC = 0, 0, 0
	g.BMAXP = NoPiece
	g.PMOB, g.PMAXC, g.PCC = 0, 0, 0
	g.PCP = NoPiece

	// Clear capture depth counters
	g.WCAP0, g.WCAP1, g.WCAP2 = 0, 0, 0
	g.BCAP0, g.BCAP1, g.BCAP2 = 0, 0, 0
	g.XMAXC = 0

	// Generate all moves (will call COUNTS via JANUS for each move)
	g.GNM(nil) // nil callback - COUNTS is called internally by JANUS
}

// COUNTS implements the mobility and capture counting logic (assembly line 169).
//
// This routine is called by JANUS for each pseudo-legal move generated.
// It accumulates evaluation data into state-indexed counter arrays.
//
// Algorithm (from assembly lines 169-222):
//  1. If PIECE==0 AND STATE==8, skip (don't count black max capture for white)
//  2. Increment MOB[STATE] (mobility counter)
//  3. If piece is Queen (PIECE==1), increment MOB[STATE] again (queens count double!)
//  4. If move is capture (last CMOVE had V flag set):
//     a. Search BK array to find captured piece index
//     b. Get piece value from POINTS table
//     c. If value > MAXC[STATE], update MAXC and PCAP
//     d. Add value to CC[STATE] (total capture count)
//  5. If STATE==4, call ON4 (full position analysis)
//  6. If STATE<0, call TREE (recursive capture analysis)
//
// Assembly reference:
//
//	COUNTS   LDA PIECE
//	         BEQ OVER       ; Skip if PIECE=0 AND STATE=8
//	         CPX #$08
//	         BNE OVER
//	         CMP BMAXP
//	         BEQ XRT
//	OVER     INC MOB,X      ; Increment mobility
//	         CMP #$01       ; Is it queen?
//	         BNE NOQ
//	         INC MOB,X      ; Queens count double
//	NOQ      BVC NOCAP      ; No capture? Skip
//	         ... (capture handling)
//
// Assembly line: 169-222
func (g *GameState) COUNTS(captureFlag bool) {
	// Get STATE as index into counter arrays
	// Assembly uses X register which holds STATE value
	stateIdx := int(g.State)
	if stateIdx < 0 {
		stateIdx = 16 + stateIdx // Handle negative indices
	}
	if stateIdx < 0 || stateIdx >= 16 {
		return // Safety check
	}

	// Special case: If PIECE==0 AND STATE==8, skip BLK MAX CAP count
	// Assembly lines 169-173
	if g.MovePiece == 0 && g.State == 8 {
		// Check if this is black's max capture piece (assembly: CMP BMAXP / BEQ XRT)
		if g.MovePiece == g.BMAXP {
			return // Skip counting
		}
	}

	// Increment mobility counter (assembly: INC MOB,X)
	// Line 176: OVER INC MOB,X
	g.Mobility[stateIdx]++

	// Queens count as double mobility (assembly lines 177-179)
	// Line 177: CMP #$01 (is piece queen?)
	// Line 179: INC MOB,X (count again)
	if g.MovePiece == PieceQueen {
		g.Mobility[stateIdx]++
	}

	// Handle captures (assembly lines 181-198)
	// Line 181: NOQ BVC NOCAP (branch if V clear - no capture)
	if captureFlag {
		// Search BK array to find which piece is at SQUARE (target square)
		// Assembly lines 182-187: ELOOP
		capturedPieceIdx := NoPiece
		var capturedValue uint8 = 0

		for y := Piece(15); y != 0xFF; y-- { // Loop from 15 down to 0
			if g.BK[y] == g.MoveSquare {
				capturedPieceIdx = y
				capturedValue = POINTS[y]
				break
			}
		}

		if capturedPieceIdx != NoPiece {
			// Check if this is the best capture for this state
			// Assembly lines 188-192
			// Line 188: LDA POINTS,Y
			// Line 189: CMP MAXC,X (compare with current max)
			// Line 190: BCC LESS (branch if less)
			if capturedValue >= g.MaxCapture[stateIdx] {
				// Line 191: STY PCAP,X (save piece index)
				// Line 192: STA MAXC,X (save capture value)
				g.PieceCaptured[stateIdx] = capturedPieceIdx
				g.MaxCapture[stateIdx] = capturedValue
			}

			// Add to total capture count (assembly lines 194-198)
			// Lines 195-197: CLC / PHP / ADC CC,X / STA CC,X / PLP
			g.CaptureCount[stateIdx] += capturedValue
		}
	}

	// Special handling for STATE==4: ON4 routine (assembly lines 200-222)
	// This generates immediate reply moves for full position analysis
	// We'll implement this later when needed

	// STATE < 0: Call TREE for capture analysis
	// We'll implement this later for Phase 8-9 (search)
}

// STRATGY evaluates the current position and returns a score (0-255).
//
// This is the EXACT evaluation formula from the 1976 original (assembly line 641).
// The weights define MicroChess's chess "personality" and must be preserved exactly.
//
// Formula breakdown (assembly lines 641-701):
//
// Phase 1 (weight 0.25 - divide by 4 via two LSR instructions):
//
//	Start with $80 (128)
//	+ WMOB + WMAXC + WCC + WCAP1 + WCAP2
//	- PMAXC - PCC - BCAP0 - BCAP1 - BCAP2 - PMOB - BMOB
//	Divide by 2 (LSR)
//	Check for underflow, set to 0 if negative
//
// Phase 2 (weight 0.5 - divide by 2 via one LSR):
//
//	Result from Phase 1
//	+ $40 (64)
//	+ WMAXC + WCC
//	- BMAXC
//	Divide by 2 (LSR)
//
// Phase 3 (weight 1.0 - no division):
//
//	Result from Phase 2
//	+ $90 (144)
//	+ 4*WCAP0 + WCAP1
//	- 2*BMAXC - 2*BMCC - BCAP1
//
// Position bonus (+2):
//
//	If SQUARE is center ($33,$34,$22,$25 = d4,e4,d5,e5)
//	OR if piece moved from back rank (development bonus)
//
// Assembly reference:
//
//	STRATGY  CLC
//	         LDA #$80
//	         ADC WMOB       ; Add white mobility
//	         ... (full formula)
//	         CPX #$33       ; Center square bonus
//	         BEQ POSN
//	         ... (more center checks)
//
// Returns: Score 0-255 where higher is better for current side
//
// Assembly line: 641-701
func (g *GameState) STRATGY() uint8 {
	// Phase 1: Weight 0.25 (assembly lines 641-658)
	// Start with $80 (128) as neutral base
	acc := int16(0x80) // Use int16 to handle overflow/underflow

	// Add white advantages
	acc += int16(g.WMOB)
	acc += int16(g.WMAXC)
	acc += int16(g.WCC)
	acc += int16(g.WCAP1)
	acc += int16(g.WCAP2)

	// Subtract position/black advantages
	acc -= int16(g.PMAXC)
	acc -= int16(g.PCC)
	acc -= int16(g.BCAP0)
	acc -= int16(g.BCAP1)
	acc -= int16(g.BCAP2)
	acc -= int16(g.PMOB)
	acc -= int16(g.BMOB)

	// Underflow prevention (assembly lines 656-657: BCS POS / LDA #$00)
	if acc < 0 {
		acc = 0
	}

	// Divide by 2 (assembly line 658: LSR)
	acc = acc >> 1

	// Phase 2: Weight 0.5 (assembly lines 659-665)
	acc += 0x40 // Add $40 (64)
	acc += int16(g.WMAXC)
	acc += int16(g.WCC)
	acc -= int16(g.BMAXC)

	// Divide by 2 (assembly line 665: LSR)
	acc = acc >> 1

	// Phase 3: Weight 1.0 (assembly lines 666-678)
	acc += 0x90 // Add $90 (144)

	// 4 * WCAP0 (assembly lines 668-671: add WCAP0 four times)
	acc += int16(g.WCAP0)
	acc += int16(g.WCAP0)
	acc += int16(g.WCAP0)
	acc += int16(g.WCAP0)

	acc += int16(g.WCAP1)

	// Subtract black advantages
	acc -= int16(g.BMAXC)
	acc -= int16(g.BMAXC) // 2 * BMAXC
	acc -= int16(g.BMCC)
	acc -= int16(g.BMCC) // 2 * BMCC
	acc -= int16(g.BCAP1)

	// Position bonus (assembly lines 679-701)
	// NOTE: In the original assembly, STRATGY is called during move generation (from ON4)
	// and has access to PIECE and SQUARE for the current move being evaluated.
	// For the 'S' command, we're calling STRATGY after move generation completes,
	// so we don't have a specific move context. We'll skip the position bonus for now.
	// This will be properly implemented in Phase 8-9 when we integrate with the full search.
	//
	// The position bonus awards +2 for:
	// 1. Moves to center squares ($22,$25,$33,$34)
	// 2. Moving non-king pieces from back rank (development)
	//
	// For now, we skip this bonus as it requires move context.

	// Clamp to valid uint8 range
	if acc < 0 {
		return 0
	}
	if acc > 255 {
		return 255
	}

	return uint8(acc)
}

// ShowEvaluation displays position evaluation details and the board.
// This is a NEW command (not in original) - the 'S' command shows evaluation breakdown.
//
// It generates all moves with STATE=4 (full analysis), calculates the STRATGY score,
// and displays mobility, captures, and evaluation breakdown before showing the board.
func (g *GameState) ShowEvaluation() {
	// Save current state
	savedState := g.State
	savedMovePiece := g.MovePiece
	savedMoveSquare := g.MoveSquare
	savedMoveN := g.MoveN

	// Set STATE = 4 for full position analysis
	g.State = 4

	// Generate all moves to populate evaluation counters
	g.GNMZ()

	// Read counters from index 4 (where STATE=4 writes them)
	g.WMOB = g.Mobility[4]
	g.WMAXC = g.MaxCapture[4]
	g.WCC = g.CaptureCount[4]
	g.WMAXP = g.PieceCaptured[4]

	// For black counters, we need to generate black's moves
	// Save white's position, reverse, generate, reverse back
	savedWMOB, savedWMAXC, savedWCC, savedWMAXP := g.WMOB, g.WMAXC, g.WCC, g.WMAXP

	g.Reverse()
	g.State = 4
	g.GNMZ()
	g.BMOB = g.Mobility[4]
	g.BMAXC = g.MaxCapture[4]
	g.BMCC = g.CaptureCount[4]
	g.BMAXP = g.PieceCaptured[4]
	g.Reverse()

	// Restore white counters
	g.WMOB, g.WMAXC, g.WCC, g.WMAXP = savedWMOB, savedWMAXC, savedWCC, savedWMAXP

	// Evaluate the position
	score := g.STRATGY()

	// Display the board first
	g.Display()

	// Then display evaluation breakdown
	_, _ = fmt.Fprintf(g.out, "Position Evaluation: %x\r\n", score)
	//_, _ = fmt.Fprintf(g.out, "Mobility: W=%d B=%d\r\n", g.WMOB, g.BMOB)
	//_, _ = fmt.Fprintf(g.out, "Max Capture: W=%d B=%d\r\n", g.WMAXC, g.BMAXC)
	//_, _ = fmt.Fprintf(g.out, "Capture Count: W=%d B=%d\r\n", g.WCC, g.BMCC)

	// Restore state
	g.State = savedState
	g.MovePiece = savedMovePiece
	g.MoveSquare = savedMoveSquare
	g.MoveN = savedMoveN
}
