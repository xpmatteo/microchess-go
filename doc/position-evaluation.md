# Cleanroom Position Evaluation

This note records the exact, implementation-agnostic formula MicroChess uses to score a move. It is derived from the published counter semantics in `doc/microchess-manual.txt` and the arithmetic visible in `go6502/microchess.asm`, but it intentionally avoids any assembly-specific details so it can be reimplemented from scratch.

## Symbols

All counters are unsigned 8-bit values. Addition and subtraction are carried out modulo 256 unless clamped or shifted as noted.

| Symbol | Meaning |
| --- | --- |
| `WMOB` | mobility achieved after the tentative move (queens count twice) |
| `WMAXC` | value (in material points) of the most valuable enemy piece we attack after the move |
| `WCC` | total material value of all enemy pieces we attack after the move |
| `WCAP0` | best-case material gain from the first ply of an exchange sequence we initiate |
| `WCAP1`, `WCAP2` | best-case gains for the 2nd and 3rd plies of that exchange |
| `PMAXC`, `PCC`, `PMOB` | our baseline `MAXC`, `CC`, `MOB` before making the move |
| `BCAP0`, `BCAP1`, `BCAP2` | opponent’s best gains from reply exchange sequences |
| `BMOB` | opponent’s mobility following our move |
| `BMAXC` | most valuable friendly piece the opponent can attack after our move |
| `BMCC` | sum of all friendly pieces the opponent can attack after our move |
| `WMAXP` | ID of the opponent’s piece currently under our best attack (non-zero iff a capture is available) |

The constants `128`, `64`, and `144` below are literal byte values (`0x80`, `0x40`, `0x90`).

## Score Formula

1. **Quarter-weight blend**
   ```
   stage1 = max(0, 128 + WMOB + WMAXC + WCC + WCAP1 + WCAP2
                    - PMAXC - PCC - BCAP0 - BCAP1 - BCAP2 - PMOB - BMOB)
   stage1 = floor(stage1 / 2)
   ```
   This collects our follow-up activity, subtracts the baseline and reply liabilities, clamps at zero, then halves the result.

2. **Half-weight blend**
   ```
   stage2 = floor((stage1 + 64 + WMAXC + WCC - BMAXC) / 2)
   ```
   This term reuses the updated totals, biases them, subtracts the opponent’s single best counter-threat, and halves again.

3. **Full-weight exchange term**
   ```
   value = stage2 + 144 + 4*WCAP0 + WCAP1 - 2*BMAXC - 2*BMCC - BCAP1
   ```
   Multiplication by four is exact integer multiplication; subtraction is modulo 256.

4. **Positional kicker**
   ```
   if toSquare ∈ {022, 025, 033, 034}           // octal ranks/files
      value += 2
   else if movingPiece ≠ KING and fromSquare on back rank
      value += 2
   ```

5. **Terminal overrides**
   ```
   if opponent can capture our king after the move: value = 0x00
   else if opponent has no legal replies and is in check: value = 0xFF
   ```

After these steps the accumulator holds the final byte-sized score used to compare moves. Any cleanroom implementation should reproduce the same arithmetic–including the intermediate clamps and halvings–to stay behaviorally identical.

## Mobility Counters

Mobility is counted directly by the move generator (`go6502/microchess.asm:160-199`). Every legal destination enumerated for the piece referenced by `PIECE` increments the counter located at `MOB + STATE`. The same byte-aligned storage is later accessed via the symbolic names:

- `STATE = 0x00` → `BMOB` (opponent replies during exchange search, `go6502/microchess.asm:206-220`)
- `STATE = 0x08` → `WMOB` (our continuation moves after making the candidate move)
- `STATE = 0x0C` → `PMOB` (baseline before the move, captured at the start of `GO`, `go6502/microchess.asm:593-600`)

Queens (piece id `0x01`) contribute two counts per destination, matching the historical rule that “each queen move counts as two” (`go6502/microchess.asm:175-178`). No other weighting is applied: the counter is literally `number of pseudo-legal moves produced`, with impossible moves suppressed earlier in the pipeline (`CMOVE`/`JANUS`). Clean implementations should mirror this by:

1. Running the move generator for the relevant side/state.
2. Incrementing the mobility total once for every legal move that survives the legality checks.
3. Incrementing it one additional time whenever the moving piece is a queen.

## Material Points Table

The `POINTS` table (`go6502/microchess.asm:878`) provides the canonical material weights used everywhere—capture evaluation, exchange stacks, and the mate detection that checks whether the king is hanging (`go6502/microchess.asm:543-546`). The indices align with the `BK`/`BOARD` piece ordering: king, queen, two rooks, two bishops, two knights, and eight pawns.

| Index | Piece         | Value (hex) | Value (decimal) |
| --- | --- | --- | --- |
| 0 | King  | `0x0B` | 11 |
| 1 | Queen | `0x0A` | 10 |
| 2 | Rook (QR) | `0x06` | 6 |
| 3 | Rook (KR) | `0x06` | 6 |
| 4 | Bishop (QB) | `0x04` | 4 |
| 5 | Bishop (KB) | `0x04` | 4 |
| 6 | Knight (QN) | `0x04` | 4 |
| 7 | Knight (KN) | `0x04` | 4 |
| 8-15 | Pawns (files a through h in octal order) | `0x02` | 2 |

Any clean implementation that needs a material value should read it from this static table using the same index the board representation uses for the piece in question. Changing these weights would constitute a behavioral change and should only be done intentionally.
