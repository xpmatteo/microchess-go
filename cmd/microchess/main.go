// ABOUTME: This is the main CLI entry point for MicroChess.
// ABOUTME: It provides a text-based interface similar to the original 1976 serial terminal version.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/matteo/microchess-go/pkg/microchess"
)

func main() {
	game := microchess.NewGame(os.Stdout)

	game.Display()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("? ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		command := strings.TrimSpace(strings.ToUpper(input))

		if !game.HandleCommand(command) {
			return
		}
	}
}
