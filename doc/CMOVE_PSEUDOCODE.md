# CMOVE Pseudocode

## Overview

CMOVE (Calculate Move) is the core move validation routine in MicroChess. It calculates the target square for a move and sets processor flags to indicate the move's legality and characteristics.

**Assembly Location**: Lines 407-469 in Microchess6502.txt

CMOVE is NOT used to validate user moves. When you enter a move via the UI:

1. INPUT routine (lines 262-273): Reads your 4 digits, finds which piece is at the FROM square
2. Enter key pressed (line 146-149): Directly calls MOVE() - no CMOVE validation
3. MOVE routine (lines 511-539): Blindly executes whatever move you entered

CMOVE has exactly two purposes, both for the AI only:

1. AI Move Generation (lines 286-352 → CMOVE): When the computer generates its moves, CMOVE filters out illegal moves (off board, blocked by own pieces)
2. Check Detection (lines 444-460 CHKCHK): When evaluating AI moves, CMOVE verifies they don't leave the king in check by making a trial move and seeing if opponent could capture the king

## Critical: Current Player Perspective

**CMOVE always analyzes moves from the current player's perspective.** It validates whether the current player can legally make a specific move.

The routine has no concept of "white" or "black" or "AI" vs "human". Instead:
- `BOARD[0-15]` always contains the **current player's pieces** (the side making the move)
- `BK[0-15]` always contains the **opponent's pieces** (the side responding to the move)

When the game switches turns, `REVERSE()` is called to:
1. Swap `BOARD` ↔ `BK` (current player becomes opponent, vice versa)
2. Transform all coordinates (`newSquare = $77 - oldSquare`)

This ensures CMOVE always sees the board from the current player's viewpoint, regardless of which color or player (AI/human) is actually moving.

**Example**: When checking if a white move is legal, CMOVE sees white pieces in `BOARD[0-15]`. After `REVERSE()`, when checking black's responses, CMOVE sees black pieces in `BOARD[0-15]`. The routine's logic remains identical.

## Global Variables Read

| Variable | Address | Description | Usage in CMOVE |
|----------|---------|-------------|----------------|
| `SQUARE` | `$B1` | Current/target square being evaluated | Read to get starting position, then updated with new position after adding move offset |
| `MOVEN` | `$B6` | Current move offset index into MOVEX table | Used to index into MOVEX to get direction offset |
| `MOVEX` | `$1589` | Direction offset table (17 bytes) | Read to get offset for current move direction |
| `BOARD` | `$50-$5F` | Current player's piece locations (16 bytes) | Scanned as part of continuous 32-byte array [BOARD+BK] to check if target square is occupied |
| `BK` | `$60-$6F` | Opponent's piece locations (16 bytes) | Scanned via BOARD,X indexing when X≥16 (memory layout: BOARD immediately followed by BK) |
| `STATE` | `$B5` | Analysis state machine value | Checked to determine if CHKCHK (check verification) should run |
| `INCHEK` | `$B4` | Check detection flag ($FF = safe, $00 = king capturable) | Set to $F9 before check detection, read after GNM to determine if king is safe |
| `PIECE` | `$B0` | Current piece index being moved (0-15) | Used by MOVE/UMOVE during CHKCHK |

## Global Variables Written

| Variable | Address | Description | How CMOVE Modifies It |
|----------|---------|-------------|----------------------|
| `SQUARE` | `$B1` | Current/target square being evaluated | Updated with calculated new position (SQUARE + MOVEX[MOVEN]) |
| `STATE` | `$B5` | Analysis state machine value | Temporarily set to $F9 during CHKCHK, then restored |
| `INCHEK` | `$B4` | Check detection flag | Set to $F9 before GNM call in CHKCHK |

## Subroutines Called

When CHKCHK is triggered (STATE between 0-7):
- `MOVE()` - Make the trial move (uses dual-stack SP1/SP2 mechanism)
- `REVERSE()` - Switch board perspective to opponent's side
- `GNM()` - Generate all moves for opponent
- `RUM()` - Reverse board and unmake move (calls REVERSE, then **falls through** to UMOVE - no explicit JSR)

## Pseudocode

```pseudocode
// CMOVE - Calculate Move and Set Flags
// This routine calculates the target square for a move and sets flags to indicate:
//   N flag (Negative): Illegal move (off board or blocked by own piece)
//   V flag (oVerflow): Capture possible (opponent piece on target square)
//   C flag (Carry):    Illegal because move leaves king in check
//
// Returns with A = $00 (legal) or A = $FF (illegal)

function CMOVE():
    // Calculate new square position
    newSquare = SQUARE + MOVEX[MOVEN]
    SQUARE = newSquare

    // Check if off board using 0x88 trick
    if (newSquare & 0x88) != 0:
        goto ILLEGAL

    // Scan all 32 pieces to see if target square is occupied
    // Note: BOARD and BK form a continuous 32-byte array in memory ($50-$6F)
    // The assembly uses BOARD,X with X from 31 down to 0:
    //   X=0-15  accesses BOARD[0-15] at $50-$5F (your pieces)
    //   X=16-31 accesses BK[0-15] at $60-$6F (opponent pieces)
    for pieceIndex from 31 down to 0:
        if BOARD[pieceIndex] == newSquare:
            // Square is occupied!

            // Check if occupied by own piece (indices 0-15)
            if pieceIndex < 16:
                goto ILLEGAL  // Blocked by own piece

            // Must be opponent's piece (indices 16-31)
            // Set V flag to indicate capture using signed overflow trick
            A = 0x7F      // Load +127 (maximum positive signed byte)
            A = A + 1     // 0x7F + 1 = 0x80 = -128 in two's complement
                          // This causes signed overflow: +127 → -128
                          // The V flag is set when sign changes unexpectedly
            if V_FLAG:    // Always true here (overflow occurred)
                goto SPX  // Jump to check-check logic

    // No piece on target square
    CLV  // Clear V flag (no capture)

SPX:
    // Should we verify this move doesn't leave king in check?
    if STATE < 0:
        goto RETL  // Skip check-check for deep analysis

    if STATE >= 8:
        goto RETL  // Skip check-check for continuation moves

    // Do the expensive CHKCHK verification
CHKCHK:
    // Save current state (order matters: A first, then flags)
    PUSH(A)
    PUSH(ProcessorStatus)

    savedState = STATE
    STATE = 0xF9  // -7: Check detection mode
    INCHEK = 0xF9  // Assume king is safe (will only change to $00 if king capturable)

    // Make the trial move using dual-stack mechanism
    // MOVE() switches to game state stack (SP2) to save move, then back to call stack (SP1)
    // This allows the move to be unmade later without corrupting the call stack
    MOVE()

    // Switch sides and generate all opponent replies
    REVERSE()
    GNM()  // Generate all moves for opponent
           // If any move can capture BK[0] (king), JANUS sets INCHEK=$00

    // Restore board: REVERSE() then fall through to UMOVE()
    // RUM() reverses the board and unmakes the move using the game state stack (SP2)
    RUM()

    // Restore saved state (order matters: flags first, then A)
    POP(ProcessorStatus)  // Restores calling context's flag state
    POP(A)
    STATE = savedState

    // Check result
    if INCHEK < 0:
        // King is safe
        goto RETL
    else:
        // King can be captured! Illegal move
        SEC  // Set carry flag to indicate check
        A = 0xFF
        return

RETL:
    // Legal move
    CLC  // Clear carry flag
    A = 0x00
    return

ILLEGAL:
    // Illegal move (off board or blocked by own piece)
    A = 0xFF      // $FF is negative in two's complement (-1)
                  // Loading $FF automatically sets the N (Negative) flag
                  // This is how BMI (Branch if Minus) detects illegal moves
    CLC           // Clear carry (C flag clear = not in check, just illegal)
    CLV           // Clear overflow (V flag clear = no capture possible)
    return
```

## Return Values via Processor Flags

CMOVE uses the 6502 processor status flags to encode multiple return states:

| Flags | A Value | Meaning | Description |
|-------|---------|---------|-------------|
| N=1, V=0, C=0 | `$FF` | Illegal move | Off board or blocked by own piece |
| N=0, V=1, C=0 | `$00` | Legal capture | Opponent piece on target square |
| N=0, V=0, C=0 | `$00` | Legal non-capture | Empty square, move is safe |
| N=1, V=?, C=1 | `$FF` | Illegal (check) | Move would leave king in check |

**Testing the flags in calling code:**
- `BMI` (Branch if Minus/Negative) - checks if move is illegal
- `BVC` (Branch if oVerflow Clear) - checks if no capture
- `BVS` (Branch if oVerflow Set) - checks if capture
- `BCC` (Branch if Carry Clear) - checks if not in check
- `BCS` (Branch if Carry Set) - checks if in check

## Memory Layout

CMOVE relies on a specific memory layout where piece arrays are contiguous:

```
Address Range | Variable    | Contents
--------------|-------------|------------------------------------------
$50-$57       | BOARD[0-7]  | Current player's pieces: King, Queen, Rooks, Bishops
$58-$5F       | BOARD[8-15] | Current player's pieces: Knights, Pawns
$60-$67       | BK[0-7]     | Opponent's pieces: King, Queen, Rooks, Bishops
$68-$6F       | BK[8-15]    | Opponent's pieces: Knights, Pawns

Key insight: When assembly uses "LDA BOARD,X" with X=16-31, it actually
accesses BK[0-15] because BK immediately follows BOARD in memory.

"Current player" and "opponent" are perspective-dependent. After REVERSE()
is called, the arrays swap roles (current player becomes opponent and vice versa).
```

This continuous layout enables efficient scanning with a single loop from index 31 down to 0.

## Key Algorithmic Points

### 1. 0x88 Edge Detection

The brilliant insight of 0x88 board representation: any off-board square has bit $08 or $80 set.

```
Valid squares: $00-$07, $10-$17, $20-$27, ... $70-$77
Off-board examples:
  $08 - right edge overflow (file 8)
  $80 - top edge overflow (rank 8)
  $FF - left/bottom wrap
```

Single instruction `AND #$88` followed by `BNE` detects all board edges.

### 2. Piece Ownership Detection

Pieces are indexed in a continuous 32-byte array:
- Indices 0-15: Your pieces (white from white's perspective)
- Indices 16-31: Opponent's pieces (black from white's perspective)

Simple comparison `CPX #$10` determines if collision is a block or a capture.

### 3. Conditional Check Verification

CHKCHK is expensive (makes trial move, reverses board, generates all opponent moves), so it's only called when:
- `STATE >= 0` (not during deep capture analysis like STATE = -1 to -5)
- `STATE < 8` (not during continuation move generation)

This optimization is critical for performance on a 1 MHz CPU.

### 4. STATE = -7 (0xF9) for Check Detection

When CHKCHK runs, it temporarily sets `STATE = $F9` (-7 in signed interpretation). This causes the JANUS analysis director to skip evaluation and just check if `BK[0]` (the king) can be captured.

See JANUS routine (line 223-234):
```assembly
NOCOUNT CPX     #$F9
        BNE     TREE
        LDA     BK              ; IS KING
        CMP     SQUARE          ; IN CHECK?
        BNE     RETJ
        LDA     #$00            ; YES - SET
        STA     INCHEK          ; INCHEK=0
RETJ    RTS
```

If any generated move has `SQUARE == BK[0]`, then `INCHEK` is set to $00, indicating the king can be captured.

**INCHEK initial state**: Set to $F9 before GNM runs. This value means "not yet modified = king is safe". Only if GNM finds a king capture does JANUS change it to $00. This is more efficient than initializing to "unsafe" and proving safety.

## Performance Characteristics

CMOVE's performance varies dramatically based on whether CHKCHK runs:

**Without CHKCHK** (STATE < 0 or STATE ≥ 8):
- ~40-60 cycles for off-board detection
- ~50-80 cycles for collision detection
- Total: **40-80 cycles** depending on piece scan length

**With CHKCHK** (STATE = 0-7):
- Base CMOVE: ~40-80 cycles
- MOVE: ~50 cycles (stack switching + board update)
- REVERSE: ~320 cycles (16 pieces × 2 transformations)
- GNM: **~2,000-10,000 cycles** (generates all opponent moves)
- RUM: ~370 cycles (REVERSE + UMOVE)
- Total: **~2,500-11,000 cycles**

**Critical insight**: CHKCHK is 30-140× slower than basic CMOVE. The STATE machine carefully controls when this expensive verification runs, making it a key optimization for 1 MHz CPU performance.

## Historical Context

CMOVE demonstrates several 6502 assembly idioms common in 1976:

1. **Flag-based returns**: Uses processor flags instead of explicit return values
2. **Computed jumps**: `BVC LINE` leverages V flag state to control flow
3. **Minimal branching**: Falls through when possible, minimizing branch instructions
4. **Page zero optimization**: All variables in fast memory ($00-$FF)

The routine is remarkably compact (~60 bytes) yet handles:
- Board edge detection
- Collision detection
- Capture identification
- Check verification (when needed)

This exemplifies the code density required for 1976 hardware constraints (~1KB total program space).

## Common Calling Patterns

CMOVE is called by three main movement routines, each handling different piece types:

### 1. SNGMV - Single-step moves (King, Knight)

```assembly
; Used for pieces that move one step at a time in 8 or 16 directions
; Example: Generate knight moves (lines 325-331)
KNIGHT  LDX     #$10
        STX     MOVEN           ; Start at move offset 16
AGNN    JSR     SNGMV           ; SNGMV calls CMOVE
        LDA     MOVEN
        CMP     #$08
        BNE     AGNN            ; Continue until offset 8

; SNGMV routine (line 357):
SNGMV   JSR     CMOVE           ; Calculate move
        BMI     ILL1            ; If illegal (N flag set), skip
        JSR     JANUS           ; Evaluate legal move
ILL1    JSR     RESET           ; Restore piece position
        DEC     MOVEN
        RTS
```

### 2. LINE - Sliding moves (Queen, Rook, Bishop)

```assembly
; Used for pieces that slide along lines until blocked or capture
; Example: Rook moves (lines 313-317)
ROOK    LDX     #$04
        STX     MOVEN           ; Moves 4 to 1 (cardinal directions)
AGNR    JSR     LINE            ; LINE calls CMOVE repeatedly
        BNE     AGNR

; LINE routine (line 367):
LINE    JSR     CMOVE           ; Calculate move
        BCC     OVL             ; If check flag clear, continue
        BVC     LINE            ; If no capture, continue sliding
OVL     BMI     ILL             ; If illegal, stop this direction
        PHP
        JSR     JANUS           ; Evaluate legal move
        PLP
        BVC     LINE            ; If no capture, keep sliding
ILL     JSR     RESET           ; Restore piece, direction exhausted
        DEC     MOVEN
        RTS
```

### 3. CMOVE direct calls - Pawn moves

```assembly
; Pawns have special logic (lines 333-352)
PAWN    LDX     #$06
        STX     MOVEN
P1      JSR     CMOVE           ; Right capture diagonal
        BVC     P2              ; If no capture (V flag clear), skip
        BMI     P2              ; If illegal (N flag set), skip
        JSR     JANUS           ; Evaluate capture
P2      JSR     RESET
        DEC     MOVEN
        ; ... continues for left capture and forward move
```

**Key pattern differences**:
- **SNGMV**: Single CMOVE call per direction, always evaluates result
- **LINE**: Repeated CMOVE calls until illegal/capture, uses BVC to check capture flag
- **Pawn**: Direct CMOVE calls with custom logic for captures vs. advances

## Go Port Considerations

When porting to Go, consider:

1. **Return struct instead of flags**:
   ```go
   type MoveResult struct {
       Legal   bool  // false if N flag would be set
       Capture bool  // true if V flag would be set
       InCheck bool  // true if C flag would be set
   }
   ```

2. **Explicit state instead of goto**:
   ```go
   if !result.Legal {
       return MoveResult{Legal: false}
   }
   ```

3. **Board as continuous array**: Keep the 32-piece array structure for efficient scanning

4. **Preserve 0x88 logic**: The `& 0x88` check is elegant and should be kept

5. **Conditional CHKCHK**: Maintain the STATE-based optimization to avoid expensive check verification when not needed
