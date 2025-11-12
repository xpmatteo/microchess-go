# CMOVE Pseudocode

## Overview

CMOVE (Calculate Move) is the core move validation routine in MicroChess. It calculates the target square for a move and sets processor flags to indicate the move's legality and characteristics.

**Assembly Location**: Lines 407-469 in Microchess6502.txt

## Global Variables Read

| Variable | Address | Description | Usage in CMOVE |
|----------|---------|-------------|----------------|
| `SQUARE` | `$B1` | Current/target square being evaluated | Read to get starting position, then updated with new position after adding move offset |
| `MOVEN` | `$B6` | Current move offset index into MOVEX table | Used to index into MOVEX to get direction offset |
| `MOVEX` | `$1589` | Direction offset table (17 bytes) | Read to get offset for current move direction |
| `BOARD` | `$50-$5F` | Piece locations for white (16 bytes) | Scanned to check if target square is occupied and by which side |
| `BK` | `$60-$6F` | Piece locations for black (16 bytes) | Scanned along with BOARD (as continuous 32-byte array) to find occupying piece |
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
- `MOVE()` - Make the trial move
- `REVERSE()` - Switch board perspective to opponent's side
- `GNM()` - Generate all moves for opponent
- `RUM()` - Reverse board and unmake move (calls REVERSE then UMOVE)

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
    for pieceIndex from 31 down to 0:
        if BOARD[pieceIndex] == newSquare:
            // Square is occupied!

            // Check if occupied by own piece (indices 0-15)
            if pieceIndex < 16:
                goto ILLEGAL  // Blocked by own piece

            // Must be opponent's piece (indices 16-31)
            // Set V flag to indicate capture
            A = 0x7F
            A = A + 1     // 0x7F + 1 = 0x80, sets V flag
            if V_FLAG:    // Always true here
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
    // Save current state
    PUSH(A)
    PUSH(ProcessorStatus)

    savedState = STATE
    STATE = 0xF9  // -7: Check detection mode
    INCHEK = 0xF9  // Assume king is safe

    // Make the trial move
    MOVE()

    // Switch sides and generate all opponent replies
    REVERSE()
    GNM()  // Generate all moves for opponent

    // Restore board (REVERSE and UMOVE)
    RUM()

    // Restore saved state
    POP(ProcessorStatus)
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
    A = 0xFF
    CLC  // Clear carry (different from check)
    CLV  // Clear overflow (no capture)
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

## Usage Example

```assembly
; Generate knight moves
KNIGHT  LDX     #$10
        STX     MOVEN           ; Start at move offset 16
AGNN    JSR     SNGMV           ; SNGMV calls CMOVE
        LDA     MOVEN
        CMP     #$08
        BNE     AGNN            ; Continue until offset 8

; SNGMV (line 357):
SNGMV   JSR     CMOVE           ; Calculate move
        BMI     ILL1            ; If illegal (N flag), skip
        JSR     JANUS           ; Evaluate legal move
ILL1    JSR     RESET           ; Restore piece position
        DEC     MOVEN
        RTS
```

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
