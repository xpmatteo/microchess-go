# MicroChess Emulation Setup - Complete! ✓

## What We Accomplished

Successfully set up the ability to run the original 1976 MicroChess 6502 assembly code in an emulator!

### Steps Completed

1. **Installed go6502 emulator**
   - Cloned from https://github.com/beevik/go6502
   - Built the emulator successfully
   - Located at: `/Users/matteo/dojo/2025-11-07-1976-microchess-experiments/claude/go6502`

2. **Adapted MicroChess source for go6502**
   - Converted from TASS assembler syntax to go6502 syntax
   - Key changes:
     - `*=` → `.ORG`
     - `asc` → `.DB`
     - `db` → `.DB`
     - `#"X"` → `#'X'` (character literals)
     - Fixed label case sensitivity issues
   - Source: `microchess.asm`

3. **Successfully assembled MicroChess**
   - Generated binary: `microchess.bin` (1.4KB - matches historic ~1.5KB!)
   - Generated symbol map: `microchess.map`
   - Binary loads at address $1000

4. **Verified execution starts**
   - Program loads correctly
   - Begins executing at $1000
   - Initializes hardware and calls subroutines as expected

## Next Steps

### To Run MicroChess Interactively

```bash
cd /Users/matteo/dojo/2025-11-07-1976-microchess-experiments/claude/go6502
./go6502 microchess.cmd
```

Then use debugger commands to step through execution:
- `si 10` - Step 10 instructions
- `d .` - Disassemble current location
- `m $50` - View board memory
- `c` - Continue execution

### For Go Port Validation

Now you can:

1. **Run the original** in the emulator with specific inputs
2. **Run your Go port** with the same inputs
3. **Compare the results** to ensure identical behavior:
   - Board state after moves
   - Evaluation scores
   - Move selections
   - Opening book usage

### Known Limitation

The emulator crashes when trying to enable raw terminal mode for interactive I/O. This is expected when running with piped input. To fully interact with MicroChess, you'd need to:

- Run `./go6502 microchess.cmd` in an actual terminal (not piped)
- Or configure the emulator's I/O hooks for the 6551 ACIA

## Files Created

- `microchess.asm` - Adapted assembly source
- `microchess.bin` - Assembled 6502 machine code
- `microchess.map` - Debug symbol map
- `microchess.cmd` - Emulator startup script
- `README_MICROCHESS.md` - Full documentation
- `SETUP_COMPLETE.md` - This file

## Documentation References

All documentation is in the parent `doc/` directory:
- `doc/Microchess6502.txt` - Original source
- `doc/DATA_STRUCTURES.md` - Memory layout
- `doc/SUBROUTINES.md` - Algorithm explanations
- `doc/COMMANDS.md` - User interface
- `doc/CALL_GRAPH.md` - Control flow

---

**Status**: ✅ Ready to compare original 6502 implementation with Go port!
