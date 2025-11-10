# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains the 1976 MicroChess game by Peter Jennings and is working toward a Go port that makes this historic program easier to understand while preserving its exact chess logic.

**Goal**: Create a faithful Go translation using idiomatic Go structures (structs, methods) to clarify the assembly's dense logic without changing the chess algorithm's behavior.

## Repository Structure

### Original Assembly Code

- **doc/Microchess6502.txt** - Complete 6502 assembly source code for MicroChess
  - Original by Peter Jennings (1976)
  - Modified by Daryl Rictor (2002) to work over serial terminal
  - Updated by Bill Forster (2005) with OCR error corrections
  - ~1.5KB of code implementing complete chess AI
  - Uses dual stack mechanism, 0x88 board representation, state machine

- **doc/6502-instruction-set.md** - Reference for 6502 CPU instructions
  - Complete instruction set with opcodes, addressing modes, cycle counts
  - Useful for understanding assembly idioms

### Documentation (Phase 1 Complete)

These documents explain the original program's workings:

- **doc/DATA_STRUCTURES.md** - Complete data structure reference
  - Page zero variable map ($50-$FF)
  - 0x88 board encoding explanation
  - Move stack structure
  - Evaluation counters organization
  - Opening book format

- **doc/SUBROUTINES.md** - All major routines documented
  - Move generation (GNM, CMOVE, piece handlers)
  - Move execution (MOVE, UMOVE, REVERSE)
  - Evaluation (COUNTS, TREE, STRATGY, CKMATE)
  - Search (JANUS state machine, CHKCHK)
  - Computer player (GO)
  - Pseudocode algorithms

- **doc/COMMANDS.md** - User interface reference
  - All available commands (C, E, P, Enter, Q, 0-7)
  - Move entry format
  - Display format explanation
  - Input masking details

- **doc/CALL_GRAPH.md** - Visual flow diagrams
  - Main program flow
  - Move generation call graph
  - STATE machine transitions
  - Stack usage diagrams
  - Evaluation flow

### Go Port (Future Phases)

The Go port will be organized as:

```
cmd/microchess/main.go     - CLI entry point
pkg/board/                 - Board representation
pkg/eval/                  - Position evaluation
pkg/microchess/            - Core engine
```

## Key Historical Context

MicroChess demonstrates remarkable code density:

- **1.5KB total** including all code and data
- **Dual stack mechanism** - separates call stack (SP) from game state stack (SP2)
- **0x88 board representation** - one-instruction edge detection
- **STATE machine** - single variable controls search depth and analysis type
- **Page zero optimization** - all variables in fast memory ($50-$FF)
- **Minimax with capture analysis** - sophisticated for 1976

## Porting Strategy

**What to Preserve**:
1. 0x88 board encoding (elegant and proven)
2. Exact evaluation formula weights
3. STATE machine values and flow
4. Piece indexing (0-15 for white)
5. Move offset table (MOVEX)
6. Opening book moves

**What to Modernize**:
1. Dual stack → Go slice for move history
2. CPU flags → explicit return struct
3. Global state → methods on GameState struct
4. Add comprehensive comments explaining "why"
5. Unit tests verifying exact behavior

## Important Assembly Idioms

When reading the original code, understand these patterns:

- **AND #$88** - detects off-board in 0x88 representation
- **TSX/TXS** - stack switching for MOVE/UMOVE
- **N/V/C flags** - CMOVE returns move legality via flags
- **STATE values** - 12, 4, 0, 8, -1 to -5, -7 control analysis
- **$77 - coord** - REVERSE coordinate transformation
- **Memory overlap** - DIS1/2/3 share space with BESTP/V/M

## Development Commands

(To be added as Go port progresses)

## Running the Original 6502 Code

The original 1976 MicroChess assembly code can be run in the go6502 emulator with real I/O!

**Location**: `go6502/` directory

**Quick Start**:
```bash
cd go6502
go run testrun.go iomem.go microchess.bin
```

Then type `C` to set up the board, `P` for computer move, or enter moves manually.

**How It Works**:
- `iomem.go` - Custom Memory implementation with console I/O at $FFF0 (output) and $FFF1 (input)
- `testrun.go` - Simple harness to run 6502 programs
- `microchess.asm` - Modified to use memory-mapped I/O instead of ACIA serial port

**See**: `go6502/RUNNING_MICROCHESS.md` for complete details

This allows direct comparison between the original 6502 code and the Go port!

## Testing Philosophy

- Unit tests for each major routine
- Verify move generation matches assembly
- Validate evaluation scores
- Test opening book sequences
- **Compare game play to original running in go6502 emulator**
- Feed identical inputs to both implementations and verify identical outputs
- use stretchr/testify for assertions