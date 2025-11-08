# MicroChess Call Graphs and Flow Diagrams

This document provides visual representations of the call relationships, control flow, and state machines in MicroChess.

---

## Main Program Flow

```
┌─────────────────────────────────────────────┐
│ CHESS (line 100)                            │
│ Main initialization and game loop           │
└──────────────────┬──────────────────────────┘
                   │
                   ├─ Initialize stacks
                   │  (SP=$FF, SP2=$C8)
                   │
                   v
         ┌─────────────────────┐
         │   Main Loop         │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ OUT (display/input) │
         └─────────┬───────────┘
                   │
                   ├─→ POUT (display board)
                   └─→ KIN (get key)
                   │
                   v
         ┌─────────────────────┐
         │ Command Dispatch    │
         └─────────┬───────────┘
                   │
                   ├─ 'C' → WHSET (setup board)
                   ├─ 'E' → REVERSE (flip board)
                   ├─ 'P' → GO (computer play)
                   ├─ Enter → MOVE (execute move)
                   ├─ 'Q' → DONE (exit)
                   └─ 0-7 → INPUT/DISMV (enter move)
                   │
                   └─→ Loop back to OUT
```

---

## Move Generation Call Graph

```
┌─────────────────────────────────────────────────────────┐
│ GNM (Generate All Moves)                                │
│ Loop through all 16 pieces                              │
└────────────────────────┬────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          v              v              v
    ┌─────────┐    ┌─────────┐    ┌──────────┐
    │  RESET  │    │ Dispatch│    │  Piece   │
    │ (restore│    │by piece │    │ handlers │
    │ square) │    │  type)  │    └────┬─────┘
    └─────────┘    └─────────┘         │
                                        │
         ┌──────────┬─────────┬─────────┼────────┬─────────┐
         │          │         │         │        │         │
         v          v         v         v        v         v
    ┌────────┐ ┌────────┐ ┌──────┐ ┌──────┐ ┌───────┐ ┌──────┐
    │  KING  │ │ QUEEN  │ │ ROOK │ │BISHOP│ │KNIGHT │ │ PAWN │
    │        │ │        │ │      │ │      │ │       │ │      │
    │ SNGMV  │ │  LINE  │ │ LINE │ │ LINE │ │SNGMV  │ │custom│
    └───┬────┘ └───┬────┘ └───┬──┘ └───┬──┘ └───┬───┘ └───┬──┘
        │          │          │        │        │         │
        └──────────┴──────────┴────────┴────────┴─────────┘
                                  │
                                  v
                          ┌───────────────┐
                          │    CMOVE      │
                          │ (calculate &  │
                          │validate move) │
                          └───────┬───────┘
                                  │
                          ┌───────┴────────┐
                          │                │
                          v                v
                   ┌─────────────┐   ┌──────────┐
                   │   JANUS     │   │ CHKCHK   │
                   │  (analyze)  │   │ (check   │
                   └──────┬──────┘   │  check)  │
                          │          └──────────┘
                          v
                   ┌─────────────┐
                   │   COUNTS    │
                   │    TREE     │
                   │  STRATGY    │
                   │   CKMATE    │
                   │    PUSH     │
                   └─────────────┘
```

---

## CMOVE (Move Calculation) Flow

```
┌─────────────────────────────────────────────┐
│ CMOVE (line 407)                            │
│ Input: SQUARE, MOVEN                        │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ newSquare = SQUARE  │
         │   + MOVEX[MOVEN]    │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ newSquare & $88     │
         │       != 0?         │
         └─────────┬───────────┘
                   │
            ┌──────┴──────┐
            │             │
           Yes           No
            │             │
            v             v
    ┌──────────────┐  ┌─────────────────┐
    │   ILLEGAL    │  │ Search BOARD    │
    │  (N flag=1)  │  │ for collision   │
    └──────────────┘  └────────┬────────┘
                                │
                         ┌──────┴──────┐
                         │             │
                    Found at      Not found
                     0-15          (empty)
                         │             │
                         v             v
                  ┌─────────────┐  ┌──────────┐
                  │  ILLEGAL    │  │  Clear   │
                  │ (own piece) │  │ V flag   │
                  └─────────────┘  └────┬─────┘
                         │               │
                    Found at             │
                     16-31               │
                         │               │
                         v               │
                  ┌─────────────┐        │
                  │ Set V flag  │        │
                  │ (capture)   │        │
                  └──────┬──────┘        │
                         │               │
                         └───────┬───────┘
                                 │
                                 v
                       ┌──────────────────┐
                       │ STATE in 0-7?    │
                       └─────────┬────────┘
                                 │
                          ┌──────┴──────┐
                          │             │
                         Yes           No
                          │             │
                          v             v
                   ┌──────────────┐  ┌──────────┐
                   │   CHKCHK     │  │  LEGAL   │
                   │ (expensive!) │  │ (N=0)    │
                   └──────┬───────┘  └──────────┘
                          │
                          v
                   ┌──────────────┐
                   │ Set C flag   │
                   │ if in check  │
                   └──────────────┘
```

---

## JANUS State Machine

```
┌─────────────────────────────────────────────┐
│ JANUS (line 162)                            │
│ Analysis director based on STATE            │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │   STATE value?      │
         └─────────┬───────────┘
                   │
      ┌────────────┼────────────┐
      │            │            │
   STATE >= 0   STATE = -7   STATE < 0
      │         ($F9)        (other)
      │            │            │
      v            v            v
┌─────────┐  ┌──────────┐  ┌──────────┐
│ COUNTS  │  │  Check   │  │  TREE    │
│(mobility│  │detection │  │(exchange │
│counting)│  │          │  │analysis) │
└────┬────┘  └────┬─────┘  └────┬─────┘
     │            │             │
     v            v             v
┌─────────┐  ┌──────────┐  ┌──────────┐
│ If V set│  │ If SQUARE│  │ Recursive│
│  then   │  │  == BK[0]│  │ capture  │
│  TREE   │  │INCHEK=0  │  │ analysis │
└─────────┘  └──────────┘  └──────────┘
     │
     v
┌─────────┐
│ If STATE│
│  == 4   │
│  ON4    │
└────┬────┘
     │
     v
┌─────────┐
│ STATE=0 │
│  MOVE   │
│ REVERSE │
│  GNMZ   │
│ REVERSE │
│ STATE=8 │
│  GNM    │
│ UMOVE   │
│ STRATGY │
└─────────┘
```

---

## STATE Variable Flow

```
Computer Move Selection (GO routine):

    STATE = $0C (12)
         │
         v
    ┌────────────────┐
    │  GNMX          │
    │  Generate      │
    │  candidate     │
    │  moves         │
    └────────────────┘
         │
         v
    STATE = $04 (4)
         │
         v
    ┌────────────────┐
    │  GNMZ          │
    │  Full analysis │
    │  of each move  │
    └───────┬────────┘
            │
            └─→ For each move:
                     │
                     v
                STATE = 0
                     │
                     v
                ┌────────────┐
                │ Generate   │
                │ immediate  │
                │ replies    │
                └─────┬──────┘
                      │
                      v
                STATE = 8
                      │
                      v
                ┌────────────┐
                │ Generate   │
                │ continuation│
                │ moves      │
                └─────┬──────┘
                      │
                      v
                 If capture:
                      │
                      v
                STATE: 4→3→2→1→0→-1→-2→-3→-4→-5
                      │
                      v
                ┌────────────┐
                │   TREE     │
                │  (exchange │
                │  analysis) │
                └────────────┘

Check Detection (CHKCHK):

    STATE = $F9 (-7)
         │
         v
    ┌────────────────┐
    │  MOVE          │
    │  REVERSE       │
    │  GNM           │
    └────────┬───────┘
             │
             v
    ┌────────────────┐
    │  JANUS checks  │
    │  if BK[0]      │
    │  captured      │
    └────────────────┘
```

---

## Move Execution Flow

```
┌─────────────────────────────────────────────┐
│ Player enters move: 4134 [Enter]            │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ INPUT/DISMV         │
         │ Build SQUARE=$34    │
         │ Find PIECE at $14   │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ [Enter] pressed     │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ MOVE                │
         └─────────┬───────────┘
                   │
                   v
      ┌────────────────────────┐
      │ Switch to SP2 stack    │
      │ TSX, STX SP1           │
      │ LDX SP2, TXS           │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Push to stack:         │
      │ 1. SQUARE (target)     │
      │ 2. Captured piece      │
      │ 3. BOARD[PIECE] (from) │
      │ 4. PIECE               │
      │ 5. MOVEN               │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Update BOARD:          │
      │ BOARD[PIECE] = SQUARE  │
      │ BOARD[captured] = $CC  │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Switch back to SP      │
      │ TSX, STX SP2           │
      │ LDX SP1, TXS           │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Return                 │
      └────────────────────────┘
```

---

## UMOVE (Unmake Move) Flow

```
┌─────────────────────────────────────────────┐
│ UMOVE (line 488)                            │
│ Restore position from move stack            │
└──────────────────┬──────────────────────────┘
                   │
                   v
      ┌────────────────────────┐
      │ Switch to SP2 stack    │
      │ TSX, STX SP1           │
      │ LDX SP2, TXS           │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Pop from stack:        │
      │ 1. MOVEN               │
      │ 2. PIECE               │
      │ 3. from_square         │
      │ 4. captured_piece      │
      │ 5. SQUARE              │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Restore BOARD:         │
      │ BOARD[PIECE]=from_sq   │
      │ If captured != $CC:    │
      │   BOARD[cap]=SQUARE    │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Switch back to SP      │
      │ (via STRV)             │
      └────────┬───────────────┘
               │
               v
      ┌────────────────────────┐
      │ Return                 │
      └────────────────────────┘
```

---

## Computer Player (GO) Flow

```
┌─────────────────────────────────────────────┐
│ GO (line 578)                               │
│ Select and play computer's move             │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Check opening book  │
         │ OMOVE >= 0?         │
         └─────────┬───────────┘
                   │
            ┌──────┴──────┐
            │             │
           Yes           No
            │             │
            v             v
    ┌──────────────┐  ┌──────────────┐
    │ DIS3 matches │  │ Set OMOVE=   │
    │ OPNING[X]?   │  │   $FF        │
    └──────┬───────┘  └──────┬───────┘
           │                 │
      ┌────┴────┐            │
      │         │            │
     Yes       No            │
      │         │            │
      v         └────────────┘
┌──────────┐                 │
│ Play     │                 │
│ book     │                 │
│ move     │                 │
│ Return   │                 │
└──────────┘                 │
                             v
                   ┌──────────────────┐
                   │ STATE = $0C      │
                   │ BESTV = $0C      │
                   │ GNMX (generate   │
                   │  candidates)     │
                   └─────────┬────────┘
                             │
                             v
                   ┌──────────────────┐
                   │ STATE = $04      │
                   │ GNMZ (analyze    │
                   │  all moves)      │
                   └─────────┬────────┘
                             │
                             v
                   ┌──────────────────┐
                   │ BESTV < $0F?     │
                   └─────────┬────────┘
                             │
                      ┌──────┴──────┐
                      │             │
                     Yes           No
                      │             │
                      v             v
              ┌──────────────┐  ┌─────────┐
              │ Load BESTP   │  │ Return  │
              │ Load BESTM   │  │  $FF    │
              │ MOVE         │  │ (mate)  │
              │ Display      │  └─────────┘
              │ Return       │
              └──────────────┘
```

---

## REVERSE (Flip Board) Flow

```
┌─────────────────────────────────────────────┐
│ REVERSE (line 382)                          │
│ Swap BOARD ↔ BK and transform coordinates   │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ X = $0F (15)        │
         │ Loop 16 times       │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ For each piece:     │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Y = BK[X]           │
         │ A = $77 - BOARD[X]  │
         │ BK[X] = A           │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ BOARD[X] = Y        │
         │ A = $77 - BOARD[X]  │
         │ BOARD[X] = A        │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ X = X - 1           │
         │ If X >= 0, loop     │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Return              │
         └─────────────────────┘

Effect: All squares transformed
    $00 ↔ $77
    $10 ↔ $67
    $34 ↔ $43
    etc.
```

---

## CHKCHK (Check Detection) Flow

```
┌─────────────────────────────────────────────┐
│ CHKCHK (line 444)                           │
│ Determine if move leaves king in check      │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Push STATE & flags  │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ STATE = $F9         │
         │ INCHEK = $FF        │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ MOVE (make move)    │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ REVERSE (swap sides)│
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ GNM (generate all   │
         │  opponent replies)  │
         └─────────┬───────────┘
                   │
                   └─→ For each reply:
                            │
                            v
                   ┌─────────────────┐
                   │ JANUS           │
                   │ If SQUARE==BK[0]│
                   │  INCHEK=0       │
                   └────────┬────────┘
                            │
         ┌──────────────────┘
         │
         v
┌─────────────────────┐
│ RUM (reverse &      │
│  unmake move)       │
└─────────┬───────────┘
          │
          v
┌─────────────────────┐
│ Pop STATE & flags   │
└─────────┬───────────┘
          │
          v
┌─────────────────────┐
│ Return with:        │
│ C=0 if INCHEK=$FF   │
│ C=1 if INCHEK=$00   │
└─────────────────────┘
```

---

## Stack Usage Diagram

```
Hardware Stack (SP):           Alternate Stack (SP2):

┌────────────────┐ $01FF      ┌────────────────┐ $C8 (initial)
│                │             │                │
│  Return        │             │  MOVEN         │ ← SP2
│  addresses     │             │  PIECE         │
│  from JSR      │             │  from_square   │
│                │             │  captured      │
│                │ ← SP        │  SQUARE        │
│                │             │                │
│                │             │  MOVEN         │
│                │             │  PIECE         │
│                │             │  from_square   │
│                │             │  captured      │
│                │             │  SQUARE        │
└────────────────┘ $0100      │                │
                               │  (deeper       │
                               │   moves...)    │
                               └────────────────┘ $00

When MOVE is called:
1. TSX, STX SP1     - Save hardware SP
2. LDX SP2, TXS     - Switch to alternate stack
3. Push 5 bytes     - Save move data
4. TSX, STX SP2     - Update SP2
5. LDX SP1, TXS     - Restore hardware SP
6. RTS              - Return using hardware stack

This allows recursive move generation without
corrupting the call stack!
```

---

## Evaluation Flow (STRATGY)

```
┌─────────────────────────────────────────────┐
│ STRATGY (line 641)                          │
│ Calculate position score                    │
└──────────────────┬──────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Base = $80 (128)    │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────────────────────┐
         │ Add 0.25 × mobility difference      │
         │  (WMOB - BMOB - PMOB)               │
         │ Add 0.25 × capture difference       │
         │  (WMAXC+WCC+WCAP1+WCAP2 -           │
         │   PMAXC-PCC-BCAP0-BCAP1-BCAP2)      │
         └─────────┬───────────────────────────┘
                   │
                   v
         ┌─────────────────────────────────────┐
         │ Add 0.50 × tactical advantage       │
         │  (WMAXC + WCC - BMAXC)              │
         └─────────┬───────────────────────────┘
                   │
                   v
         ┌─────────────────────────────────────┐
         │ Add 1.00 × material/threats         │
         │  (4×WCAP0 + WCAP1 -                 │
         │   2×BMAXC - 2×BMCC - BCAP1)         │
         └─────────┬───────────────────────────┘
                   │
                   v
         ┌─────────────────────────────────────┐
         │ If SQUARE in {$22,$25,$33,$34}      │
         │   (center squares)                  │
         │ OR piece moved out of back rank:    │
         │   Add 2 points                      │
         └─────────┬───────────────────────────┘
                   │
                   v
         ┌─────────────────────┐
         │ Continue to CKMATE  │
         │ (check detection)   │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ If king in check:   │
         │   Return 0          │
         │ If no moves & safe: │
         │   Return $FF (mate) │
         │ Else:               │
         │   Continue to PUSH  │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────┐
         │ PUSH                │
         │ Compare with BESTV  │
         │ Save if better      │
         └─────────────────────┘
```

---

## Overall Architecture Summary

```
┌──────────────────────────────────────────────────────────┐
│                     MicroChess Architecture              │
└──────────────────────────────────────────────────────────┘

User Interface Layer:
    CHESS → OUT → POUT/KIN → Command Dispatch

Move Generation Layer:
    GNM → Piece Handlers → CMOVE → JANUS

Evaluation Layer:
    COUNTS → TREE → STRATGY → CKMATE → PUSH

Move Execution Layer:
    MOVE/UMOVE (with dual stack)
    REVERSE (perspective flip)

AI Layer:
    GO → Opening Book OR Search
         (GNMX + GNMZ with STATE machine)

┌────────────────────────────────────────┐
│ Key Design Patterns:                  │
│                                        │
│ 1. STATE machine controls depth/mode  │
│ 2. Dual stack separates concerns      │
│ 3. Flags (N,V,C) carry move results   │
│ 4. Page zero for all state (fast)     │
│ 5. Recursive move generation          │
│ 6. Iterative deepening via STATE      │
└────────────────────────────────────────┘
```

---

## Critical Insights for Porting

1. **STATE Machine**: The single most important control mechanism
   - Different values trigger different behaviors in JANUS
   - Coordinates search depth and analysis type
   - Must be preserved exactly in port

2. **Dual Stack**: Enables recursive move generation
   - SP = function call stack
   - SP2 = game state stack
   - Go port: use separate slice for move history

3. **Flag-Based Returns**: CMOVE uses CPU flags for results
   - N flag = illegal move
   - V flag = capture
   - C flag = check
   - Go port: return struct with these booleans

4. **0x88 Board**: Edge detection in one instruction
   - Worth preserving in Go (it's elegant!)
   - Alternative: use 64-square with explicit bounds check

5. **REVERSE Trick**: Enables opponent analysis
   - Transform all coordinates: $77 - coord
   - Swap BOARD ↔ BK arrays
   - Brilliant space-saving technique

This call graph documentation should help understand the intricate relationships between routines and guide the Go port implementation!
