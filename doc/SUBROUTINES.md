# MicroChess Subroutines

This document describes all major subroutines in the 1976 MicroChess 6502 assembly implementation, organized by functional area.

## Table of Contents

1. [Control Flow](#control-flow)
2. [Move Generation](#move-generation)
3. [Move Execution](#move-execution)
4. [Position Evaluation](#position-evaluation)
5. [Search & Analysis](#search--analysis)
6. [Computer Player](#computer-player)
7. [Utility Routines](#utility-routines)
8. [Display & I/O](#display--io)

---

## Control Flow

### CHESS (line 100)

**Purpose**: Main initialization and game loop

**Flow**:
1. Clear decimal mode (CLD)
2. Initialize hardware stack (X = $FF, TXS)
3. Initialize alternate stack (SP2 = $C8)
4. Loop:
   - Call OUT (display board and get input)
   - Dispatch based on input key
   - Repeat forever

**Commands Dispatched**:
- 'C' ($43) → WHSET (setup board)
- 'E' ($45) → REVERSE (flip board)
- 'P' ($40) → GO (computer plays)
- Enter ($0D) → MOVE (execute player move)
- 'Q' ($41) → DONE (exit to system)
- 0-7 → INPUT (enter move square)

**Special Behavior**:
- After setup (C), displays "CCC" and sets OMOVE = $1B (start of opening book)
- After reverse (E), displays "EEE" and toggles REV flag
- After computer move (P), displays result across all three display positions

**Assembly Reference**: Lines 100-153

---

### INPUT (line 262)

**Purpose**: Process numeric input (0-7) for move entry

**Input**:
- Accumulator = key pressed (validated to be $00-$07)

**Algorithm**:
1. Compare with $08 (BCS → ERROR if >= 8)
2. Call DISMV to rotate key into move accumulator
3. Search BOARD array for piece at DIS2 (from square)
4. Store piece index in DIS1 and PIECE
5. Jump to CHESS to continue game loop

**Side Effects**:
- Updates DIS1, DIS2, DIS3 (move entry registers)
- Updates PIECE (for subsequent MOVE command)

**Assembly Reference**: Lines 262-273

---

### DISMV (line 625)

**Purpose**: Rotate input key into move display registers

**Input**:
- Accumulator = digit (0-7)

**Algorithm**:
```
Shift DIS3 left 4 bits
Shift DIS2 left 4 bits (with carry from DIS3)
OR accumulator with DIS3
Store in DIS3 and SQUARE
```

**Effect**: Builds up a move from two digits:
- First digit → high nibble of DIS3
- Second digit → low nibble of DIS3
- Result is a square in 0x88 format

**Example**:
```
Input: 3, 4
After first '3':  DIS3 = $30
After second '4': DIS3 = $34 (square d4)
```

**Assembly Reference**: Lines 625-633

---

## Move Generation

### GNM (line 286)

**Purpose**: Generate all moves for current side

**Algorithm**:
```
Set PIECE = $10 (start at piece 16, will decrement to 15)
Loop:
    Decrement PIECE
    If PIECE < 0, return (all pieces done)

    RESET (load piece's current square)
    Set MOVEN based on piece type

    Dispatch to piece-specific handler:
        0      → KING
        1      → QUEEN
        2-3    → ROOK
        4-5    → BISHOP
        6-7    → KNIGHT
        8-15   → PAWN
```

**Piece Type Detection** (lines 296-331):
- CPY #$08, BPL → PAWN
- CPY #$06, BPL → KNIGHT
- CPY #$04, BPL → BISHOP
- CPY #$01, BEQ → QUEEN
- BPL → ROOK
- Else → KING

**Assembly Reference**: Lines 286-352

---

### GNMZ (line 280)

**Purpose**: Clear counters and call GNM

**Algorithm**:
1. Load X = $10 (16 counters)
2. Clear loop:
   - Store $00 in COUNT,X (clears $DE-$EE)
   - Decrement X
   - Loop until X < 0
3. Call GNM

**Cleared Variables**:
All evaluation counters from $DE to $EE (17 bytes)

**Assembly Reference**: Lines 280-285

---

### CMOVE (line 407)

**Purpose**: Calculate and validate a single move

**Input**:
- SQUARE = current position
- MOVEN = index into MOVEX table

**Output** (via flags):
- **N flag set** = Illegal (off board or own piece)
- **V flag set** = Capture (opponent piece on target square)
- **C flag set** = Move causes check (if check checking enabled)

**Algorithm**:
```
1. newSquare = SQUARE + MOVEX[MOVEN]
2. If newSquare & $88 != 0:
     Return ILLEGAL (N=1, C=0, V=0)
3. Search BOARD array for piece at newSquare:
     If found in indices 0-15 (own pieces):
         Return ILLEGAL
     If found in indices 16-31 (opponent):
         Set V flag (capture)
         Continue to step 4
     If not found:
         Clear V flag (no capture)
4. If STATE in range 0-7:
     Call CHKCHK (check if move is legal)
     Return with C flag indicating check
5. Return LEGAL (N=0)
```

**Special Cases**:
- CHKCHK only called for STATE 0-7 (expensive operation)
- $88 boundary test catches all edge cases in one instruction

**Assembly Reference**: Lines 407-469

---

### SNGMV (line 357)

**Purpose**: Generate single-step moves (King, Knight)

**Algorithm**:
```
1. Call CMOVE to calculate move
2. If legal (N flag clear):
     Call JANUS to evaluate
3. Call RESET to restore SQUARE
4. Decrement MOVEN
5. Return
```

**Used By**:
- KING (8 directions, MOVEN $08 down to $01)
- KNIGHT (8 L-shapes, MOVEN $10 down to $09)

**Assembly Reference**: Lines 357-362

---

### LINE (line 367)

**Purpose**: Generate sliding moves (Queen, Rook, Bishop)

**Algorithm**:
```
Loop:
    Call CMOVE to calculate move
    If carry set (check) OR overflow clear (no capture):
        Continue in same direction
    If move illegal (N flag set):
        Break loop
    Call JANUS to evaluate
    If overflow set (capture):
        Break loop (can't slide past capture)
End loop

Call RESET to restore SQUARE
Decrement MOVEN (next direction)
Return
```

**Key Insight**: Uses V flag (capture) to stop sliding

**Used By**:
- QUEEN (8 directions, MOVEN $08 down to $01)
- ROOK (4 directions, MOVEN $04 down to $01)
- BISHOP (4 directions, MOVEN $08 down to $05)

**Assembly Reference**: Lines 367-377

---

### Piece-Specific Handlers

#### KING (line 306)

```
Loop (8 directions):
    SNGMV (single step)
    If MOVEN != 0, continue
Return to NEWP (next piece)
```

#### QUEEN (line 309)

```
Loop (8 directions):
    LINE (sliding moves)
    If MOVEN != 0, continue
Return to NEWP
```

#### ROOK (line 313)

```
Set MOVEN = $04 (4 orthogonal directions)
Loop:
    LINE (sliding moves)
    If MOVEN != 0, continue
Return to NEWP
```

#### BISHOP (line 319)

```
Loop:
    LINE (sliding moves)
    If MOVEN == $04, break (done with 4 diagonals)
    Continue
Return to NEWP
```

#### KNIGHT (line 325)

```
Set MOVEN = $10 (16 total, but only 9-16 are knight moves)
Loop:
    SNGMV (L-shaped jump)
    If MOVEN != $08, continue (stop at 8)
Return to NEWP
```

#### PAWN (line 333)

```
Set MOVEN = $06
P1:
    CMOVE (right capture?)
    If capture AND legal:
        Call JANUS
    RESET
    Decrement MOVEN
    If MOVEN == $05:
        Goto P1 (try left capture)

P3:
    CMOVE (forward move)
    If capture OR illegal:
        Return to NEWP (pawns can't capture forward)
    Call JANUS
    If square AND $F0 == $20 (reached 3rd rank):
        Goto P3 (try double move)
Return to NEWP
```

**Special Pawn Logic**:
- MOVEN 6 = right capture diagonal
- MOVEN 5 = left capture diagonal
- MOVEN 4 = forward one square
- After reaching 3rd rank, tries forward again (double move)

**Assembly Reference**: Lines 333-352

---

## Move Execution

### MOVE (line 511)

**Purpose**: Execute a move and save state for UMOVE

**Algorithm**:
```
1. Switch to alternate stack (SP2):
     TSX, STX SP1, LDX SP2, TXS

2. Push move data:
     SQUARE (target) → stack
     Search BOARD for piece at SQUARE
     If found:
         Mark as captured ($CC in BOARD)
         Push piece index
     Else:
         Push $CC (no capture)
     Push BOARD[PIECE] (from square)
     Push PIECE
     Push MOVEN

3. Update board:
     BOARD[PIECE] = SQUARE (move piece)

4. Switch back to hardware stack:
     TSX, STX SP2, LDX SP1, TXS

5. Return
```

**Stack Layout** (5 bytes pushed):
```
[MOVEN] [PIECE] [from_square] [captured_piece] [SQUARE]
 ↑ SP2                                          ↑ initial SP2
```

**Assembly Reference**: Lines 511-539

---

### UMOVE (line 488)

**Purpose**: Undo a move made by MOVE

**Algorithm**:
```
1. Switch to alternate stack:
     TSX, STX SP1, LDX SP2, TXS

2. Pop move data (reverse order):
     PLA → MOVEN
     PLA → PIECE
     PLA → BOARD[PIECE] (restore from square)
     PLA → captured piece index
     If captured piece != $CC:
         PLA → SQUARE
         BOARD[captured_piece] = SQUARE (restore captured piece)
     Else:
         PLA → SQUARE (just restore target square)

3. Switch back to hardware stack:
     (via STRV routine)

4. Return
```

**Assembly Reference**: Lines 488-504

---

### REVERSE (line 382)

**Purpose**: Flip board perspective (swap white/black)

**Algorithm**:
```
For X = $0F down to $00:
    Load BK[X]
    Calculate: $77 - BOARD[X]
    Store in BK[X]
    Load BOARD[X]
    Store in BK[X] original value (swap)
    Calculate: $77 - BOARD[X]
    Store in BOARD[X]
    Decrement X
Loop until X < 0
```

**Effect**:
- BOARD ↔ BK arrays swapped
- All square coordinates transformed: new = $77 - old
- Allows analyzing position from opponent's perspective

**Coordinate Transformation Example**:
```
$00 (a1) → $77 (h8)
$34 (e4) → $43 (d5)
$77 (h8) → $00 (a1)
```

**Assembly Reference**: Lines 382-395

---

### RESET (line 473)

**Purpose**: Restore piece to its current board position

**Algorithm**:
```
Load X = PIECE
Load BOARD[X]
Store in SQUARE
Return
```

**Usage**: Called before trying each new direction in move generation to reset SQUARE to piece's actual location

**Assembly Reference**: Lines 473-476

---

## Position Evaluation

### COUNTS (line 169)

**Purpose**: Count mobility and captures for current position

**Called By**: JANUS when STATE >= 0

**Algorithm**:
```
If PIECE == 0 AND STATE == 8:
    If PIECE == BMAXP:
        Return (don't count black's king captures for white)

Increment MOB[STATE] (mobility)

If PIECE == $01 (Queen):
    Increment MOB[STATE] again (queens count double)

If overflow flag set (capture):
    Search BK array for SQUARE
    When found at index Y:
        Load POINTS[Y] (piece value)
        If POINTS[Y] > MAXC[STATE]:
            Store Y in PCAP[STATE] (best captured piece)
            Store POINTS[Y] in MAXC[STATE]
        Add POINTS[Y] to CC[STATE] (total captures)

If STATE == 4:
    Branch to ON4 (special handling)

Return
```

**Special Case - ON4** (line 208):
```
If STATE == 4:
    Save XMAXC = WCAP0
    Set STATE = 0
    MOVE (make the move)
    REVERSE
    GNMZ (generate all replies)
    REVERSE
    Set STATE = 8
    GNM (generate continuations)
    UMOVE (undo the move)
    Jump to STRATGY (evaluate position)
```

**Assembly Reference**: Lines 169-222

---

### STRATGY (line 641)

**Purpose**: Evaluate position and return score

**Input**: All evaluation counters populated by COUNTS

**Output**: Accumulator = position score (0-255)

**Algorithm** (weighted sum):

```
score = $80 (base value 128)

// Add 0.25 × (parameters)
score += (WMOB + WMAXC + WCC + WCAP1 + WCAP2
          - PMAXC - PCC - BCAP0 - BCAP1 - BCAP2 - PMOB - BMOB) >> 2

// Add 0.50 × (parameters)
score += (WMAXC + WCC - BMAXC) >> 1

// Add 1.00 × (parameters)
score += 4×WCAP0 + WCAP1 - 2×BMAXC - 2×BMCC - BCAP1

// Add positional bonus
If SQUARE in {$22, $25, $33, $34}:  // Center squares
    score += 2
Else if PIECE != 0 AND BOARD[PIECE] < $10:  // Out of back rank
    score += 2

Continue to CKMATE
```

**Evaluation Components**:
- **Mobility** (MOB): Number of legal moves
- **Max Capture** (MAXC): Highest value capturable piece
- **Capture Count** (CC): Sum of all capturable pieces
- **Actual Captures** (CAP): Pieces actually captured at various depths
- **Positional**: Bonus for center control and piece development

**Assembly Reference**: Lines 641-695

---

### CKMATE (line 545)

**Purpose**: Check for check, checkmate, or stalemate

**Algorithm**:
```
If BMAXC == POINTS[0]:  // Can black capture my king?
    Return $00 (illegal position - in check)

If BMOB == 0:  // Black has no moves
    If WMAXP != 0:  // And white's king is not in check
        Return $FF (checkmate!)

Restore STATE = 4
Continue to PUSH (compare with best move)
```

**Return Values**:
- $00 = Illegal (king in check)
- $FF = Checkmate
- (falls through to PUSH otherwise)

**Assembly Reference**: Lines 545-559

---

### PUSH (line 564)

**Purpose**: Compare current move value with best move and save if better

**Input**:
- Accumulator = current move score

**Algorithm**:
```
If score <= BESTV:
    Return (not better)

BESTV = score
BESTP = PIECE
BESTM = SQUARE

Print "." to indicate progress
Return
```

**Side Effect**: Prints "." for each move considered (in original, flashed display)

**Assembly Reference**: Lines 564-573

---

## Search & Analysis

### JANUS (line 162)

**Purpose**: Direct analysis flow based on STATE

**Algorithm**:
```
Load X = STATE
If X < 0:
    Branch to NOCOUNT

// STATE >= 0: Call COUNTS
Call COUNTS subroutine
Return

NOCOUNT:
    If STATE == $F9:
        Branch to check detection logic
    Else:
        Branch to TREE (exchange analysis)
```

**Check Detection** (STATE = $F9, line 229):
```
If SQUARE == BK[0]:  // Captured opponent's king?
    INCHEK = 0 (move is illegal)
Return
```

**State Routing**:
- STATE >= 0 → COUNTS (mobility counting)
- STATE == -7 ($F9) → Check detection
- STATE < 0 (others) → TREE (capture analysis)

**Assembly Reference**: Lines 162-234

---

### TREE (line 240)

**Purpose**: Analyze capture exchanges recursively

**Condition**: Only called if V flag set (capture occurred)

**Algorithm**:
```
Search BK array for SQUARE (captured piece)
When found at index Y:
    Load POINTS[Y]
    If POINTS[Y] > BCAP0[X]:  // X = STATE
        Store in BCAP0[X] (save best capture)

Decrement STATE

If STATE == $FB (-5):
    Branch to UPTREE (max recursion depth)

Call GENRM (generate reply moves)

UPTREE:
Increment STATE
Return
```

**Recursion Depth**: STATE goes from 4 → 3 → 2 → 1 → 0 → -1 → -2 → -3 → -4 → -5

**Assembly Reference**: Lines 240-258

---

### CHKCHK (line 444)

**Purpose**: Determine if a move is legal (doesn't leave king in check)

**Warning**: Expensive operation - only called for STATE 0-7

**Algorithm**:
```
Save current STATE (PHA, PHP)
Set STATE = $F9 (check detection mode)
Set INCHEK = $FF (assume legal)

MOVE (make the move)
REVERSE (switch sides)
GNM (generate all opponent replies)
  // JANUS with STATE=$F9 will set INCHEK=0 if king capturable
RUM (reverse and unmove)

Restore STATE (PLP, PLA)

If INCHEK == $FF:
    Return with C clear (legal)
Else:
    Return with C set (illegal - in check)
```

**Performance**: This is the most expensive routine - generates all opponent moves to detect check

**Assembly Reference**: Lines 444-460

---

## Computer Player

### GO (line 578)

**Purpose**: Select and execute computer's move

**Algorithm**:
```
// Check opening book
Load X = OMOVE
If X >= 0:
    If DIS3 == OPNING[X]:
        Decrement X by 3
        Load response from OPNING
        Display and execute
        Return

// Out of book or mismatch
Set OMOVE = $FF (exit book)

// Generate candidate moves
Set STATE = $0C
Set BESTV = $0C (minimum threshold)
X = $14
Call GNMX (generate with STATE=$0C)

// Analyze each candidate
Set STATE = $04
Call GNMZ (generate and evaluate all moves)

// Check if any legal move found
Load X = BESTV
If X < $0F:
    Return $FF (no legal moves - mate or stalemate)

// Execute best move
Load piece from BESTP
Store in PIECE
Load square from BESTM
Store in SQUARE
Call MOVE
Return to CHESS
```

**Opening Book Logic**:
- Checks if opponent's move (DIS3) matches expected move
- If yes, plays pre-programmed response
- If no, exits book and starts thinking

**Assembly Reference**: Lines 578-620

---

## Utility Routines

### GENRM (line 480)

**Purpose**: Make move, generate replies, unmove

**Algorithm**:
```
Call MOVE
Call GENR2

GENR2:
Call REVERSE
Call GNM
Call RUM

RUM:
Call REVERSE
Call UMOVE
Return
```

**Usage**: Used in TREE for analyzing capture sequences

**Assembly Reference**: Lines 480-487

---

## Display & I/O

### OUT (line 110)

**Purpose**: Display board and get input

**Algorithm**:
```
Call POUT (display board)
Call KIN (get key input)
Return with key in accumulator
```

**Assembly Reference**: Lines 110-114

---

### POUT (line 702)

**Purpose**: Print board to serial terminal

**Algorithm**:
```
Print copyright banner (POUT13)
Print column labels (POUT10)
Print top border (POUT5)

For each rank (Y = $00 to $70, step $10):
    Print "|"
    For each file (X = 0 to 7):
        Search BOARD for piece at Y
        If found:
            Print piece character (POUT4)
        Else:
            Print empty square (* or space based on color)
        Increment Y
    Print "|"
    Print rank number (POUT12)
    Print CRLF
    Print border

Print LED values ($FB, $FA, $F9 as hex)
Print CRLF
Return
```

**Board Characters** (CPL/CPH tables):
- W/B = White/Black
- K = King, Q = Queen, R = Rook, B = Bishop
- N = Knight, P = Pawn

**Assembly Reference**: Lines 702-810

---

### KIN (line 812)

**Purpose**: Get keyboard input from serial port

**Algorithm**:
```
Print "?"
Call SYSKIN (wait for serial input)
AND #$4F (mask to keep 0-7 and A-Z)
Return with key in accumulator
```

**Masking**: Converts ASCII to useful range

**Assembly Reference**: Lines 812-816

---

### POUT5 (line 760)

**Purpose**: Print horizontal border line

**Algorithm**:
```
Save X
Load X = $19 (25 dashes)
Loop:
    Print "-"
    Decrement X
Loop until X == 0
Restore X
Print CRLF
Return
```

**Assembly Reference**: Lines 760-770

---

## Serial I/O (6551 ACIA)

### Init_6551 (line 821)

**Purpose**: Initialize 6551 serial chip

**Configuration**:
- Control: $1F (19.2K baud, 8 bits, 1 stop bit)
- Command: $0B (no parity, echo off, RX int off, DTR active low)

**Assembly Reference**: Lines 821-825

---

### SYSKIN (line 829)

**Purpose**: Wait for and read character from serial port

**Algorithm**:
```
Loop:
    Load ACIASta (status register)
    AND #$08 (test receive buffer full)
    If zero, loop (wait)
Load ACIAdat (data register)
Return with character in accumulator
```

**Assembly Reference**: Lines 829-833

---

### SYSCHOUT (line 837)

**Purpose**: Send character to serial port

**Algorithm**:
```
Push accumulator (save character)
Loop:
    Load ACIASta (status register)
    AND #$10 (test transmit buffer empty)
    If zero, loop (wait)
Pop accumulator (restore character)
Store in ACIAdat (data register)
Return
```

**Assembly Reference**: Lines 837-843

---

### SYSHEXOUT (line 845)

**Purpose**: Print byte as two hex digits

**Algorithm**:
```
Push accumulator
Shift right 4 times (high nibble)
Call PRINTDIG
Pop accumulator

PRINTDIG:
AND #$0F (mask low nibble)
Convert to hex character using HEXDIGDATA table
Call SYSCHOUT
Return
```

**Assembly Reference**: Lines 845-857

---

## Summary

The subroutines demonstrate remarkable code density and clever optimization:

- **Move Generation**: Unified handling via piece type dispatch
- **Evaluation**: Weighted multi-depth analysis
- **Stack Management**: Dual stacks separate concerns
- **State Machine**: Single variable (STATE) controls complex flow
- **I/O**: Clean abstraction over hardware

Total code size: ~1.5KB including all data tables - an extraordinary achievement for a functional chess program!
