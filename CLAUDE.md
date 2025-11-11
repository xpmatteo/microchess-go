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


## Running the Original 6502 Code

The original 1976 MicroChess assembly code can be run in the go6502 emulator with real I/O!

Example:

```bash
printf 'CQ' | make play-6502
```

This way we can see what is output by the original program in response to our input

**See**: `go6502/RUNNING_MICROCHESS.md` for complete details

This allows direct comparison between the original 6502 code and the Go port!

## Testing Philosophy

### Test Types

**Unit Tests**:
- Each major routine has corresponding unit tests in `pkg/*/` directories
- Test individual functions (board representation, coordinate transformations, etc.)
- Use `github.com/stretchr/testify` for assertions

**Acceptance Tests** (`acceptance/`):
- YAML-based test fixtures in `acceptance/testdata/*.yaml`
- Define command sequences and expected display output
- Test complete user workflows (setup, reverse, etc.)
- Compare Go port output to expected board states
- Line endings are normalized (`\r\n` → `\n`) for cross-platform compatibility

**6502 Comparison Tests** (`acceptance/emulator_test.go`):
- Feed identical inputs to both Go port and original 6502 emulator
- Parse and compare board displays from both implementations
- Verify LED display values match (DIS1, DIS2, DIS3)
- Ensure behavior is byte-for-byte equivalent to 1976 original

### Test Guidelines

- **Always run `golangci-lint run`** before concluding any implementation step
- **Use `t.Skip()`** for tests of unimplemented functionality (e.g., move input)
  - Tests should compile but skip execution with clear message
  - Example: `t.Skip("Move input not yet implemented - scheduled for future phase")`
- **YAML test format**: See `acceptance/testdata/setup-and-quit.yaml` for structure
- **Piped input testing**: Use `printf 'CPQ' | go run ./cmd/microchess` for quick validation
- **Note**: `timeout` command is not installed; do not try to use it