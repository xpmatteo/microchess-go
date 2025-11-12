# MicroChess Go Port

A Go port of Peter Jennings' historic 1976 MicroChess program - one of the first chess programs for microcomputers.

## About

This project is an experiment in AI-assisted modernization: how do we plan and execute a gradual port, given that we want to move away from 6502 assembly?

We also aim to make the remarkable MicroChess program easier to understand while preserving its exact chess logic and algorithms. The original program packed a complete chess AI into just ~1.5KB of 6502 assembly code!

## Project Status

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

The program starts with an empty board (matching the original 6502 assembly behavior) and accepts:

- **C** - Setup the board to initial chess position
- **Q** - Quit the program

## Example Session

```
$ ./microchess
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

? C
MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|WR|WN|WB|WK|WQ|WB|WN|WR|00
|WP|WP|WP|WP|WP|WP|WP|WP|10
|  |**|  |**|  |**|  |**|20
|**|  |**|  |**|  |**|  |30
|  |**|  |**|  |**|  |**|40
|**|  |**|  |**|  |**|  |50
|BP|BP|BP|BP|BP|BP|BP|BP|60
|BR|BN|BB|BK|BQ|BB|BN|BR|70
-------------------------
 00 01 02 03 04 05 06 07
CC CC CC

? Q
```

## Testing

Run the test suite:

```bash
go test ./...
```

The project includes:
- **Unit tests** - Test individual components (board, types, etc.)
- **Acceptance tests** - Test complete user workflows end-to-end

Run only acceptance tests:
```bash
go test ./acceptance/...
```

All core types, board representation, and user commands are covered by tests.

## Architecture

- **pkg/board/** - 0x88 board representation and Square type
- **pkg/microchess/** - Core game types, state, and command handling
- **cmd/microchess/** - CLI interface (thin wrapper around GameState)
- **acceptance/** - End-to-end acceptance tests

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
