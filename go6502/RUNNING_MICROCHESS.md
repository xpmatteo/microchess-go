# Running 1976 MicroChess with Real I/O! ✅

## Success!

The original 1976 MicroChess is now running in the go6502 emulator with **real console I/O**!

## How It Works

We created a custom Memory implementation (`iomem.go`) that intercepts memory reads/writes at specific addresses:

- **$FFF0** - Character output (writes go to stdout)
- **$FFF1** - Character input (reads from stdin, blocking)

This is the idiomatic way to implement I/O in 6502 emulators.

## Running MicroChess

```bash
cd /Users/matteo/dojo/2025-11-07-1976-microchess-experiments/claude/go6502

# Run MicroChess interactively
go run testrun.go iomem.go microchess.bin
```

## Playing Chess

Once running, you'll see the chess board and a `?` prompt. Commands:

### Setup Commands
- `C` - Initialize/setup the chess board with pieces
- `E` - Reverse/flip the board view
- `P` - Computer makes a move (think and play)

### Making Moves
Squares are numbered 00-77 in hex (0x88 board representation):
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

To move e.g., pawn from 61 to 51:
1. Type `6`, `1`, `5`, `1`, then Enter

### Pieces Display
- `WR`/`BR` - White/Black Rook
- `WN`/`BN` - White/Black Knight
- `WB`/`BB` - White/Black Bishop
- `WK`/`BK` - White/Black King
- `WQ`/`BQ` - White/Black Queen
- `WP`/`BP` - White/Black Pawn

## Files Created

1. **iomem.go** - Custom Memory implementation with I/O support
2. **testrun.go** - Simple test harness to run 6502 programs with I/O
3. **microchess.asm** - Modified to use $FFF0/$FFF1 instead of ACIA serial
4. **realio.asm** - Test program demonstrating the I/O mechanism

## Example Session

```
$ go run testrun.go iomem.go microchess.bin
Loaded 1389 bytes at $1000
Running...
---

MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|  |**|  |**|  |**|  |**|00
...
?C
...
|WR|WN|WB|WK|WQ|WB|WN|WR|00
...
?
```

## Technical Details

### Memory-Mapped I/O Implementation

The `IoMemory` type in `iomem.go`:
- Wraps the standard `FlatMemory`
- Intercepts `LoadByte($FFF1)` to read from stdin
- Intercepts `StoreByte($FFF0, ch)` to write to stdout
- All other memory operations pass through normally

### MicroChess Modifications

Changed the I/O routines (lines 817-830):
- `syskin`: reads from `GETCH` ($FFF1)
- `syschout`: writes to `PUTCH` ($FFF0)
- `Init_6551`: no-op (no initialization needed)

This is a minimal change - just 3 simple routines instead of the original ACIA driver!

## Comparing with Go Port

Now you can:
1. Run the original 6502 code
2. Run your Go reimplementation
3. Feed identical inputs to both
4. Compare board states, evaluations, and move selections
5. Validate your Go port produces identical behavior!

## Next Steps

To make it more user-friendly, you could:
- Add algebraic notation parsing
- Create a nicer board display
- Add move validation hints
- Implement a simple GUI

But the core 1976 MicroChess algorithm is now running exactly as Peter Jennings wrote it!
