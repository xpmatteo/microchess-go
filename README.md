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
- **Go Port**: Matteo Vaccari 2025 with the help of Claude Code
