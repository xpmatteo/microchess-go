# MicroChess User Commands

This document describes the user interface and available commands in the 1976 MicroChess game.

## Overview

MicroChess operates through a simple text-based interface over a serial terminal connection (originally a KIM-1 with 6551 ACIA serial chip, later adapted for RS-232).

The game displays a chess board using ASCII characters and accepts single-key commands.

---

## Display Format

### Board Display Example

```
MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|WR WN WB WQ WK WB WN WR|70
|WP WP WP WP WP WP WP WP|60
| *  *  *  *  *  *  *  *|50
|  *  *  *  *  *  *  * *|40
| *  *  *  *  *  *  *  *|30
|  *  *  *  *  *  *  * *|20
|BP BP BP BP BP BP BP BP|10
|BR BN BB BQ BK BB BN BR|00
-------------------------
 00 01 02 03 04 05 06 07
34 56 78

```

### Display Elements

**Column Labels**: 00 01 02 03 04 05 06 07 (files a-h)

**Row Labels**: 70 60 50 40 30 20 10 00 (ranks 8-1 in hex)

**Pieces**:
- `W` = White piece
- `B` = Black piece
- `K` = King
- `Q` = Queen
- `R` = Rook
- `B` = Bishop (note: same as Black, context distinguishes)
- `N` = Knight
- `P` = Pawn

**Empty Squares**:
- `*` = Black square
- ` ` (space) = White square

**LED Values** (bottom line): Three hex bytes showing DIS1, DIS2, DIS3
- In original KIM-1: displayed on 7-segment LEDs
- In serial version: printed as hex values

---

## Available Commands

### C - Setup Board (line 116)

**Input**: Press 'C' (or 'c')

**Action**:
1. Loads initial chess position from SETW table
2. Copies 32 bytes into BOARD array
3. Sets OMOVE = $1B (activates opening book)
4. Displays "CCC" on LED values
5. Returns to main loop

**Board Setup**:
```
White pieces: R N B Q K B N R (rank 0)
White pawns:  P P P P P P P P (rank 1)
(empty ranks 2-5)
Black pawns:  P P P P P P P P (rank 6)
Black pieces: R N B Q K B N R (rank 7)
```

**Use Case**: Start a new game or reset the board

**Assembly Reference**: Lines 116-126

---

### E - Reverse Board (line 128)

**Input**: Press 'E' (or 'e')

**Action**:
1. Calls REVERSE routine
2. Swaps BOARD ↔ BK arrays
3. Transforms all coordinates: new = $77 - old
4. Toggles REV flag: REV = 1 - REV
5. Displays "EEE" on LED values
6. Returns to main loop

**Effect**:
- Flips board perspective (white on bottom ↔ black on bottom)
- Changes square $00 (a1) to $77 (h8), etc.
- Useful for playing from black's perspective

**REV Flag**:
- 0 = Normal (white on bottom)
- 1 = Reversed (black on bottom)

**Use Case**: Change which side is displayed at bottom of board

**Assembly Reference**: Lines 128-136

---

### P - Computer Play (line 138)

**Input**: Press 'P' (or 'p')

**Action**:
1. Calls GO routine (computer selects move)
2. GO checks opening book first
3. If out of book, generates and evaluates all moves
4. Selects best move (highest BESTV score)
5. Executes the move
6. Displays move on LED values
7. Returns to main loop

**Opening Book Behavior**:
- If still in book AND opponent's last move matches expected move
- Plays pre-programmed response
- Otherwise exits book (OMOVE = $FF) and starts thinking

**Thinking Indicator**:
- Prints "." for each move considered during analysis
- This provides feedback that computer is working

**Return Codes**:
- Returns to CHESS normally after move
- Returns $FF if no legal moves (checkmate or stalemate)

**Use Case**: Make the computer play the next move

**Assembly Reference**: Lines 138-144, 578-620

---

### Enter - Execute Player Move (line 146)

**Input**: Press Enter (CR, $0D)

**Action**:
1. Calls MOVE routine with current PIECE and SQUARE
2. MOVE executes the move entered via digit keys
3. Updates BOARD array
4. Displays updated board
5. Returns to main loop

**Prerequisites**:
- Must have entered a valid move using digit keys (see below)
- PIECE must contain piece index (set by first two digits)
- SQUARE must contain destination square (set by last two digits)

**Validation**:
- No validation at this point!
- Player responsible for entering legal moves
- Illegal moves may cause undefined behavior

**Use Case**: Execute a move after entering it with digit keys

**Assembly Reference**: Lines 146-149

---

### Q - Quit (line 150)

**Input**: Press 'Q' (or 'q')

**Action**:
1. Jumps to $FF00 (system monitor/OS)
2. Exits MicroChess program

**Note**: $FF00 is a placeholder - should be set to actual system entry point

**Use Case**: Exit the game and return to system

**Assembly Reference**: Lines 150-153

---

### 0-7 - Enter Move Digits (line 262)

**Input**: Press digits 0-7

**Action**:
1. Validates input is in range $00-$07
2. Calls DISMV to rotate digit into move accumulator
3. After four digits, SQUARE contains destination in 0x88 format

**Move Entry Sequence**:

1. **First digit** (0-7): From square file (column)
2. **Second digit** (0-7): From square rank (row)
   - These two digits identify the piece to move
   - Code searches BOARD array for piece at this square
   - Sets PIECE to the piece index found
3. **Third digit** (0-7): Destination file
4. **Fourth digit** (0-7): Destination rank
   - These two digits form the SQUARE (destination)
5. **Press Enter**: Execute the move

**Example - Moving White Pawn from e2 to e4**:
```
Initial position: White pawn at e2 ($14)
Desired move: e2-e4 ($34)

Press: 4  (file e = 4)
Press: 1  (rank 2 = 1)  → DIS2 = $14, finds pawn, PIECE set
Press: 4  (file e = 4)
Press: 3  (rank 4 = 3)  → SQUARE = $34
Press: Enter           → Move executed
```

**Display During Entry**:
- LED values show accumulated input
- After two digits: from square
- After four digits: to square

**Coordinate System**:
```
Files: 0=a, 1=b, 2=c, 3=d, 4=e, 5=f, 6=g, 7=h
Ranks: 0=1, 1=2, 2=3, 3=4, 4=5, 5=6, 6=7, 7=8
```

**Error Handling**:
- Digits > 7: Branch to ERROR, display board, no move made
- Invalid from square: Sets PIECE to invalid value, behavior undefined
- No validation of move legality - player must ensure move is legal

**Assembly Reference**: Lines 262-273, 625-633

---

## Typical Game Flow

### Starting a New Game

```
1. Press 'C'               → Board set to initial position
2. Press digit sequence    → Enter white's first move
3. Press Enter             → Execute move
4. Press 'P'               → Computer plays black's response
5. Repeat steps 2-4
```

### Playing as Black

```
1. Press 'C'               → Setup board
2. Press 'E'               → Reverse board (black on bottom)
3. Press 'P'               → Computer plays white's first move
4. Press digit sequence    → Enter black's response
5. Press Enter             → Execute move
6. Repeat steps 3-5
```

### Undoing Board Orientation

```
Press 'E' again            → Toggle back to original orientation
```

---

## Input Masking

All keyboard input is processed through KIN routine (line 812):

```assembly
KIN:
    LDA #"?"
    JSR SYSCHOUT    ; Print prompt
    JSR SYSKIN      ; Get character
    AND #$4F        ; Mask to 0-7 and A-Z
    RTS
```

**Masking Effect** (`AND #$4F`):
- Accepts: '0'-'7' ($30-$37 → $00-$07)
- Accepts: 'A'-'Z' ($41-$5A → $01-$1A)
- Accepts: 'a'-'z' ($61-$7A → $01-$1A)
- Result: Commands can be uppercase or lowercase

**Special Keys**:
- 'C'/'c' ($43/$63 → $43): Setup
- 'E'/'e' ($45/$65 → $45): Reverse
- 'P'/'p' ($50/$70 → $40): Play (note: lowercase 'p' becomes $40!)
- 'Q'/'q' ($51/$71 → $41): Quit
- Enter ($0D → $0D): Move
- '0'-'7' ($30-$37 → $00-$07): Digits

**Note**: The masking means 'p' ($70) becomes $40, which is why the code checks for `CMP #$40` (line 138) rather than 'P' ($50).

---

## Display States

### Normal Display

Shows current board position with pieces and empty squares.

### After 'C' Command

```
34 CC CC
```
- Displays "CCC" in LED values
- Indicates board setup complete

### After 'E' Command

```
34 EE EE
```
- Displays "EEE" in LED values
- Indicates board reversed

### During Move Entry

```
34 14 00
```
- Shows accumulated move input
- First byte: piece/from square
- Second byte: intermediate state
- Third byte: destination square

### During Computer Thinking

```
...............
```
- Prints "." for each move evaluated
- Provides progress feedback
- Can be many dots for complex positions

---

## Error Handling

**Invalid Digit** (> 7):
- Branches to ERROR (line 273)
- Jumps back to CHESS (displays board, continues)
- No move executed
- No error message (just returns to prompt)

**Invalid Move**:
- No validation in UI code
- Player must know chess rules
- Illegal moves may corrupt game state
- Historical limitation of 1KB program size

**No Legal Moves** (Checkmate/Stalemate):
- GO returns $FF (line 619)
- Program returns to system ($FF00)
- Game ends

---

## Implementation Notes

### KIM-1 Original Interface

The original 1976 KIM-1 version:
- Used 6-digit 7-segment LED display
- Used 23-key hexadecimal keypad
- Displayed move in hex (e.g., "14 34" for e2-e4)
- Required debouncing (OLDKY check, lines 112-114)

### Serial Terminal Adaptation (2002)

Daryl Rictor's modification:
- Replaced LED/keypad with serial I/O
- Added POUT routine for board display
- Added text-based piece representation
- Removed debouncing (serial input is inherently debounced)
- Added copyright banner

### Design Philosophy

The interface demonstrates extreme minimalism:
- Single-key commands (no complex parsing)
- No prompts (just "?" before each input)
- No error messages (fails silently)
- No move validation (trust the player)
- No algebraic notation (direct square entry)

This reflects the constraints of 1976:
- Limited memory (every byte counts)
- Limited display (6 LEDs)
- Limited input (hex keypad)
- No room for user-friendly features

The result is austere but functional - a complete chess program in ~1.5KB!

---

## Quick Reference Card

| Key | Command | Description |
|-----|---------|-------------|
| C | Setup | Load initial chess position |
| E | Reverse | Flip board perspective |
| P | Play | Computer makes a move |
| 0-7 | Digit | Enter move coordinate |
| Enter | Execute | Make the entered move |
| Q | Quit | Exit to system |

**Move Entry Format**: `[from_file][from_rank][to_file][to_rank][Enter]`

**Example**: `4134` + Enter = e2 to e4
