# MicroChess Go Port

A Go port of Peter Jennings' historic 1976 MicroChess program - one of the first chess programs for microcomputers.

## About

This project aims to make the remarkable MicroChess program easier to understand while preserving its exact chess logic and algorithms. The original program packed a complete chess AI into just ~1.5KB of 6502 assembly code!

## Project Status

**Phase 2 Complete** - Basic board display and initial setup

✅ Implemented:
- 0x88 board representation
- Basic types (Square, Piece, GameState)
- Initial board setup
- Board display in original terminal style
- Simple REPL with quit command

🚧 Coming Next (Phase 3):
- Board setup command ('C')
- Board reverse command ('E')
- Interactive board manipulation

## How to Build

```bash
go build -o microchess cmd/microchess/main.go
```

## How to Run

```bash
./microchess
```

Or run directly:

```bash
go run cmd/microchess/main.go
```

## Current Features

The program currently displays the initial chess position and accepts the following command:

- **Q** - Quit the program

## Example Session

```
$ ./microchess
MicroChess (c) 1976-2025 Peter Jennings

 00 01 02 03 04 05 06 07
-------------------------
|BR BN BB BQ BK BB BN BR|70
|BP BP BP BP BP BP BP BP|60
| *  *  *  *  *  *  *  *|50
|  *  *  *  *  *  *  * *|40
| *  *  *  *  *  *  *  *|30
|  *  *  *  *  *  *  * *|20
|WP WP WP WP WP WP WP WP|10
|WR WN WB WQ WK WB WN WR|00
-------------------------
 00 01 02 03 04 05 06 07

? Q
Goodbye!
```

## Testing

Run the test suite:

```bash
go test ./...
```

All core types and board representation are covered by unit tests.

## Architecture

- **pkg/board/** - 0x88 board representation and Square type
- **pkg/microchess/** - Core game types and state
- **cmd/microchess/** - CLI interface

## Original Source

The original MicroChess 6502 assembly source code and documentation are in the `doc/` directory:

- `doc/Microchess6502.txt` - Complete assembly source
- `doc/DATA_STRUCTURES.md` - Data structure documentation
- `doc/SUBROUTINES.md` - Routine documentation
- `doc/COMMANDS.md` - User interface documentation
- `doc/CALL_GRAPH.md` - Flow diagrams

## Credits

- **Original Author**: Peter Jennings (1976)
- **Serial Port Adaptation**: Daryl Rictor (2002)
- **OCR Corrections**: Bill Forster (2005)
- **Go Port**: 2025

## License

This is a historical preservation project. The original MicroChess was distributed as public domain software.

## Learn More

See `doc/PORTING_PLAN.md` for the complete development roadmap.
