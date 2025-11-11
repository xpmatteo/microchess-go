# Running the Original 1976 MicroChess on go6502

This directory contains the go6502 emulator and the assembled MicroChess binary.

## Setup

### 1. Build go6502 (if not already built)

```bash
go build
```

This creates the `go6502` executable.

### 2. Assemble MicroChess

The MicroChess source has already been adapted from the original 6502 assembly to go6502's assembler syntax and assembled. To reassemble:

```bash
./go6502 -a microchess.asm
```

This produces:
- `microchess.bin` - The 6502 machine code binary (~1.4KB, matching the historic ~1.5KB size!)
- `microchess.map` - Symbol map for debugging

### 3. Run MicroChess

```bash
go run testrun.go iomem.go microchess.bin
```

This will:
1. Load the MicroChess binary at address $1000
2. Set the program counter to $1000 (start address)
3. Switch the terminal to raw mode for character-by-character input
4. Run the program until it hits a BRK instruction

**Important**: The terminal operates in raw mode - you do NOT need to press Enter after typing commands. Each character is sent immediately to the program. Press Ctrl+C to exit and restore normal terminal behavior.

## MicroChess Commands

The program runs directly without a debugger. When you see the chess board prompt (`?`), you can:

- `C` - Set up a new game (initialize board)
- `E` - Reverse/flip the board view
- `P` - Computer makes a move (play)
- Enter a move: Type FROM square (e.g., `12`) then TO square (e.g., `33`) - no Enter key needed
- `Q` - Quit to system ($FF00)

**Note**: In raw terminal mode, characters are NOT echoed to the screen. The program will display output but you won't see your keystrokes as you type them.

### Move Entry Format

Squares are numbered in hexadecimal from 00-77 using the 0x88 board representation:
```
   00 01 02 03 04 05 06 07
10 11 12 13 14 15 16 17
20 21 22 23 24 25 26 27
30 31 32 33 34 35 36 37
40 41 42 43 44 45 46 47
50 51 52 53 54 55 56 57
60 61 62 63 64 65 66 67
70 71 72 73 74 75 76 77
```

To move a piece from square 12 to square 33:
1. Press `1`, `2` (from square)
2. Press `3`, `3` (to square)
3. Press Enter

## Technical Notes

### Syntax Changes from Original

The original `Microchess6502.txt` used TASS assembler syntax. For go6502, the following changes were made:

1. `cpu 65c02` → `.ARCH 65c02`
2. `*=$1000` → `.ORG $1000`
3. `asc "text"` → `.DB "text"`
4. `db $XX` → `.DB $XX`
5. Character literals: `#"X"` → `#'X'` (double quotes to single quotes)
6. Second .ORG removed (not allowed after code) - used `.ALIGN` instead
7. Fixed case sensitivity issues - go6502 labels are case-sensitive

### Memory Map

- `$0050-$005F` - BOARD (16 piece positions)
- `$0060-$006F` - BK (opponent piece positions after REVERSE)
- `$00B0-$00FC` - Page zero variables (state, counters, etc.)
- `$1000-$1579` - Code
- `$1580-$158C` - Data tables (SETW, MOVEX, POINTS, OPNING)
- `$7F70-$7F73` - 6551 ACIA I/O ports (for serial communication)

### I/O Configuration

The original code expects a 6551 ACIA (Asynchronous Communications Interface Adapter) at $7F70-$7F73 for serial communication at 19.2Kbps. The go6502 emulator would need host I/O hooks configured to handle this properly.

## Comparing with Go Reimplementation

Once you have the original running in the emulator, you can:

1. Set up identical board positions in both versions
2. Make the same moves in both
3. Compare evaluation scores and move selections
4. Verify the Go port produces identical behavior

This provides a reference implementation to validate the Go port's correctness.

## Files

- `microchess.asm` - MicroChess source adapted for go6502 assembler
- `microchess.bin` - Assembled binary
- `microchess.map` - Symbol map
- `microchess.cmd` - go6502 script to load and start MicroChess
- `go6502` - The emulator executable

## References

- Original source: `../doc/Microchess6502.txt`
- Data structures: `../doc/DATA_STRUCTURES.md`
- Subroutines: `../doc/SUBROUTINES.md`
- Commands: `../doc/COMMANDS.md`
- Call graph: `../doc/CALL_GRAPH.md`
- go6502 documentation: https://github.com/beevik/go6502
