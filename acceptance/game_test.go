// ABOUTME: This file contains acceptance tests for MicroChess end-to-end functionality.
// ABOUTME: It tests complete user workflows like setting up the board and playing games.

package acceptance

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/matteo/microchess-go/pkg/microchess"
	"github.com/stretchr/testify/assert"
)

// captureOutput captures stdout during test execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestInitialDisplay tests the initial board display matches expected output
func TestInitialDisplay(t *testing.T) {
	game := microchess.NewGame()

	output := captureOutput(func() {
		game.Display()
	})

	expected := `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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

`

	assert.Equal(t, expected, output, "Initial display should match expected output")
}

// TestSetupBoardCommand tests the 'C' command produces the correct full output
func TestSetupBoardCommand(t *testing.T) {
	game := microchess.NewGame()

	output := captureOutput(func() {
		shouldContinue := game.HandleCommand("C")
		assert.True(t, shouldContinue, "C command should not quit")
	})

	expected := `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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

`

	assert.Equal(t, expected, output, "Setup command output should match expected")

	// Verify the game state was actually updated
	assert.Equal(t, uint8(0xCC), game.DIS1, "DIS1 should be 0xCC")
	assert.Equal(t, uint8(0xCC), game.DIS2, "DIS2 should be 0xCC")
	assert.Equal(t, uint8(0xCC), game.DIS3, "DIS3 should be 0xCC")

	// Verify piece positions
	assert.Equal(t, uint8(0x03), uint8(game.Board[microchess.PieceKing]), "White king should be at 0x03")
	assert.Equal(t, uint8(0x73), uint8(game.BK[microchess.PieceKing]), "Black king should be at 0x73")
}

// TestQuitCommand tests the 'Q' command returns false (quit)
func TestQuitCommand(t *testing.T) {
	game := microchess.NewGame()

	shouldContinue := game.HandleCommand("Q")

	assert.False(t, shouldContinue, "Q command should return false to quit")
}

// TestUnknownCommand tests that unknown commands are handled gracefully
func TestUnknownCommand(t *testing.T) {
	game := microchess.NewGame()

	output := captureOutput(func() {
		shouldContinue := game.HandleCommand("X")
		assert.True(t, shouldContinue, "Unknown command should not quit")
	})

	expected := `Unknown command: X
Available commands: C (setup), Q (quit)
`

	assert.Equal(t, expected, output, "Unknown command should show error message")
}

// TestSetupAndQuitSequence tests the complete workflow: start -> setup -> quit
// This simulates: printf "C\nQ\n" | go run cmd/microchess/main.go
func TestSetupAndQuitSequence(t *testing.T) {
	game := microchess.NewGame()

	// Build the full output as user would see it
	var fullOutput bytes.Buffer

	// Capture initial display
	output1 := captureOutput(func() {
		game.Display()
	})
	fullOutput.WriteString(output1)

	// Execute 'C' command
	output2 := captureOutput(func() {
		shouldContinue := game.HandleCommand("C")
		assert.True(t, shouldContinue, "C command should not quit")
	})
	fullOutput.WriteString(output2)

	// Execute 'Q' command (no output expected)
	shouldContinue := game.HandleCommand("Q")
	assert.False(t, shouldContinue, "Q command should quit")

	// Expected full console output
	expected := `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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

`

	assert.Equal(t, expected, fullOutput.String(), "Full console output should match expected")
}
