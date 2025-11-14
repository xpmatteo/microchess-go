# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains the 1976 MicroChess game by Peter Jennings and is working toward a Go port that makes this historic program easier to understand while preserving its exact chess logic.

**Goal**: Create a faithful Go translation using idiomatic Go structures (structs, methods) to clarify the assembly's dense logic without changing the chess algorithm's behavior.

## Repository Structure

### Original Assembly Code

- **doc/Microchess6502.txt** - Complete 6502 assembly source code for MicroChess
- **doc/6502-instruction-set.md** - Reference for 6502 CPU instructions
- **doc/microchess-manual.txt** - Original program manual from 1976


### Go Port

The Go port is organized as:

```
acceptance                 - Acceptance tests
cmd/microchess/main.go     - CLI entry point
pkg/board/                 - Board representation
pkg/eval/                  - Position evaluation
pkg/microchess/            - Core engine
```

### Move Entry Format

- **Move entry**: Enter 4 digits (FROM square, TO square in octal) followed by CARRIAGE RETURN (`\r`)
  - Example: `0122\r` moves the knight at position 01 to position 22 (b1 to c3)
  - Example: `1333\r` moves the piece at position 13 to position 33
  - The display shows the piece code and coordinates as you type: `06 01 22` (06 = knight) or `0F 13 33` (0F = pawn)
  - After the move executes, the FROM square shows `FF` (empty) and the piece appears at the TO square
- **USE CARRIAGE RETURN (`\r`), NOT NEWLINE (`\n`)** 
- **DO NOT use the `F` key** - The original 1976 manual describes pressing `F` to execute moves, but this does NOT apply to our serial terminal version
- **Board coordinates**: Octal notation where first digit is rank (0-7 from top of the board), second digit is file (0-7 from left)
- **Single-letter commands use any line ending**: Commands like `C`, `E`, `P`, `Q` work with either `\r` or `\n`

## Board orientation

The AI player pieces are at the top of the board; the human player pieces are at the bottom of the board.  
The default setting has white at the top, and black at the bottom, and this can be reversed with the 'E' command. 
After one 'E', the AI player will have the black pieces.

## Avoid standard chess algebraic notation

Prefer octal notation in comments and documentation, eg 00 instead of f1; octal coordinates are invariant after reverse ('E').

## Development Commands

### Running the Go Port

```bash
# Build and run interactively (character-by-character input in raw terminal mode)
make run

# Test with piped input (for automated testing)
printf 'CPQ' | make run
printf 'CEPEQ' | make run
```

### Testing

```bash
# Run all tests
go test ./...

# Run acceptance tests only
go test ./acceptance/...

# Run linter
make lint

# Run gofmt
make fmt
```

### Comparing with Original

```bash
# Run original 6502 code
printf 'CQ' | make play-6502

# Compare outputs side-by-side
printf 'CEQ' | make run > go-output.txt
printf 'CEQ' | make play-6502 > 6502-output.txt
diff go-output.txt 6502-output.txt
```

## Tests

**Unit Tests**:
- Each major routine has corresponding unit tests in `pkg/*/` directories
- Use `github.com/stretchr/testify` for assertions

**Acceptance Tests** (`acceptance/`):
- ALWAYS create acceptance tests before implementing new features
- ALWAYS create acceptance tests by observing the 6502 behaviour first (see below)
- YAML-based test fixtures in `acceptance/testdata/XX-subdir/*.yaml`
- Check for example `acceptance/testdata/02-moving/move-knight-on-black-square.yaml`
- The AT run by default the same input against both 6502 code and Go code, expecting exactly the same output
- Use `skip_6502:` for extension features that the original does not support 


## Notes

- **Always run `make fmt test lint`** before concluding any implementation step
- `timeout` command is not installed; do not try to use it

