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

- **doc/microchess-manual.txt** - Original program manual from 1976
  - Explains available command
  - Data structures and variables
  - Algorithm


### Go Port

The Go port will be organized as:

```
cmd/microchess/main.go     - CLI entry point
pkg/board/                 - Board representation
pkg/eval/                  - Position evaluation
pkg/microchess/            - Core engine
```

## Input Handling

The Go port implements character-by-character input matching the original 1976 serial terminal behavior:

### Raw Terminal Mode (Interactive Use)

When running interactively (`go run ./cmd/microchess`), the program switches stdin to raw mode using `golang.org/x/term.MakeRaw()`. This provides:

- **Character-by-character input**: Each keypress is processed immediately without waiting for Enter
- **Character echo**: Typed characters are echoed to the terminal (matching original SYSCHOUT behavior)
- **Commands execute instantly**: Pressing `C` immediately sets up the board, `E` reverses, `P` prints, `Q` quits
- **Carriage return + line feed**: Output uses `\r\n` for proper line positioning in raw mode

This matches the original assembly's **KIN routine** (line 812) which reads one character at a time from the serial port using a polling loop.

### Piped Input Mode (Testing)

When input is piped (`printf 'CPQ' | ./microchess`), the program detects this using `term.IsTerminal()` and:

- **Reads character-by-character** but skips raw mode setup
- **No character echo**: Characters aren't echoed to avoid cluttering test output
- **Handles EOF gracefully**: Exits cleanly when pipe closes
- **Enables automated testing**: Allows scripted input for acceptance tests and comparisons

### Move Entry Format

**CRITICAL**: The 2002 Daryl Rictor serial terminal version uses a different move entry format than the original 1976 KIM-1 version:

- **Move entry**: Enter 4 digits (FROM square, TO square in octal) followed by CARRIAGE RETURN (`\r`)
  - Example: `0122\r` moves the knight at position 01 to position 22 (b1 to c3)
  - Example: `1333\r` moves the piece at position 13 to position 33
  - The display shows the piece code and coordinates as you type: `06 01 22` (06 = knight) or `0F 13 33` (0F = pawn)
  - After the move executes, the FROM square shows `FF` (empty) and the piece appears at the TO square
- **USE CARRIAGE RETURN (`\r`), NOT NEWLINE (`\n`)** - This is crucial for move execution
- **DO NOT use the `F` key** - The original 1976 manual describes pressing `F` to execute moves, but this does NOT apply to our serial terminal version
- **Board coordinates**: Octal notation where first digit is rank (0-7 from white's side), second digit is file (0-7 from left)
- **Single-letter commands use any line ending**: Commands like `C`, `E`, `P`, `Q` work with either `\r` or `\n`

### Implementation Details

- **Auto-detection**: Uses `term.IsTerminal(os.Stdin.Fd())` to detect terminal vs pipe
- **Single-byte reads**: `os.Stdin.Read(buf)` with 1-byte buffer, just like original ACIA reads
- **Uppercase conversion**: Characters are converted to uppercase (original: `AND #$4F`)
- **Location**: See `cmd/microchess/main.go` lines 14-74

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

### Running the Go Port

```bash
# Build and run interactively (character-by-character input in raw terminal mode)
go run ./cmd/microchess

# Build binary
go build -o microchess ./cmd/microchess

# Test with piped input (for automated testing)
printf 'CPQ' | go run ./cmd/microchess
printf 'CEPEQ' | ./microchess
```

### Testing

```bash
# Run all tests
go test ./...

# Run acceptance tests only
go test ./acceptance/...

# Run with verbose output
go test ./... -v

# Run linter
golangci-lint run
```

### Comparing with Original

```bash
# Run original 6502 code with same input
printf 'CQ' | make play-6502

# Compare outputs side-by-side
printf 'CEQ' | go run ./cmd/microchess > go-output.txt
printf 'CEQ' | make play-6502 > 6502-output.txt
diff go-output.txt 6502-output.txt
```

## Testing Philosophy

### Test Types

**Unit Tests**:
- Each major routine has corresponding unit tests in `pkg/*/` directories
- Test individual functions (board representation, coordinate transformations, etc.)
- Use `github.com/stretchr/testify` for assertions

**Acceptance Tests** (`acceptance/`):
- ALWAYS create acceptance tests before implementing new features
- ALWAYS create acceptance tests by observing the 6502 behaviour (see below)
- YAML-based test fixtures in `acceptance/testdata/*.yaml`
- Define command sequences and expected display output
- Test complete user workflows (setup, reverse, etc.)
- Compare Go port output to expected board states
- Line endings are normalized (`\r\n` → `\n`) for cross-platform compatibility
- Runs the same input against both 6502 code and Go code


### Test Guidelines

- **Always run `golangci-lint run`** before concluding any implementation step
- **Piped input testing**: Use `printf 'CPQ' | go run ./cmd/microchess` for quick validation
- **Note**: `timeout` command is not installed; do not try to use it

### Creating Acceptance Tests

**Workflow**: Develop acceptance tests by first observing the original 1976 program's behavior:

1. **Run the original with test input**:
   ```bash
   printf 'CEQ' | make play-6502 > original-output.txt
   ```

2. **Extract the board displays** from the output to understand expected behavior

3. **Create YAML test fixture** in `acceptance/testdata/` using the original's output:
   ```yaml
   name: "Setup, Reverse, and Quit Sequence"
   description: "Tests double reverse returning board to original orientation"
   steps:
     - command: "DISPLAY"
       should_continue: true
       expected_display: |
         [paste board display from original output]
     - command: "C"
       should_continue: true
       expected_display: |
         [paste board display after C command]
   ```

4. **Run the Go port** with same input and verify it matches:
   ```bash
   printf 'CEQ' | go run ./cmd/microchess
   go test ./acceptance/... -v
   ```

This ensures the Go port's behavior is validated against the actual 1976 original, not against assumptions.