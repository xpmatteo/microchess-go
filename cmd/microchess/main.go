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

		switch command {
		case "Q":
			// Quit program
			return

		case "C":
			// Setup board (SETUP routine, line 665)
			game.SetupBoard()
			// Set LED display to "CC CC CC" to indicate setup
			game.DIS1 = 0xCC
			game.DIS2 = 0xCC
			game.DIS3 = 0xCC
			displayBoard(game)

		default:
			fmt.Println("Unknown command:", command)
			fmt.Println("Available commands: C (setup), Q (quit)")
		}
	}
}

// displayBoard prints the chess board in the style of the original POUT routine (line 702).
// The display shows coordinates and piece positions using the 0x88 encoding.
func displayBoard(game *microchess.GameState) {
	fmt.Println("MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com")
	fmt.Println(" 00 01 02 03 04 05 06 07")
	fmt.Println("-------------------------")

	// Display ranks 0 to 7 (original displays 00-70)
	// The original scans Y from 0x00 to 0x77 in 0x88 format
	for rank := 0; rank <= 7; rank++ {
		fmt.Print("|")

		for file := 0; file < 8; file++ {
			sq := board.Square((rank << 4) | file)
			piece, found, isWhite := game.FindPieceAt(sq)

			if found {
				fmt.Print(microchess.GetPieceChar(piece, isWhite))
			} else {
				// Checkerboard pattern for empty squares
				// Original: check if (file + rank) is odd for asterisk
				if (rank+file)%2 == 1 {
					fmt.Print("**")
				} else {
					fmt.Print("  ")
				}
			}

			fmt.Print("|")
		}

		// Print rank number in hex on the right (00, 10, 20, ...)
		fmt.Printf("%X0\n", rank)
	}

	fmt.Println("-------------------------")
	fmt.Println(" 00 01 02 03 04 05 06 07")

	// Print LED display (DIS1 DIS2 DIS3)
	fmt.Printf("%02X %02X %02X\n", game.DIS1, game.DIS2, game.DIS3)
	fmt.Println()
}
