# MicroChess Data Structures

This document describes the data structures used in the 1976 MicroChess 6502 assembly implementation.

## Memory Map Overview

MicroChess uses **page zero** (addresses $00-$FF) extensively for variables, as this provides:
- Faster access (2 cycles vs 4 cycles)
- Shorter instructions (2 bytes vs 3 bytes)
- Critical optimization for 1 MHz KIM-1 microcomputer

All game state is stored in page zero addresses $50-$FF.

---

## Board Representation

### BOARD Array ($50-$5F, 16 bytes)

The **BOARD** array stores the current square location for each of the 16 pieces on one side.

**Address Range**: `$50` to `$5F` (16 bytes)

**Piece Indexing**:
```
Index   Piece Type          Initial Square
-----   --------------      --------------
$00     King                $03 (white) / $73 (black)
$01     Queen               $04 (white) / $74 (black)
$02     Rook (queenside)    $00 (white) / $70 (black)
$03     Rook (kingside)     $07 (white) / $77 (black)
$04     Bishop (queenside)  $02 (white) / $72 (black)
$05     Bishop (kingside)   $05 (white) / $75 (black)
$06     Knight (queenside)  $01 (white) / $71 (black)
$07     Knight (kingside)   $06 (white) / $76 (black)
$08-$0F Pawns (8 pawns)     $10-$17 (white) / $60-$67 (black)
```

**Value Encoding**: Each byte contains a square number in **0x88 representation** (see below).

**Special Values**:
- `$CC` - Piece has been captured (square set to $CC in MOVE routine)

### BK Array ($60-$6F, 16 bytes)

The **BK** array serves dual purposes:

1. **During REVERSE operation**: Temporary storage for swapping board perspective
2. **During move generation**: Used to find captured pieces

**Address Range**: `$60` to `$6F` (16 bytes)

**Usage**: When `REVERSE` is called:
- BOARD ↔ BK arrays are swapped
- Each square coordinate is transformed: `new = $77 - old`
- Allows the engine to analyze opponent's replies from their perspective

---

## 0x88 Board Encoding

MicroChess uses the **0x88 board representation**, a clever encoding that makes edge detection efficient.

### Square Encoding

Each square is encoded as: `$RF` where:
- `R` = rank (row) from 0-7
- `F` = file (column) from 0-7

```
Rank 7:  $70 $71 $72 $73 $74 $75 $76 $77
Rank 6:  $60 $61 $62 $63 $64 $65 $66 $67
Rank 5:  $50 $51 $52 $53 $54 $55 $56 $57
Rank 4:  $40 $41 $42 $43 $44 $45 $46 $47
Rank 3:  $30 $31 $32 $33 $34 $35 $36 $37
Rank 2:  $20 $21 $22 $23 $24 $25 $26 $27
Rank 1:  $10 $11 $12 $13 $14 $15 $16 $17
Rank 0:  $00 $01 $02 $03 $04 $05 $06 $07
         a   b   c   d   e   f   g   h
```

### Edge Detection

The brilliant insight: **Any off-board square has bit $08 or $80 set**.

```assembly
AND #$88          ; Test if off board
BNE ILLEGAL       ; Non-zero = illegal square
```

This single instruction detects all board edges! Examples:
- $00 + $FF = $FF (bit $80 set - off left edge)
- $07 + $01 = $08 (bit $08 set - off right edge)
- $70 + $10 = $80 (bit $80 set - off top edge)
- $00 + $F0 = $F0 (both bits set - off bottom edge)

---

## Game State Variables

### Move Generation State

| Address | Name    | Description |
|---------|---------|-------------|
| `$B0`   | PIECE   | Current piece index being moved (0-15) |
| `$B1`   | SQUARE  | Current/target square being evaluated |
| `$B2`   | SP2     | Alternate stack pointer (starts at $C8) |
| `$B3`   | SP1     | Saved hardware stack pointer |
| `$B4`   | INCHEK  | $FF = safe, $00 = king can be captured |
| `$B5`   | STATE   | Analysis state machine value (see below) |
| `$B6`   | MOVEN   | Current move offset index into MOVEX table |
| `$B7`   | REV     | Board orientation: $00 = normal, $01 = reversed |

### STATE Variable Values

The **STATE** variable controls the analysis depth and behavior:

| Value  | Hex    | Meaning |
|--------|--------|---------|
| 12     | `$0C`  | Initial candidate move generation (COUNTS disabled) |
| 4      | `$04`  | Full move analysis with evaluation |
| 0      | `$00`  | Immediate reply move generation |
| 8      | `$08`  | Continuation move generation |
| -1 to -5 | `$FF-$FB` | Deep capture analysis (TREE recursion) |
| -7     | `$F9`  | Check detection mode (CHKCHK) |

---

## Evaluation Counters

The evaluation system uses arrays of counters indexed by STATE for different analysis depths.

### Counter Arrays ($DD-$F2)

These variables are organized as parallel arrays indexed by STATE:

**Base Counters** (used with STATE as index):
- **MOB** ($E3 base): Mobility count (number of legal moves)
- **MAXC** ($E4 base): Maximum capture value this position
- **CC** ($E5 base): Total capture count (sum of all capturable pieces)
- **PCAP** ($E6 base): Index of best piece that can be captured

**Named Instances** (for clarity in critical positions):

| Address | Name   | Purpose |
|---------|--------|---------|
| `$EB`   | WMOB   | White's mobility count |
| `$EC`   | WMAXC  | White's maximum capture value |
| `$ED`   | WCC    | White's total capture count |
| `$EE`   | WMAXP  | White's best capturable piece |
| `$E3`   | BMOB   | Black's mobility count |
| `$E4`   | BMAXC  | Black's maximum capture value |
| `$E5`   | BMCC   | Black's total capture count |
| `$E6`   | BMAXP  | Black's best capturable piece |
| `$EF`   | PMOB   | Position mobility |
| `$F0`   | PMAXC  | Position max capture |
| `$F1`   | PCC    | Position capture count |
| `$F2`   | PCP    | Position captured piece |

### Capture Depth Counters

| Address | Name   | Depth |
|---------|--------|-------|
| `$DD`   | WCAP0  | White's actual captures (depth 0) |
| `$DF`   | WCAP2  | White's captures at depth 2 |
| `$E1`   | WCAP1  | White's captures at depth 1 |
| `$E2`   | BCAP0  | Black's actual captures (depth 0) |
| `$DE`   | BCAP2  | Black's captures at depth 2 |
| `$E0`   | BCAP1  | Black's captures at depth 1 |
| `$E8`   | XMAXC  | Saved maximum capture value |

Note: These overlap with COUNT ($DE) as they're used in different contexts.

---

## Best Move Tracking

| Address | Name   | Description |
|---------|--------|-------------|
| `$FB`   | BESTP  | Best piece to move (piece index 0-15) |
| `$FA`   | BESTV  | Best move value/score |
| `$F9`   | BESTM  | Best destination square |

**Memory Sharing**: These addresses are reused for display:
- `$FB` = DIS1 (from square / piece display)
- `$FA` = DIS2 (current board square in display)
- `$F9` = DIS3 (to square / move entry)

This overlap saves precious page zero space. During move entry, the display variables overwrite the best move data, which is acceptable as best move is only needed during computer thinking.

---

## Move Stack Structure

The **dual stack mechanism** is one of MicroChess's cleverest features.

### Why Two Stacks?

- **Hardware Stack (SP)**: Used for JSR/RTS (subroutine calls)
- **Alternate Stack (SP2)**: Used for MOVE/UMOVE (game state)

This separation allows recursive move generation without corrupting the call stack.

### Stack Switching Code

```assembly
MOVE:
    TSX                ; Transfer SP to X
    STX SP1            ; Save hardware stack pointer
    LDX SP2            ; Load alternate stack pointer
    TXS                ; Activate alternate stack

    ... push move data ...

    TSX                ; Get current alternate SP
    STX SP2            ; Save it
    LDX SP1            ; Restore hardware SP
    TXS                ; Reactivate hardware stack
    RTS
```

### Move Record Format

Each MOVE pushes **5 bytes** onto the alternate stack (in this order):

1. **Target Square** - Where piece is moving to
2. **Captured Piece Index** - Which piece was captured (or $CC if none)
3. **Original BOARD Value** - Old location from BOARD array
4. **Piece Index** - Which piece is moving (0-15)
5. **MOVEN** - Move offset index (for debugging/replay)

UMOVE pops these in **reverse order** to restore the position.

---

## Opening Book

| Address | Name   | Description |
|---------|--------|-------------|
| `$DC`   | OMOVE  | Current index into opening book ($FF = out of book) |

### Opening Book Format (OPNING table, line 881)

The opening book is stored as a sequence of move pairs (from_square, to_square):

```assembly
OPNING:
    db $99, $25, $0B, $25, $01, $00, $33, $25
    db $07, $36, $34, $0D, $34, $34, $0E, $52
    db $25, $0D, $45, $35, $04, $55, $22, $06
    db $43, $33, $0F, $CC
```

- Decremented by 3 after each move (2 bytes for move + 1 for index)
- Terminated by $CC
- If opponent's move doesn't match expected move, OMOVE set to $FF (exit book)

---

## Constant Tables

### MOVEX - Direction Offsets ($1580, line 875)

Movement direction offsets for move generation:

```
Index   Offset   Direction
-----   ------   ---------
0       $00      (unused/placeholder)
1       $F0      Up 1 rank (-16)
2       $FF      Left 1 file (-1)
3       $01      Right 1 file (+1)
4       $10      Down 1 rank (+16)
5       $11      Down-right diagonal (+17)
6       $0F      Down-left diagonal (+15)
7       $EF      Up-left diagonal (-17)
8       $F1      Up-right diagonal (-15)
9-16    Various  Knight L-shaped moves
```

**Usage**: `newSquare = SQUARE + MOVEX[MOVEN]`

### POINTS - Piece Values (line 878)

Material values for captured pieces:

```
Index   Value   Piece Type
-----   -----   ----------
0       $0B     King (11 - special, shouldn't be capturable)
1       $0A     Queen (10)
2-3     $06     Rooks (6 each)
4-5     $04     Bishops (4 each)
6-7     $04     Knights (4 each)
8-15    $02     Pawns (2 each)
```

These values are used in COUNTS to evaluate captures and in STRATGY for position evaluation.

### SETW - Initial Board Setup (line 870)

Initial piece positions (32 bytes for both sides):

```assembly
SETW:
    db $03, $04, $00, $07, $02, $05, $01, $06  ; White pieces
    db $10, $17, $11, $16, $12, $15, $14, $13  ; White pawns
    db $73, $74, $70, $77, $72, $75, $71, $76  ; Black pieces
    db $60, $67, $61, $66, $62, $65, $64, $63  ; Black pawns
```

Loaded into BOARD array when 'C' (setup) command is issued.

---

## Display Variables

| Address | Name   | Description |
|---------|--------|-------------|
| `$FB`   | DIS1   | From square / piece to display |
| `$FA`   | DIS2   | Current square in display loop |
| `$F9`   | DIS3   | To square / accumulated move input |
| `$F3`   | OLDKY  | Previous key (for debouncing - unused in serial version) |

**Note**: DIS1/DIS2/DIS3 share memory with BESTP/BESTV/BESTM (see above).

---

## Memory Layout Summary

```
Page Zero Memory Map:

$50-$5F  BOARD[16]      Piece locations (white's perspective)
$60-$6F  BK[16]         Alternate board / capture search
$B0      PIECE          Current piece being moved
$B1      SQUARE         Current square being evaluated
$B2      SP2            Alternate stack pointer
$B3      SP1            Saved hardware stack pointer
$B4      INCHEK         Check detection flag
$B5      STATE          Analysis state machine
$B6      MOVEN          Move offset table index
$B7      REV            Board reversed flag
$DC      OMOVE          Opening book index
$DD-$E2  Capture counters (WCAP/BCAP at various depths)
$E3-$E6  Black mobility counters (BMOB, BMAXC, BMCC, BMAXP)
$E8      XMAXC          Saved max capture
$EB-$EE  White mobility counters (WMOB, WMAXC, WCC, WMAXP)
$EF-$F2  Position counters (PMOB, PMAXC, PCC, PCP)
$F3      OLDKY          Old key (debounce)
$F9-$FB  BESTM/DIS3, BESTV/DIS2, BESTP/DIS1 (overlapping)
$FC      temp           Temporary variable

Code/Data in RAM:

$1000-$15FF  Program code
$1580        SETW initial board setup
$1589        MOVEX direction offset table
$1591        POINTS piece value table
$1599        OPNING opening book data
```

---

## Key Insights

1. **0x88 Encoding**: Allows one-instruction edge detection - brilliant for 6502
2. **Dual Stack**: Separates call stack from game state - enables clean recursion
3. **Page Zero**: All variables in fast memory - critical for performance
4. **Memory Reuse**: DIS/BEST overlap saves 3 precious bytes
5. **STATE Machine**: Single variable controls complex analysis flow
6. **Piece Indexing**: Simple array indices (0-15) identify pieces efficiently

This data structure design demonstrates remarkable ingenuity given the extreme constraints of 1976-era hardware (1 MHz CPU, ~1KB usable RAM).
