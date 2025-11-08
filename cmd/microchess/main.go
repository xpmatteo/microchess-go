// ABOUTME: This is the main CLI entry point for MicroChess.
// ABOUTME: It provides a text-based interface similar to the original 1976 serial terminal version.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/matteo/microchess-go/pkg/board"
	"github.com/matteo/microchess-go/pkg/microchess"
)

func main() {
	game := microchess.NewGame()

	fmt.Println("MicroChess (c) 1976-2025 Peter Jennings")
	fmt.Println()

	displayBoard(game)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("? ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		command := strings.TrimSpace(strings.ToUpper(input))

		if command == "Q" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Println("Unknown command:", command)
		fmt.Println("Available commands: Q (quit)")
	}
}

// displayBoard prints the chess board in the style of the original POUT routine (line 702).
// The display shows coordinates and piece positions using the 0x88 encoding.
func displayBoard(game *microchess.GameState) {
	fmt.Println(" 00 01 02 03 04 05 06 07")
	fmt.Println("-------------------------")

	// Display ranks 7 down to 0 (8th rank to 1st rank)
	for rank := 7; rank >= 0; rank-- {
		fmt.Print("|")

		for file := 0; file < 8; file++ {
			sq := board.Square((rank << 4) | file)
			piece, found, isWhite := game.FindPieceAt(sq)

			if found {
				fmt.Print(microchess.GetPieceChar(piece, isWhite))
			} else {
				// Checkerboard pattern for empty squares
				if (rank+file)%2 == 0 {
					fmt.Print(" *")
				} else {
					fmt.Print("  ")
				}
			}

			if file < 7 {
				fmt.Print(" ")
			}
		}

		// Print rank number in hex on the right
		fmt.Printf("|%X0\n", rank)
	}

	fmt.Println("-------------------------")
	fmt.Println(" 00 01 02 03 04 05 06 07")
	fmt.Println()
}
