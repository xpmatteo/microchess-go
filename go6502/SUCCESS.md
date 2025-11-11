# 🎉 SUCCESS: Running 1976 MicroChess with Real I/O!

## Achievement

The original 1976 MicroChess assembly code is now **running with real console I/O** in the go6502 emulator!

## Quick Start

```bash
cd /Users/matteo/dojo/2025-11-07-1976-microchess-experiments/claude/go6502
go run testrun.go iomem.go microchess.bin
```

Commands:
- `C` - Set up the chess board
- `P` - Computer makes a move
- `E` - Reverse/flip board view
- `0-7` digits then Enter - Make a move (e.g., `6143` = e2-e4)
- Ctrl+C - Quit

## What We Built

### 1. Custom I/O Memory (`iomem.go`)
```go
type IoMemory struct {
    mem    *cpu.FlatMemory
    reader *bufio.Reader
}
```

- Implements the `Memory` interface from go6502/cpu
- Intercepts reads from $FFF1 (character input from stdin)
- Intercepts writes to $FFF0 (character output to stdout)
- All other memory operations pass through to FlatMemory

### 2. Simple Test Harness (`testrun.go`)
```go
func main() {
    mem := NewIoMemory()
    c := cpu.NewCPU(cpu.CMOS, mem)
    // Load program, set PC, run until BRK
}
```

- Loads 6502 binary at $1000
- Creates CPU with I/O-enabled memory
- Runs until BRK instruction
- Simple ~60 lines of Go code

### 3. Modified MicroChess Assembly
Changed three I/O routines:
```assembly
; Old ACIA version (complex):
syskin:  lda ACIASta
         and #$08
         beq syskin
         lda ACIADat
         rts

; New memory-mapped version (simple):
syskin:  lda GETCH    ; Read from $FFF1
         rts
```

Same for output - just write to $FFF0!

## Technical Details

### Memory-Mapped I/O Addresses
- `$FFF0` - Character output (write only)
- `$FFF1` - Character input (read only, blocking)

### Why This Works
The go6502 CPU emulator calls our Memory implementation's LoadByte/StoreByte methods for every memory access. We intercept the I/O addresses and route them to stdin/stdout.

### Changes to Original Code
Minimal - just replaced the 6551 ACIA driver routines:
- `Init_6551` - now just returns (no setup needed)
- `syskin` - reads from $FFF1
- `syschout` - writes to $FFF0

Total change: ~10 lines of assembly!

## Files Created

1. **iomem.go** - Custom Memory with I/O (73 lines)
2. **testrun.go** - Test harness (56 lines)
3. **microchess.asm** - Modified assembly (minimal changes)
4. **microchess.bin** - Assembled binary (1,389 bytes!)
5. **RUNNING_MICROCHESS.md** - Complete documentation
6. **SUCCESS.md** - This file

## Test Programs

Created along the way:
- `hello.asm` - Prints "Hello, MicroChess!" (output test)
- `testio.asm` - Echo input to output (I/O test)
- `realio.asm` - Real I/O demo (final test before MicroChess)

## What You Can Do Now

### 1. Play Chess Against the Original 1976 AI
```bash
go run testrun.go iomem.go microchess.bin
```

### 2. Validate Your Go Port
Feed identical inputs to both:
- Original 6502 running in emulator
- Your Go reimplementation

Compare outputs - they should match exactly!

### 3. Trace Execution
Use go6502's debugger commands to:
- Set breakpoints
- Examine memory
- Step through code
- Watch registers

### 4. Study the Algorithm
See the actual 1976 code executing in real-time!

## Performance

The 1.4KB binary includes:
- Complete chess move generation
- Position evaluation
- Multi-ply search with capture analysis
- Opening book
- Check/checkmate detection
- Terminal I/O routines
- Board display

All in less space than a single PNG icon file!

## Historical Significance

This is the **actual code** that ran on KIM-1 computers in 1976, now playable on modern systems. The algorithm is identical - we just adapted the I/O layer.

Peter Jennings wrote this in an era when:
- 1 MHz was fast
- 1KB of RAM was a lot
- Every byte mattered
- Clever coding was essential

Now we can study and appreciate that cleverness while it actually runs!

## Next Steps

With the original running, you can:

1. **Implement the Go port** - Phases 3-12 of PORTING_PLAN.md
2. **Validate correctness** - Compare Go vs 6502 execution
3. **Learn from the master** - Study the original's execution flow
4. **Preserve history** - Document this amazing piece of computing history

## Acknowledgments

- **Peter Jennings** (1976) - Original MicroChess author
- **Daryl Rictor** (2002) - Serial terminal adaptation
- **Bill Forster** (2005) - OCR error corrections
- **Brett Vickers** - go6502 emulator (https://github.com/beevik/go6502)
- **Captain Matt** (2025) - This preservation project

---

**Mission Accomplished!** 🚀♟️

The 1976 MicroChess lives again!
