# GNM (Generate New Move) Routine Analysis

## Overview

GNM (Generate New Move) is the core move generation routine in MicroChess. It systematically generates all pseudo-legal moves for one side by iterating through all 16 pieces (indexes 0-15) and generating moves appropriate to each piece type. After each move is calculated, JANUS is called to evaluate or analyze the position.

**Location**: Lines 280-352 in Microchess6502.txt

## Purpose

The GNM routine serves as MicroChess's move generator - the fundamental component that:
1. Iterates through all pieces for the current side (pieces 0-15)
2. Determines each piece's type (King, Queen, Rook, Bishop, Knight, or Pawn)
3. Generates all possible moves for that piece using type-specific logic
4. Calls JANUS after each generated move for evaluation/counting
5. Returns when all pieces have been processed

## Inputs

### Global Variables (Page Zero)
- **PIECE** ($B0): Current piece index being processed (0-15)
  - 0 = King
  - 1 = Queen
  - 2-3 = Rooks
  - 4-5 = Bishops
  - 6-7 = Knights
  - 8-15 = Pawns

- **BOARD** ($50-$5F): Array containing current location of each piece (16 bytes)
  - BOARD[PIECE] gives the 0x88 coordinate of piece PIECE

- **SQUARE** ($B1): Working variable for destination square during move calculation

- **MOVEN** ($B6): Move direction index into MOVEX table
  - Different starting values for different piece types
  - Decremented as moves are tried

- **STATE** ($B5): Controls depth/type of analysis
  - Negative values: Special modes (e.g., -7 for check detection)
  - 0, 4, 8, 12: Normal search states at different depths

- **MOVEX** ($1580+): Move offset table
  - Index 0: $00 (king/queen move 0)
  - Index 1: $F0 (king/queen move 1: -16 = one rank down)
  - Index 2: $FF (king/queen move 2: -1 = one file left)
  - Index 3: $01 (king/queen move 3: +1 = one file right)
  - Index 4: $10 (king/queen move 4: +16 = one rank up)
  - Index 5: $11 (king/queen move 5: diagonal)
  - Index 6: $0F (king/queen move 6: diagonal)
  - Index 7: $EF (king/queen move 7: diagonal)
  - Index 8: $F1 (king/queen move 8: diagonal)
  - Index 9-16: Knight moves (L-shaped offsets like $DF, $E1, $EE, $F2, $12, $0E, $1F, $21)

## Outputs

### Side Effects
- **Modifies PIECE**: Decremented from 15 down to 0 (actually $10 to $00 in code)
- **Modifies MOVEN**: Set and decremented based on piece type
- **Modifies SQUARE**: Updated during move calculation
- **Calls JANUS**: After each legal move, causing various counters to be updated
  - MOB (mobility counters)
  - MAXC (maximum capture values)
  - CC (capture counts)
  - PCAP (piece captured indexes)

### Return Value
None (RTS when PIECE becomes negative, i.e., all pieces processed)

## Entry Points

### GNMZ (Line 280)
**Purpose**: Clear counters, then fall through to GNM

Clears 17 bytes starting at COUNT ($DE):
- COUNT/BCAP2
- WCAP2
- BCAP1
- WCAP1
- BCAP0
- MOB/BMOB
- MAXC/BMAXC
- CC/BMCC
- PCAP/BMAXP
- (and 8 more bytes through $EE)

### GNMX (Line 281)
**Purpose**: Entry point that clears counters without reloading X=$10

Used when caller has already set up X register.

### GNM (Line 286)
**Purpose**: Main entry point - sets PIECE=$10, then generates moves

## Algorithm Flow

```
GNM:
    PIECE = 0x10 (16 pieces: 0-15)

    NEWP:  // Process next piece
        PIECE = PIECE - 1
        if PIECE < 0:
            return  // All pieces processed

        RESET()  // Set SQUARE = BOARD[PIECE] (piece's current location)
        MOVEN = 8  // Default starting move index for most pieces

        // Determine piece type by index
        if PIECE >= 8:
            goto PAWN
        if PIECE >= 6:
            goto KNIGHT
        if PIECE >= 4:
            goto BISHOP
        if PIECE == 1:
            goto QUEEN
        if PIECE >= 1:
            goto ROOK
        // else fall through to KING (PIECE == 0)
```

### Piece-Specific Move Generation

#### KING (Piece 0)
```
KING:
    // Generate 8 single-step moves (MOVEN 8 down to 1)
    while MOVEN != 0:
        SNGMV()     // Try one move direction
        MOVEN--
    goto NEWP

SNGMV (Single Move):
    CMOVE()         // Calculate move, sets flags
    if N flag set:  // Illegal move
        goto ILL1
    JANUS()         // Evaluate position
ILL1:
    RESET()         // Restore SQUARE = BOARD[PIECE]
    MOVEN--
    return
```

#### QUEEN (Piece 1)
```
QUEEN:
    // Generate 8 sliding moves (MOVEN 8 down to 1)
    while MOVEN != 0:
        LINE()      // Try one direction until blocked
        MOVEN--
    goto NEWP

LINE (Sliding Line):
    loop:
        CMOVE()         // Calculate next square in this direction
        if C flag set and V flag clear:  // Check without capture
            continue loop
        if N flag set:  // Hit edge or illegal
            goto ILL
        save flags
        JANUS()         // Evaluate
        restore flags
        if V flag clear:  // Not a capture
            continue loop  // Keep sliding
    ILL:
        RESET()         // Restore starting square
        MOVEN--
        return
```

#### ROOK (Pieces 2-3)
```
ROOK:
    MOVEN = 4       // Only 4 orthogonal directions
AGNR:
    LINE()          // Generate sliding moves
    if MOVEN != 0:
        goto AGNR
    goto NEWP
```

#### BISHOP (Pieces 4-5)
```
BISHOP:
    // Start with MOVEN=8, generate moves 8,7,6,5 (diagonal directions)
BISHOP_LOOP:
    LINE()
    if MOVEN != 4:  // Stop at move 4 (only use diagonal offsets)
        goto BISHOP_LOOP
    goto NEWP
```

#### KNIGHT (Pieces 6-7)
```
KNIGHT:
    MOVEN = 0x10 (16)  // Knight has 16 move slots, uses 8
AGNN:
    SNGMV()            // Single step (no sliding)
    if MOVEN != 8:     // Process moves 16 down to 9
        goto AGNN
    goto NEWP
```

#### PAWN (Pieces 8-15)
```
PAWN:
    MOVEN = 6           // Start with right capture

P1: // Try right diagonal capture (MOVEN=6)
    CMOVE()
    if V flag clear:    // Not a capture
        goto P2
    if N flag set:      // Illegal
        goto P2
    JANUS()             // Legal capture

P2:
    RESET()
    MOVEN--             // Now MOVEN=5 (left capture)
    if MOVEN != 5:
        goto P1

P3: // Forward move(s)
    CMOVE()
    if V flag set:      // Blocked by piece
        goto NEWP
    if N flag set:      // Off board
        goto NEWP
    JANUS()             // Legal forward move

    // Check if pawn reached 3rd rank (can do double move)
    if (SQUARE & 0xF0) == 0x20:  // On rank 2
        goto P3         // Try another forward move
    goto NEWP
```

## Key Subroutines Called

### RESET (Line 473)
**Purpose**: Restore SQUARE to piece's current board position
```
RESET:
    X = PIECE
    A = BOARD[X]    // Get piece's location
    SQUARE = A      // Set working square
    return
```

### CMOVE (Line 407)
**Purpose**: Calculate destination square and determine move legality

**Inputs**:
- SQUARE: Current position
- MOVEN: Index into MOVEX table
- MOVEX[MOVEN]: Offset to add

**Outputs** (via CPU flags):
- **N flag**: Set if move is illegal (off board or captures own piece)
- **V flag**: Set if move captures opponent piece
- **C flag**: Set if move leaves king in check (only when STATE=0-7)

**Algorithm**:
```
CMOVE:
    A = SQUARE
    X = MOVEN
    A = A + MOVEX[X]    // Add move offset
    SQUARE = A           // New position

    if (A & 0x88) != 0:  // Off board (0x88 board check)
        return ILLEGAL   // N=1, C=0, V=0

    // Check if destination occupied
    for X = 0x20 down to 0:
        if BOARD[X] == SQUARE:
            if X < 0x10:     // Occupied by own piece
                return ILLEGAL
            else:            // Occupied by opponent
                return CAPTURE  // N=0, V=1, C determined by check-check

    // Empty square
    CLV  // V=0 (no capture)

    // Should we verify not moving into check?
    if STATE < 0:
        return LEGAL     // N=0, V=0, C=0
    if STATE >= 8:
        return LEGAL

    // Do check-check (expensive)
    CHKCHK:
        save A, flags
        STATE = 0xF9     // Special check-detection mode
        INCHEK = 0xF9    // Flag for king capture detection
        MOVE()           // Make the trial move
        REVERSE()        // Switch sides
        GNM()            // Generate all opponent replies
        REVERSE()        // Switch back
        UMOVE()          // Unmake trial move
        restore flags, A
        if INCHEK == 0:  // King was captured
            return IN_CHECK  // N=1, C=1
        return LEGAL     // N=0, C=0
```

### JANUS (Line 162)
**Purpose**: Direct analysis after each move

**Behavior** depends on STATE:
- **STATE=12**: Count moves (initial ply)
- **STATE=4**: Count moves and analyze captures (main ply)
- **STATE=0**: Count opponent replies
- **STATE=8**: Count continuation moves
- **STATE=-1 to -5**: Capture exchange analysis
- **STATE=-7 (0xF9)**: Check detection mode

**Key actions**:
- Increments mobility counters (MOB)
- Tracks maximum captures (MAXC)
- Accumulates capture counts (CC)
- Triggers deeper analysis at STATE=4 (calls MOVE, REVERSE, recursive GNM)

## Move Direction Tables

### MOVEX Values (0x88 Board Offsets)
```
Index   Hex    Dec    Direction
-----   ---    ---    ---------
  0     $00      0    (placeholder/king center)
  1     $F0    -16    South (rank -1)
  2     $FF     -1    West (file -1)
  3     $01     +1    East (file +1)
  4     $10    +16    North (rank +1)
  5     $11    +17    Northeast
  6     $0F    +15    Northwest
  7     $EF    -17    Southwest
  8     $F1    -15    Southeast
  9     $DF    -33    Knight move
 10     $E1    -31    Knight move
 11     $EE    -18    Knight move
 12     $F2    -14    Knight move
 13     $12    +18    Knight move
 14     $0E    +14    Knight move
 15     $1F    +31    Knight move
 16     $21    +33    Knight move
```

### Piece Type Move Ranges
- **King**: MOVEN 8→1 (directions 1-8: all adjacent squares)
- **Queen**: MOVEN 8→1 (directions 1-8: all 8 rays)
- **Rook**: MOVEN 4→1 (directions 1-4: orthogonal rays)
- **Bishop**: MOVEN 8→5 (directions 5-8: diagonal rays)
- **Knight**: MOVEN 16→9 (directions 9-16: L-shaped jumps)
- **Pawn**: MOVEN 6, 5, 4 (6=right capture, 5=left capture, 4=forward)

## Detailed Pseudocode for Reimplementation

```go
// Entry point: Clear counters and generate all moves
func GNMZ() {
    // Clear 17 bytes of counters
    for x := 0x10; x >= 0; x-- {
        COUNT[x] = 0  // Clears COUNT through PMOB range
    }
    GNM()
}

// Main move generation routine
func GNM() {
    PIECE = 0x10  // Start with piece 16 (will decrement to 15 first)

    for {
        PIECE--
        if PIECE < 0 {
            return  // All 16 pieces processed
        }

        // Get piece's current location
        RESET()  // Sets SQUARE = BOARD[PIECE]

        // Set default MOVEN for most pieces
        MOVEN = 8

        // Dispatch based on piece type
        switch {
        case PIECE >= 8:
            generatePawnMoves()
        case PIECE >= 6:
            generateKnightMoves()
        case PIECE >= 4:
            generateBishopMoves()
        case PIECE == 1:
            generateQueenMoves()
        case PIECE >= 1:
            generateRookMoves()
        default: // PIECE == 0
            generateKingMoves()
        }
    }
}

func generateKingMoves() {
    // Generate 8 single-step moves in all directions
    for MOVEN != 0 {
        flags := CMOVE()
        if !flags.Negative {  // Legal move
            JANUS()
        }
        RESET()
        MOVEN--
    }
}

func generateQueenMoves() {
    // Generate 8 sliding rays
    for MOVEN != 0 {
        generateLine()
        MOVEN--
    }
}

func generateRookMoves() {
    MOVEN = 4  // Only 4 orthogonal directions
    for MOVEN != 0 {
        generateLine()
        MOVEN--
    }
}

func generateBishopMoves() {
    // Generate 4 diagonal rays (moves 8,7,6,5)
    for MOVEN != 4 {
        generateLine()
        MOVEN--
    }
}

func generateKnightMoves() {
    MOVEN = 0x10  // Start at move 16
    for MOVEN != 8 {  // Process down to move 9
        flags := CMOVE()
        if !flags.Negative {
            JANUS()
        }
        RESET()
        MOVEN--
    }
}

func generatePawnMoves() {
    MOVEN = 6  // Right diagonal capture

    // Try right capture
    flags := CMOVE()
    if flags.Overflow && !flags.Negative {  // V=1, N=0 means legal capture
        JANUS()
    }

    // Try left capture
    RESET()
    MOVEN = 5
    flags = CMOVE()
    if flags.Overflow && !flags.Negative {
        JANUS()
    }

    // Try forward move(s)
    RESET()
    MOVEN = 4
    for {
        flags = CMOVE()
        if flags.Overflow || flags.Negative {  // Blocked or illegal
            return
        }
        JANUS()  // Legal forward move

        // Check if on 3rd rank (can make double move)
        if (SQUARE & 0xF0) != 0x20 {
            return  // Not on rank 2, done
        }
        // On rank 2, try another forward move (no RESET - keep moving forward)
    }
}

func generateLine() {
    // Generate all moves in one direction until blocked
    for {
        flags := CMOVE()

        // If check without capture, continue sliding
        if flags.Carry && !flags.Overflow {
            continue
        }

        // If illegal (off board or own piece), stop this direction
        if flags.Negative {
            RESET()
            return
        }

        // Legal move - evaluate it
        JANUS()

        // If capture, stop sliding in this direction
        if flags.Overflow {
            RESET()
            return
        }

        // Empty square, continue sliding
    }
}

func RESET() {
    SQUARE = BOARD[PIECE]
}

// Move calculation - returns flags indicating move legality
type MoveFlags struct {
    Negative bool  // N: Illegal (off board or captures own piece)
    Overflow bool  // V: Capture (legal capture of opponent piece)
    Carry    bool  // C: In check (only when STATE in range for check-check)
}

func CMOVE() MoveFlags {
    // Calculate destination square
    newSquare := SQUARE + MOVEX[MOVEN]
    SQUARE = newSquare

    // Check if off board (0x88 representation)
    if (newSquare & 0x88) != 0 {
        return MoveFlags{Negative: true, Overflow: false, Carry: false}
    }

    // Check if square occupied
    for x := 0x1F; x >= 0; x-- {
        if BOARD[x] == SQUARE {
            // Occupied
            if x < 0x10 {
                // Occupied by own piece (x=0-15)
                return MoveFlags{Negative: true, Overflow: false, Carry: false}
            }
            // Occupied by opponent piece (x=16-31)
            // Fall through to check-check
            goto IsCapture
        }
    }

    // Empty square
    goto EmptySquare

IsCapture:
    flags := MoveFlags{Negative: false, Overflow: true}

    // Should we do check-check?
    if STATE < 0 || STATE >= 8 {
        return flags  // Skip check-check
    }

    // Do expensive check verification
    flags.Carry = performCheckCheck()
    if flags.Carry {
        flags.Negative = true  // In check = illegal
    }
    return flags

EmptySquare:
    flags := MoveFlags{Negative: false, Overflow: false}

    // Should we do check-check?
    if STATE < 0 || STATE >= 8 {
        return flags
    }

    flags.Carry = performCheckCheck()
    if flags.Carry {
        flags.Negative = true
    }
    return flags
}

func performCheckCheck() bool {
    // Save state
    savedState := STATE
    savedInchek := INCHEK

    STATE = 0xF9      // Check detection mode
    INCHEK = 0xF9     // Will be set to 0 if king captured

    MOVE()            // Make trial move
    REVERSE()         // Switch sides
    GNM()             // Generate all opponent moves
    REVERSE()         // Switch back
    UMOVE()           // Unmake move

    STATE = savedState

    // If INCHEK==0, king was captured (we're in check)
    if INCHEK == 0 {
        INCHEK = savedInchek
        return true   // In check
    }

    INCHEK = savedInchek
    return false      // Not in check
}
```

## Critical Implementation Notes

1. **0x88 Board Representation**: The `AND #$88` check in CMOVE detects off-board squares instantly. In 0x88, valid squares have bit 7 and bit 3 both clear. Any move that leaves the board will set one of these bits.

2. **Piece Indexing**: Pieces 0-15 are the side to move, 16-31 are opponent. This enables the simple `CPX #$10` check in CMOVE to distinguish own vs opponent pieces.

3. **Move Direction Table**: MOVEX offsets are carefully chosen for 0x88 board. For example, $F0 (-16) moves down one rank, $01 moves right one file.

4. **State Machine**: The STATE variable controls search depth and type. GNM is recursively called with different STATE values to generate moves at different plies.

5. **Flag Usage**: CMOVE uses CPU flags as return values:
   - N (Negative): Move is illegal
   - V (Overflow): Move captures a piece
   - C (Carry): Move leaves king in check

6. **Pawn Special Case**: Pawns have unique movement (forward non-capture, diagonal capture only) and can move two squares from starting rank.

7. **Check-Check Overhead**: The CHKCHK routine is expensive (generates all opponent replies) so it's only invoked when STATE is in the range 0-7.

8. **JANUS Integration**: After each legal move, JANUS is called to either count it (mobility), evaluate it (capture analysis), or generate deeper variations (recursive search).

## Testing Strategy

To verify a Go reimplementation:

1. **Unit Test Each Piece Type**: Create board positions with isolated pieces and verify all generated moves match expected legal moves.

2. **Compare Move Counts**: For standard positions, compare total move counts (mobility) with original 6502 code output.

3. **Verify 0x88 Edge Detection**: Test pieces at board edges (files a/h, ranks 1/8) to ensure off-board detection works.

4. **Pawn Edge Cases**: Test pawn double-moves from rank 2, captures, and promotion squares.

5. **Check Detection**: Verify CHKCHK correctly identifies moves that leave king in check.

6. **State Machine**: Test GNM at different STATE values and verify correct counter updates.

## References

- Original source: Lines 280-352 (GNM), 407-469 (CMOVE), 162-234 (JANUS)
- Move table: Line 875-876 (MOVEX)
- Board setup: Line 870-873 (SETW)
- Piece values: Line 878-879 (POINTS)
