// ABOUTME: This is the main CLI entry point for MicroChess.
// ABOUTME: It provides a text-based interface similar to the original 1976 serial terminal version.

package main

import (
	"fmt"
	"os"

	"github.com/matteo/microchess-go/pkg/microchess"
	"golang.org/x/term"
)

func main() {
	game := microchess.NewGame(os.Stdout)
	game.Display()

	// Check if stdin is a terminal or a pipe
	fd := int(os.Stdin.Fd())
	isTerminal := term.IsTerminal(fd)

	if isTerminal {
		// Terminal mode: use raw mode for character-by-character input
		// This matches the original 1976 MicroChess serial terminal behavior
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not set raw terminal mode: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			_ = term.Restore(fd, oldState)
		}()

		// Read one character at a time, just like the original assembly (syskin routine)
		buf := make([]byte, 1)
		for {
			fmt.Print("? ")

			// Blocking read of single character
			_, err := os.Stdin.Read(buf)
			if err != nil {
				fmt.Println("\r\nError reading input:", err)
				continue
			}

			char := buf[0]

			// Echo the character (original does this via syschout)
			fmt.Printf("%c", char)

			// Handle the character
			if !game.HandleCharacter(char) {
				fmt.Println("\r") // Clean newline before exit
				return
			}
		}
	} else {
		// Piped input mode: read character-by-character but don't use raw mode
		// This allows testing with printf 'CPQ' | ./microchess
		buf := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				// EOF or error - exit gracefully
				return
			}

			char := buf[0]

			// Handle the character (no echo needed for piped input)
			if !game.HandleCharacter(char) {
				return
			}
		}
	}
}
