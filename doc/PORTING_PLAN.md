# MicroChess 6502 to Go Port - Implementation Plan

## Project Goal

Create a Go port that makes the historic 1976 MicroChess program easier to understand while preserving its exact chess logic and algorithms.

## Approach

**Hybrid style**: Faithful translation using Go idioms (structs, methods) to clarify the assembly's logic without changing behavior.

## User Requirements

Based on discussion with Captain Matt:

- **Port Style**: Hybrid approach - translate faithfully but use Go features where they make sense (e.g., structs instead of memory offsets)
- **Interface**: CLI text-based, similar to the original serial terminal interface
- **Chess Logic**: Preserve exactly - keep the exact same evaluation, opening book, and move generation to maintain historical accuracy
- **Primary Goal**: Making this historic program easier to understand - "It's quite astounding how much logic is packed in so very little code"

---

## Implementation Phases

**Key Principle**: Every phase produces a **runnable, demoable program** that shows tangible progress!

### ✅ Phase 1: Document the Original Assembly Program (COMPLETED)

Before writing any Go code, thoroughly document how the original MicroChess works.

**Deliverables**:
1. ✅ **doc/DATA_STRUCTURES.md** - Page zero variable map, board representation, move stack, evaluation counters
2. ✅ **doc/SUBROUTINES.md** - Major routines with descriptions, algorithms in pseudocode
3. ✅ **doc/COMMANDS.md** - User interface commands, move entry, display format
4. ✅ **doc/CALL_GRAPH.md** - Visual call graphs, state machine diagrams, flow charts

**Status**: Complete - All four documentation files created with comprehensive coverage

---

### ✅ Phase 1.5: 6502 Emulator Setup (BONUS!) - COMPLETED

**Goal**: Run the original 1976 assembly code with real I/O for validation

**What You Can Demo**: The actual original 6502 MicroChess running and playable!

**Achievement**: Successfully set up go6502 emulator with custom I/O support to run the original assembly code.

**Files Created**:
- `go6502/iomem.go` - Custom Memory implementation with console I/O
- `go6502/testrun.go` - Simple harness to run 6502 programs
- `go6502/microchess.asm` - Modified to use memory-mapped I/O ($FFF0/$FFF1)
- `go6502/microchess.bin` - Assembled binary (1.4KB!)
- `go6502/RUNNING_MICROCHESS.md` - Complete documentation

**How to Run**:
```bash
cd go6502
go run testrun.go iomem.go microchess.bin
# Type 'C' to setup, 'P' for computer move
```

**Technical Achievement**:
- Implemented Memory interface with I/O hooks at $FFF0 (output) and $FFF1 (input)
- Modified original ACIA serial routines to use simple memory-mapped I/O
- Minimal changes to assembly (just 3 simple routines)
- Full console I/O working - can see board, enter moves, play games!

**Why This Matters**:
- **Validation**: Can now compare Go port directly to running original
- **Reference**: Can trace execution, examine state, debug differences
- **Historical**: Preserves ability to run original code exactly as written
- **Testing**: Feed identical inputs to both versions, verify identical behavior

**Status**: Complete - Original 1976 MicroChess is now runnable!

---

### ✅ Phase 2: Board Display (First Demo!) - COMPLETED

**Goal**: Create a program that displays a chess board

**What You Can Demo**: Run the program and see a pretty chess board!

**Tasks**:
- [x] Create Go module structure
- [x] Implement basic types:
   ```go
   type Square uint8
   type Piece uint8
   ```
- [x] Implement board state with initial position
- [x] Create CLI that displays the board (like POUT)
- [x] Add 'Q' command to quit

**Demo Commands**:
```bash
go run cmd/microchess/main.go
# Shows initial chess position
# Press Q to quit
```

**Output** (matching original 6502 assembly behavior):
```
MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|BP|**|  |**|  |**|  |**|00
|**|  |**|  |**|  |**|  |10
|  |**|  |**|  |**|  |**|20
|**|  |**|  |**|  |**|  |30
|  |**|  |**|  |**|  |**|40
|**|  |**|  |**|  |**|  |50
|  |**|  |**|  |**|  |**|60
|**|  |**|  |**|  |**|  |70
-------------------------
 00 01 02 03 04 05 06 07
00 00 00

? _
```

Note: The original starts with an "uninitialized" board showing just one black pawn. The 'C' command is needed to set up pieces.

**Deliverables**:
- [x] Working Go module
- [x] Basic types (Square, Piece, GameState)
- [x] Board display function
- [x] Simple REPL loop
- [x] README with how to run
- [x] **Can demo**: Shows a chess board!

---

### ✅ Phase 3: Board Setup and Reset (Interactive Demo!) - COMPLETED

**Goal**: Add 'C' command to reset board, 'E' command to flip it

**What You Can Demo**: Interact with the board - reset it, flip perspective!

**Tasks**:
- [x] Implement REVERSE routine
- [x] Implement board setup from InitialSetup
- [x] Add 'C' command (setup)
- [x] Add 'E' command (reverse)
- [x] Track reversed state

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup board (shows "CC CC CC")
? E              # Reverse board (shows "EE EE EE")
? E              # Reverse back
? Q              # Quit
```

**What It Shows**:
- ✅ Board resets to initial position
- ✅ LED display shows "CC CC CC" after setup
- ✅ Board flips between white/black perspective
- ✅ Coordinate transformation working (0x77 - square)
- ✅ REV flag tracked in GameState
- ✅ Double reverse restores original position

**Deliverables**:
- [x] REVERSE implementation (pkg/microchess/types.go:160)
- [x] Setup board function (complete)
- [x] Command dispatcher (C, E, Q all implemented)
- [x] Unit tests (TestReverse, TestHandleCommandE)
- [x] Acceptance tests (TestReverseCommand, TestSetupReverseAndQuitSequence)
- [x] **Can demo**: Interactive board manipulation works perfectly!

**Status**: Complete - All functionality implemented and tested

---

### ⬜ Phase 4: Move Entry and Display (Manual Chess!)

**Goal**: Enter and display moves (no validation yet)

**What You Can Demo**: Play chess manually - you enforce the rules, program just moves pieces!

**Tasks**:
1. Implement move entry (0-7 digit input)
2. Implement DISMV (build move from digits)
3. Implement basic MOVE (update board, no capture handling yet)
4. Show entered move in LED display area
5. Find piece at "from" square

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup board
? 4              # e-file
? 1              # rank 2 (e2)
? 4              # e-file
? 3              # rank 4 (e4)
? [Enter]        # Execute move
# Board updates, pawn now on e4!
? Q
```

**What It Shows**:
- Move entry working
- Pieces moving on board
- Display updating
- No validation - you can make illegal moves!
- LED area shows move (14 34)

**Deliverables**:
- INPUT/DISMV implementation
- Basic MOVE implementation
- Move execution (no undo yet)
- **Can demo**: Manual chess game!

---

### ⬜ Phase 5: Legal Move Generation (Show All Moves!)

**Goal**: Generate and display all legal moves for a position

**What You Can Demo**: Enter a position, program shows all legal moves!

**Tasks**:
1. Implement CMOVE (move calculation and validation)
2. Implement piece-specific move generators
3. Implement GNM (generate all moves)
4. Add 'L' command (list moves) - NEW, not in original!
5. Add move validation (reject illegal moves)

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup board
? L              # List all legal moves
# Prints: e2-e3, e2-e4, d2-d3, d2-d4, ... (20 moves)
? 4 1 4 3        # e2-e4
? [Enter]        # Legal move - executes
? 4 3 4 5        # Try e4-e6 (illegal!)
# Error: Illegal move
? Q
```

**What It Shows**:
- Complete move generation working
- All piece types handled
- Board edge detection (0x88)
- Collision detection
- Illegal move rejection

**Deliverables**:
- CMOVE implementation
- All piece move generators
- GNM implementation
- Move validation
- 'L' command for debugging
- **Can demo**: Shows legal moves, rejects illegal ones!

---

### ⬜ Phase 6: Undo Moves (Time Travel Chess!)

**Goal**: Add move undo capability

**What You Can Demo**: Make moves, undo them, try different variations!

**Tasks**:
1. Implement move stack (replaces dual stack)
2. Implement UMOVE (undo move)
3. Complete MOVE with capture handling
4. Add 'U' command (undo) - NEW, not in original!
5. Track captured pieces

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup
? 4 1 4 3        # e2-e4
? [Enter]
? 3 6 3 5        # d7-d6
? [Enter]
? U              # Undo d7-d6
# Board reverts!
? U              # Undo e2-e4
# Back to start
? Q
```

**What It Shows**:
- Move history tracking
- Perfect undo functionality
- Captured pieces restored
- Can explore variations

**Deliverables**:
- Move stack implementation
- Complete MOVE with captures
- UMOVE implementation
- 'U' command
- **Can demo**: Time travel through game!

---

### ⬜ Phase 7: Position Evaluation (Show Me the Score!)

**Goal**: Evaluate positions and show scores

**What You Can Demo**: Make moves, see position evaluation change!

**Tasks**:
1. Implement COUNTS (mobility counting)
2. Implement STRATGY (position evaluation)
3. Add 'S' command (show score) - NEW!
4. Display evaluation after each move
5. Show evaluation breakdown

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup
? S              # Show score
# Score: 128 (equal position)
# Mobility: W=20 B=20
# Material: Equal
? 4 1 4 3        # e2-e4
? [Enter]
? S
# Score: 132 (slight advantage to white)
# Mobility: W=20 B=20
# Center control: +2
? Q
```

**What It Shows**:
- Position evaluation working
- Mobility counting
- Positional bonuses
- Material counting
- Score changes with moves

**Deliverables**:
- COUNTS implementation
- STRATGY implementation
- Evaluation display
- 'S' command
- **Can demo**: See position scores!

---

### ⬜ Phase 8: Single-Ply Search (Computer Suggests Move!)

**Goal**: Computer suggests best move (1-ply search)

**What You Can Demo**: Computer analyzes position and suggests a move!

**Tasks**:
1. Implement JANUS (state machine) - simplified for now
2. Implement PUSH (compare moves)
3. Implement search at depth 1
4. Add 'H' command (hint) - NEW!
5. Show thinking progress (dots)

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup
? H              # Get hint
# Thinking.....................
# Best move: e2-e4 (score: 132)
? 4 1 4 3        # Play the hint
? [Enter]
? H              # Hint for black
# Thinking.....................
# Best move: e7-e5 (score: 130)
? Q
```

**What It Shows**:
- Move evaluation working
- Best move selection
- Thinking progress indicator
- Computer "intelligence"

**Deliverables**:
- JANUS state machine (simplified)
- PUSH implementation
- Single-ply search
- 'H' command
- **Can demo**: Computer suggests moves!

---

### ⬜ Phase 9: Full Computer Player (Computer Plays!)

**Goal**: Computer can play a full game

**What You Can Demo**: Play against the computer!

**Tasks**:
1. Implement full JANUS with all states
2. Implement TREE (capture analysis)
3. Implement CHKCHK (check detection)
4. Implement GO (computer move selection)
5. Implement opening book
6. Restore 'P' command (computer play)

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup
? P              # Computer plays
# Thinking.....
# Computer plays: e2-e4
? 4 6 4 5        # e7-e6 (you play)
? [Enter]
? P              # Computer responds
# Thinking.....
# Computer plays: d2-d4
# ... continue game ...
? Q
```

**What It Shows**:
- Full chess AI working
- Opening book
- Multi-ply search
- Capture analysis
- Check detection
- Complete game play

**Deliverables**:
- Complete JANUS state machine
- TREE implementation
- CHKCHK implementation
- GO implementation
- Opening book
- 'P' command restored
- **Can demo**: Full chess game vs computer!

---

### ⬜ Phase 10: Polish and Performance (Fast Computer!)

**Goal**: Optimize and improve user experience

**What You Can Demo**: Computer thinks faster, interface is smoother!

**Tasks**:
1. Profile and optimize hot paths
2. Add game state display (whose turn, check status)
3. Add move history display
4. Improve progress indicators
5. Add game save/load
6. Clean up output formatting

**Demo Commands**:
```bash
go run cmd/microchess/main.go
? C              # Setup
? P              # Computer plays (faster now!)
# Thinking... (3 seconds instead of 10)
# Computer plays: e2-e4
? M              # Show move history - NEW!
# 1. e2-e4
? G              # Save game - NEW!
# Game saved to microchess.save
? Q
```

**What It Shows**:
- Improved performance
- Better UX
- Game persistence
- Move history
- Professional polish

**Deliverables**:
- Performance optimizations
- Enhanced UI
- Game save/load
- Move history
- Status display
- **Can demo**: Production-ready chess game!

---

### ⬜ Phase 11: Testing and Validation

**Goal**: Verify correctness against original

**What You Can Demo**: Side-by-side comparison with original running in emulator!

**BREAKTHROUGH**: We can now run the original 1976 6502 assembly code with real I/O in the go6502 emulator! See `go6502/RUNNING_MICROCHESS.md` for details.

**Running the Original**:
```bash
cd go6502
go run testrun.go iomem.go microchess.bin
# Type 'C' to set up board, 'P' for computer move
```

**How It Works**:
- Custom Memory implementation (`iomem.go`) with I/O at $FFF0 (output) and $FFF1 (input)
- Modified MicroChess to use memory-mapped I/O instead of ACIA serial port
- Simple test harness (`testrun.go`) runs 6502 programs with console I/O

**Validation Tasks**:
1. Set up identical positions in both original and Go port
2. Compare move generation with original emulated code
3. Compare evaluation scores
4. Compare move selection
5. Feed identical inputs to both versions
6. Verify identical outputs
7. Play test games side-by-side
8. Document any differences

**Demo**:
```bash
# Run the original 6502 code
cd go6502
echo "C" | go run testrun.go iomem.go microchess.bin

# Run validation suite
cd ..
go test ./... -v

# Run comparison tool
go run cmd/validate/main.go
# Compare move generation:
# Position: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
# Original (6502): 20 moves
# Go port:         20 moves ✓
# Evaluation:
# Original (6502): 128
# Go port:         128 ✓
```

**Deliverables**:
- Comprehensive test suite
- Validation tool comparing Go port to running 6502 emulator
- Comparison reports
- Bug fixes
- **Can demo**: Proof of correctness against actual running 6502 code!

---

### ⬜ Phase 12: Documentation and Examples

**Goal**: Complete documentation and examples

**What You Can Demo**: Show example games, tutorial mode!

**Tasks**:
1. Add example games
2. Create tutorial mode
3. Complete README with screenshots
4. Add architecture docs
5. Document differences from original
6. Create demo videos/GIFs

**Demo**:
```bash
# Play example game
go run cmd/microchess/main.go --replay examples/famous_game.txt

# Tutorial mode
go run cmd/microchess/main.go --tutorial
# This position is called the "Scandinavian Defense"
# Black has played d7-d5 to challenge white's e4 pawn...

# Generate README screenshots
go run cmd/microchess/main.go --screenshot > docs/board.txt
```

**Deliverables**:
- Complete documentation
- Example games
- Tutorial mode
- Architecture docs
- Demo materials
- **Can demo**: Complete project showcase!

**File: pkg/board/board.go**

Implement:
```go
type Square uint8  // 0x88 representation

// Board square encoding constants
const (
    SquareA1 Square = 0x00
    SquareH8 Square = 0x77
    // ... etc
)

// Helper functions
func (s Square) IsValid() bool        // Check if square is on board
func (s Square) String() string       // e.g., "e4"
func (s Square) Rank() int            // 0-7
func (s Square) File() int            // 0-7
func ParseSquare(s string) Square     // "e4" -> 0x34
```

**File: pkg/microchess/types.go**

Implement:
```go
type Piece uint8  // Index 0-15 (white), 16-31 (black) - but we work with one side

const (
    PieceKing Piece = iota  // 0
    PieceQueen              // 1
    PieceRook1              // 2
    PieceRook2              // 3
    PieceBishop1            // 4
    PieceBishop2            // 5
    PieceKnight1            // 6
    PieceKnight2            // 7
    PiecePawn1              // 8
    PiecePawn2              // 9
    // ... through PiecePawn8 = 15
)

type State int8   // Analysis state: 12, 4, 0, 8, -1 to -5, -7

const (
    StateCandidateGeneration State = 12
    StateFullAnalysis        State = 4
    StateImmediateReply      State = 0
    StateContinuation        State = 8
    StateCheckDetection      State = -7
    // Capture analysis: -1 to -5
)

type MoveRecord struct {
    TargetSquare   Square
    CapturedPiece  Piece  // 0xFF if none
    OriginalSquare Square
    Piece          Piece
    MoveIndex      uint8
}

type MoveResult struct {
    Legal   bool  // false if off board or own piece
    Capture bool  // true if opponent piece on target
    InCheck bool  // true if move results in check (only if checked)
}

type GameState struct {
    // Board representation
    Board     [16]Square  // Current position of each piece
    BK        [16]Square  // Alternate board for REVERSE

    // Move generation state
    Piece     Piece
    Square    Square
    MoveN     uint8
    State     State
    InCheck   bool
    Reversed  bool

    // Move stack
    MoveStack []MoveRecord

    // Evaluation counters (indexed by state offset)
    Mobility      [16]uint8
    MaxCapture    [16]uint8
    CaptureCount  [16]uint8
    PieceCaptured [16]Piece

    // Specific counter instances
    WMOB, WMAXC, WCC   uint8
    WMAXP              Piece
    BMOB, BMAXC, BMCC  uint8
    BMAXP              Piece
    PMOB, PMAXC, PCC   uint8
    PCP                Piece

    // Capture depth counters
    WCAP0, WCAP1, WCAP2 uint8
    BCAP0, BCAP1, BCAP2 uint8
    XMAXC               uint8

    // Best move tracking
    BestPiece  Piece
    BestValue  uint8
    BestSquare Square

    // Opening book
    OpeningMove int8  // -1 if out of book

    // Display (for move entry)
    DIS1, DIS2, DIS3 uint8
}

// Constants
var MoveOffsets = [17]int8{
    0x00,        // 0: unused
    -0x10,       // 1: up 1 rank
    -0x01,       // 2: left 1 file
    0x01,        // 3: right 1 file
    0x10,        // 4: down 1 rank
    0x11,        // 5: down-right diagonal
    0x0F,        // 6: down-left diagonal
    -0x11,       // 7: up-left diagonal
    -0x0F,       // 8: up-right diagonal
    -0x21,       // 9: knight moves...
    -0x12, -0x0E, 0x0E, 0x12, 0x21, 0x1F, -0x1F,  // 10-16: knight moves
}

var PieceValues = [16]uint8{
    11,      // King (special)
    10,      // Queen
    6, 6,    // Rooks
    4, 4,    // Bishops
    4, 4,    // Knights
    2, 2, 2, 2, 2, 2, 2, 2,  // Pawns
}

var InitialSetup = [32]Square{
    // White pieces (indices 0-7)
    0x03, 0x04, 0x00, 0x07, 0x02, 0x05, 0x01, 0x06,
    // White pawns (indices 8-15)
    0x10, 0x17, 0x11, 0x16, 0x12, 0x15, 0x14, 0x13,
    // Black pieces (indices 16-23)
    0x73, 0x74, 0x70, 0x77, 0x72, 0x75, 0x71, 0x76,
    // Black pawns (indices 24-31)
    0x60, 0x67, 0x61, 0x66, 0x62, 0x65, 0x64, 0x63,
}

var OpeningBook = []uint8{
    0x99, 0x25, 0x0B, 0x25, 0x01, 0x00, 0x33, 0x25,
    0x07, 0x36, 0x34, 0x0D, 0x34, 0x34, 0x0E, 0x52,
    0x25, 0x0D, 0x45, 0x35, 0x04, 0x55, 0x22, 0x06,
    0x43, 0x33, 0x0F, 0xCC,
}
```

**Tests**:
- Test 0x88 square validation
- Test square string conversion
- Test initial board setup
- Test constants are correct

**Deliverables**:
- board.go with Square type and helpers
- types.go with all core types and constants
- Comprehensive unit tests
- Comments mapping to assembly line numbers

---

### Phase 4: Move Generation (Core Algorithm)

Port the move generation routines - the heart of the chess engine.

**File: pkg/microchess/moves.go**

Implement (in order):
1. `func (g *GameState) CalculateMove(square Square, moveIdx uint8) MoveResult`
   - Port of CMOVE (line 407)
   - Calculate newSquare = square + MoveOffsets[moveIdx]
   - Check 0x88 boundaries
   - Detect collisions and captures
   - Return MoveResult{Legal, Capture, InCheck}
   - Add comment: "// CMOVE (line 407): Calculate and validate move"

2. `func (g *GameState) ResetPiecePosition()`
   - Port of RESET (line 473)
   - Load piece's current square from Board
   - "// RESET (line 473): Restore piece to current position"

3. `func (g *GameState) GenerateSingleMove() bool`
   - Port of SNGMV (line 357)
   - Call CalculateMove
   - If legal, call analyzeMove
   - Decrement MoveN
   - "// SNGMV (line 357): Generate single-step moves"

4. `func (g *GameState) GenerateLine() bool`
   - Port of LINE (line 367)
   - Sliding moves until blocked
   - "// LINE (line 367): Generate sliding moves"

5. Piece-specific generators:
   - `func (g *GameState) generateKingMoves()`    // Line 306
   - `func (g *GameState) generateQueenMoves()`   // Line 309
   - `func (g *GameState) generateRookMoves()`    // Line 313
   - `func (g *GameState) generateBishopMoves()`  // Line 319
   - `func (g *GameState) generateKnightMoves()`  // Line 325
   - `func (g *GameState) generatePawnMoves()`    // Line 333

6. `func (g *GameState) GenerateAllMoves()`
   - Port of GNM (line 286)
   - Loop through pieces 15 down to 0
   - Dispatch to piece-specific handlers
   - "// GNM (line 286): Generate all moves for current side"

7. `func (g *GameState) ClearCounters()`
   - Port of GNMZ (line 280)
   - Clear evaluation counters
   - Call GenerateAllMoves
   - "// GNMZ (line 280): Clear counters and generate moves"

**Tests**:
- Test CalculateMove for all edge cases
- Test each piece type's move generation
- Verify move counts for known positions
- Test 0x88 boundary detection
- Test capture detection

**Deliverables**:
- moves.go with all move generation logic
- Extensive unit tests with known positions
- Comments explaining assembly equivalents

---

### Phase 5: Move Execution

Port the move make/unmake system and board manipulation.

**File: pkg/microchess/execute.go**

Implement:
1. `func (g *GameState) MakeMove(piece Piece, target Square)`
   - Port of MOVE (line 511)
   - Push MoveRecord to stack
   - Update board
   - Mark captured pieces
   - "// MOVE (line 511): Execute move and save to stack"

2. `func (g *GameState) UnmakeMove()`
   - Port of UMOVE (line 488)
   - Pop MoveRecord from stack
   - Restore board state
   - Restore captured pieces
   - "// UMOVE (line 488): Undo last move from stack"

3. `func (g *GameState) ReverseBoard()`
   - Port of REVERSE (line 382)
   - Swap Board ↔ BK
   - Transform all coordinates: new = 0x77 - old
   - "// REVERSE (line 382): Flip board perspective"

4. `func (g *GameState) GenerateReplies()`
   - Port of GENRM (line 480)
   - MakeMove, Reverse, GenerateAllMoves, Reverse, UnmakeMove
   - "// GENRM (line 480): Generate opponent replies"

**Tests**:
- Test MakeMove/UnmakeMove restore state
- Test multiple move sequences
- Test ReverseBoard coordinate transformation
- Test captured piece handling
- Verify stack grows/shrinks correctly

**Deliverables**:
- execute.go with move execution logic
- Unit tests verifying exact state restoration
- Stack manipulation tests

---

### Phase 6: Position Evaluation

Port the evaluation and counting systems.

**File: pkg/eval/evaluate.go**

Implement:
1. `func (g *GameState) CountMobility()`
   - Port of COUNTS (line 169)
   - Count legal moves
   - Track maximum captures
   - Queens count double
   - "// COUNTS (line 169): Count mobility and captures"

2. `func (g *GameState) AnalyzeExchanges()`
   - Port of TREE (line 240)
   - Recursive capture analysis
   - Decrement State for deeper analysis
   - "// TREE (line 240): Analyze capture sequences"

3. `func (g *GameState) EvaluatePosition() uint8`
   - Port of STRATGY (line 641)
   - Weighted evaluation formula
   - Positional bonuses for center squares
   - Return score 0-255
   - "// STRATGY (line 641): Evaluate position and return score"
   - **CRITICAL**: Preserve exact weights and formula

4. `func (g *GameState) CheckForMate() uint8`
   - Port of CKMATE (line 545)
   - Detect check, checkmate, stalemate
   - Return 0x00 (illegal), 0xFF (mate), or continue
   - "// CKMATE (line 545): Check for checkmate/stalemate"

5. `func (g *GameState) CompareAndSaveBest(score uint8)`
   - Port of PUSH (line 564)
   - Compare with BestValue
   - Save if better
   - Print "." for progress
   - "// PUSH (line 564): Save move if best so far"

**Tests**:
- Test evaluation on known positions
- Verify mobility counting
- Test checkmate detection
- Compare scores to expected values
- Test center square bonuses

**Deliverables**:
- evaluate.go with evaluation logic
- Tests with known position scores
- Documentation of evaluation formula

---

### Phase 7: Search & Analysis

Port the JANUS state machine and search logic.

**File: pkg/microchess/search.go**

Implement:
1. `func (g *GameState) AnalyzeMove()`
   - Port of JANUS (line 162)
   - State machine based on State value
   - Route to COUNTS, TREE, or check detection
   - "// JANUS (line 162): Analysis director"

2. `func (g *GameState) CheckCheck() bool`
   - Port of CHKCHK (line 444)
   - Save state, set State = -7
   - MakeMove, Reverse, GenerateAllMoves
   - Detect if king capturable
   - Restore state
   - "// CHKCHK (line 444): Verify move doesn't leave king in check"
   - **WARNING**: Mark as expensive operation

**State Machine Logic**:
```go
switch g.State {
case StateCandidateGeneration:  // 12
    // Skip COUNTS
case StateFullAnalysis:  // 4
    g.CountMobility()
    // Special ON4 handling
case StateImmediateReply:  // 0
    g.CountMobility()
case StateContinuation:  // 8
    g.CountMobility()
case StateCheckDetection:  // -7
    // Check if BK[0] capturable
default:  // -1 to -5
    g.AnalyzeExchanges()
}
```

**Tests**:
- Test state machine routing
- Test check detection
- Verify state transitions
- Test with various State values

**Deliverables**:
- search.go with JANUS state machine
- CHKCHK implementation
- Tests for state transitions

---

### Phase 8: Computer Player

Port the move selection and opening book.

**File: pkg/microchess/computer.go**

Implement:
1. `func (g *GameState) SelectBestMove() (Piece, Square, error)`
   - Port of GO (line 578)
   - Check opening book first
   - If out of book, generate and analyze
   - Return best move
   - "// GO (line 578): Computer move selection"

2. Opening book logic:
   - Check if opponent move matches expected
   - Play pre-programmed response
   - Exit book if mismatch (OpeningMove = -1)

3. Search logic:
   - State = 12, generate candidates
   - State = 4, full analysis
   - Select highest BestValue
   - Return error if no legal moves

**Tests**:
- Test opening book sequences
- Test book exit on mismatch
- Test move selection from various positions
- Verify "thinking" progress indicators

**Deliverables**:
- computer.go with AI logic
- Opening book implementation
- Tests with known opening sequences

---

### Phase 9: CLI Interface

Create the text-based user interface.

**File: cmd/microchess/main.go**

Implement:
1. Board display (port of POUT, line 702):
   ```
   MicroChess (c) 1976-2005 Peter Jennings
   00 01 02 03 04 05 06 07
   -------------------------
   |WR WN WB WQ WK WB WN WR|70
   ...
   ```

2. Command loop:
   - 'C' - Setup board
   - 'E' - Reverse board
   - 'P' - Computer play
   - Enter - Execute player move
   - 'Q' - Quit
   - 0-7 - Enter move digits

3. Move input parser:
   - Build move from 4 digits
   - Find piece at from square
   - Validate and execute

4. Display helpers:
   - Print board state
   - Print piece characters
   - Print LED values (move display)
   - Print progress dots during thinking

**Tests**:
- Test board display formatting
- Test command parsing
- Test move input parsing
- Integration test: play a game

**Deliverables**:
- main.go with complete CLI
- User-friendly interface
- Help text and instructions

---

### Phase 10: Testing & Validation

Comprehensive testing to verify correctness.

**Tasks**:
1. **Unit Tests**: Achieve high coverage
   - Each routine tested independently
   - Edge cases covered
   - Boundary conditions tested

2. **Integration Tests**:
   - Play through known games
   - Verify opening book sequences
   - Test checkmate scenarios
   - Test stalemate scenarios

3. **Validation Against Original**:
   - Compare move generation for positions
   - Compare evaluation scores
   - Verify same moves selected
   - Test identical behavior

4. **Performance Testing**:
   - Measure move generation speed
   - Measure evaluation speed
   - Profile for bottlenecks
   - (Not critical, but interesting)

5. **Edge Case Testing**:
   - Capture sequences
   - Check detection
   - Pawn double moves
   - Piece collision detection

**Deliverables**:
- Comprehensive test suite
- Validation report comparing to original
- Performance metrics
- Bug fixes as needed

---

### Phase 11: Final Documentation

Complete the project documentation.

**Tasks**:
1. **Code Comments**:
   - Every function has assembly line reference
   - Complex algorithms explained with "why"
   - Historical context noted where relevant
   - Examples in comments

2. **README.md Enhancements**:
   - Complete "How to Play" section
   - Example game session
   - Build/install instructions
   - Project history and motivation

3. **ARCHITECTURE.md** (New):
   - Go package structure
   - Design decisions explained
   - Mapping between assembly and Go
   - State machine documentation
   - Testing strategy

4. **DIFFERENCES.md** (New):
   - Document any differences from original
   - Explain why differences exist
   - Note any bugs found/fixed
   - Modernizations made

5. **Examples**:
   - Create example games in examples/
   - Document interesting positions
   - Tutorial for using as a library

**Deliverables**:
- Comprehensive code comments
- Complete README
- Architecture documentation
- Difference documentation
- Example games

---

## File Structure

Final structure:
```
microchess-go/
├── cmd/
│   └── microchess/
│       └── main.go                    # CLI entry point
├── pkg/
│   ├── board/
│   │   ├── board.go                   # Board representation
│   │   └── board_test.go              # Board tests
│   ├── eval/
│   │   ├── evaluate.go                # Position evaluation
│   │   └── evaluate_test.go           # Evaluation tests
│   └── microchess/
│       ├── types.go                   # Core types and constants
│       ├── moves.go                   # Move generation
│       ├── moves_test.go              # Move generation tests
│       ├── execute.go                 # Move execution
│       ├── execute_test.go            # Move execution tests
│       ├── search.go                  # Search and JANUS
│       ├── search_test.go             # Search tests
│       ├── computer.go                # Computer player
│       └── computer_test.go           # Computer player tests
├── examples/
│   ├── sample_game.txt                # Example game
│   └── interesting_positions.txt      # Notable positions
├── doc/
│   ├── Microchess6502.txt             # Original assembly (preserved)
│   ├── 6502-instruction-set.md        # Reference (preserved)
│   ├── DATA_STRUCTURES.md             # ✅ Phase 1
│   ├── SUBROUTINES.md                 # ✅ Phase 1
│   ├── COMMANDS.md                    # ✅ Phase 1
│   ├── CALL_GRAPH.md                  # ✅ Phase 1
│   ├── PORTING_PLAN.md                # This file
│   ├── ARCHITECTURE.md                # Phase 11
│   └── DIFFERENCES.md                 # Phase 11
├── CLAUDE.md                          # Updated throughout
├── README.md                          # Phase 2, enhanced Phase 11
├── go.mod                             # Phase 2
└── go.sum                             # Generated
```

---

## Key Preservation Decisions

These must be maintained exactly:

1. **0x88 Board Encoding**
   - Square = (rank << 4) | file
   - Edge detection: square & 0x88 != 0
   - Elegant and proven - keep it!

2. **Exact Evaluation Weights**
   - STRATGY formula weights define chess personality
   - Must match assembly exactly:
     - 0.25× mobility difference
     - 0.50× tactical advantage
     - 1.00× material/threats
     - +2 for center/development

3. **STATE Machine Values**
   - 12, 4, 0, 8, -1 to -5, -7
   - These control search depth
   - Critical to preserve

4. **Piece Indexing**
   - 0-7: pieces (K, Q, R, R, B, B, N, N)
   - 8-15: pawns
   - Used throughout for piece identification

5. **Move Offset Table**
   - MoveOffsets array values exact
   - Used for all move generation
   - Battle-tested values

6. **Opening Book**
   - Exact sequence from OPNING table
   - Matching logic for book exit
   - Historical opening choices

---

## Modernizations

These improve understandability:

1. **Dual Stack → Go Slice**
   - Replace SP1/SP2 mechanism
   - Use []MoveRecord stack
   - Push/Pop methods clearer than assembly

2. **CPU Flags → Return Struct**
   - N/V/C flags → MoveResult{Legal, Capture, InCheck}
   - Explicit is better than implicit

3. **Global State → Methods**
   - Assembly uses page zero globals
   - Go uses methods on GameState
   - Clearer ownership and scope

4. **Extensive Comments**
   - Explain "why" not just "what"
   - Reference assembly line numbers
   - Historical context

5. **Unit Tests**
   - Verify each routine independently
   - Catch regressions
   - Document expected behavior

6. **Type Safety**
   - Square, Piece, State types
   - Compiler catches errors
   - Self-documenting code

---

## Success Criteria

The port is successful when:

1. [x] All documentation complete and accurate
2. [ ] All unit tests pass
3. [ ] Move generation matches assembly output
4. [ ] Evaluation scores match for test positions
5. [ ] Opening book plays identical moves
6. [ ] Computer selects same moves as original
7. [ ] CLI interface is functional and user-friendly
8. [ ] Code is well-commented and understandable
9. [ ] README provides complete usage guide
10. [ ] Project achieves goal: makes historic program understandable

---

## Development Workflow

For each phase:

1. **Read Documentation**: Review relevant .md files
2. **Write Tests First**: TDD approach where practical
3. **Implement**: Port assembly to Go
4. **Comment**: Add assembly line references
5. **Test**: Verify correctness
6. **Document**: Update docs as needed
7. **Review**: Check preservation vs modernization
8. **Commit**: Small, focused commits

---

## Notes on Historical Accuracy

When porting, remember:

- This is **1976 code** for a **1 MHz CPU** with **~1KB RAM**
- Every byte counted - hence the clever tricks
- Some patterns seem odd by modern standards
- The constraints drove brilliant solutions
- Preserve the cleverness while explaining it

The goal is not to make a "better" chess engine, but to make **this specific historic implementation** accessible to modern developers.

---

## Resources

- **Original Source**: doc/Microchess6502.txt
- **Data Structures**: doc/DATA_STRUCTURES.md
- **Subroutines**: doc/SUBROUTINES.md
- **Commands**: doc/COMMANDS.md
- **Call Graphs**: doc/CALL_GRAPH.md
- **6502 Reference**: doc/6502-instruction-set.md
- **Running Emulator**: go6502/RUNNING_MICROCHESS.md
- **Emulator Code**: go6502/ directory (go6502 emulator with I/O support)

---

## Contact & Acknowledgments

**Original Author**: Peter Jennings (1976)
**Serial Port Adaptation**: Daryl Rictor (2002)
**OCR Corrections**: Bill Forster (2005)
**Go Port**: [Your name] (2025)

**Project Goal**: Preserve computing history by making this remarkable 1976 chess program accessible to modern developers.

---

*This plan was created on 2025-11-08 as part of the MicroChess preservation and education project.*
