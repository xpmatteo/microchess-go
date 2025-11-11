// ABOUTME: This file contains acceptance tests for MicroChess end-to-end functionality.
// ABOUTME: It tests complete user workflows like setting up the board and playing games.

package acceptance

import (
	"bytes"
	"testing"

	"github.com/matteo/microchess-go/pkg/microchess"
	"github.com/stretchr/testify/assert"
)

// TestInitialDisplay tests the initial board display matches expected output
func TestInitialDisplay(t *testing.T) {
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	game.Display()
	output := buf.String()

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
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	shouldContinue := game.HandleCommand("C")
	assert.True(t, shouldContinue, "C command should not quit")

	output := buf.String()

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
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	shouldContinue := game.HandleCommand("Q")

	assert.False(t, shouldContinue, "Q command should return false to quit")
}

// TestUnknownCommand tests that unknown commands are handled gracefully
func TestUnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	shouldContinue := game.HandleCommand("X")
	assert.True(t, shouldContinue, "Unknown command should not quit")

	output := buf.String()

	expected := `Unknown command: X
Available commands: C (setup), E (reverse), Q (quit)
`

	assert.Equal(t, expected, output, "Unknown command should show error message")
}

// TestReverseCommand tests the 'E' command produces the correct full output
func TestReverseCommand(t *testing.T) {
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)
	game.SetupBoard() // Setup first to have pieces

	shouldContinue := game.HandleCommand("E")
	assert.True(t, shouldContinue, "E command should not quit")

	output := buf.String()

	// After E command, board is flipped: black pieces at bottom, white at top
	expected := `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|BR|BN|BB|BQ|BK|BB|BN|BR|00
|BP|BP|BP|BP|BP|BP|BP|BP|10
|  |**|  |**|  |**|  |**|20
|**|  |**|  |**|  |**|  |30
|  |**|  |**|  |**|  |**|40
|**|  |**|  |**|  |**|  |50
|WP|WP|WP|WP|WP|WP|WP|WP|60
|WR|WN|WB|WQ|WK|WB|WN|WR|70
-------------------------
 00 01 02 03 04 05 06 07
EE EE EE

`

	assert.Equal(t, expected, output, "Reverse command output should match expected")

	// Verify the game state was actually updated
	assert.Equal(t, uint8(0xEE), game.DIS1, "DIS1 should be 0xEE")
	assert.Equal(t, uint8(0xEE), game.DIS2, "DIS2 should be 0xEE")
	assert.Equal(t, uint8(0xEE), game.DIS3, "DIS3 should be 0xEE")
	assert.True(t, game.Reversed, "Reversed flag should be true")
}

// TestSetupAndQuitSequence tests the complete workflow: start -> setup -> quit
// This simulates: printf "C\nQ\n" | go run cmd/microchess/main.go
func TestSetupAndQuitSequence(t *testing.T) {
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	// Capture initial display
	game.Display()

	// Execute 'C' command
	shouldContinue := game.HandleCommand("C")
	assert.True(t, shouldContinue, "C command should not quit")

	// Execute 'Q' command (no output expected)
	shouldContinue = game.HandleCommand("Q")
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

	assert.Equal(t, expected, buf.String(), "Full console output should match expected")
}

// TestSetupReverseAndQuitSequence tests: start -> setup -> reverse -> reverse -> quit
// This simulates: printf "C\nE\nE\nQ\n" | go run cmd/microchess/main.go
func TestSetupReverseAndQuitSequence(t *testing.T) {
	type commandStep struct {
		command         string
		shouldContinue  bool
		expectedDisplay string
	}

	steps := []commandStep{
		{
			command:        "DISPLAY", // Initial display (not a command, just triggers Display())
			shouldContinue: true,
			expectedDisplay: `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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

`,
		},
		{
			command:        "C",
			shouldContinue: true,
			expectedDisplay: `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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

`,
		},
		{
			command:        "E",
			shouldContinue: true,
			expectedDisplay: `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
 00 01 02 03 04 05 06 07
-------------------------
|BR|BN|BB|BQ|BK|BB|BN|BR|00
|BP|BP|BP|BP|BP|BP|BP|BP|10
|  |**|  |**|  |**|  |**|20
|**|  |**|  |**|  |**|  |30
|  |**|  |**|  |**|  |**|40
|**|  |**|  |**|  |**|  |50
|WP|WP|WP|WP|WP|WP|WP|WP|60
|WR|WN|WB|WQ|WK|WB|WN|WR|70
-------------------------
 00 01 02 03 04 05 06 07
EE EE EE

`,
		},
		{
			command:        "E",
			shouldContinue: true,
			expectedDisplay: `MicroChess (c) 1996-2005 Peter Jennings, www.benlo.com
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
EE EE EE

`,
		},
		{
			command:         "Q",
			shouldContinue:  false,
			expectedDisplay: "", // Q produces no output
		},
	}

	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	for i, step := range steps {
		buf.Reset()

		var shouldContinue bool
		if step.command == "DISPLAY" {
			game.Display()
			shouldContinue = true
		} else {
			shouldContinue = game.HandleCommand(step.command)
		}

		assert.Equal(t, step.shouldContinue, shouldContinue,
			"Step %d (%s): shouldContinue mismatch", i, step.command)
		assert.Equal(t, step.expectedDisplay, buf.String(),
			"Step %d (%s): display output mismatch", i, step.command)
	}

	// Verify final state - should be back to normal after double reverse
	assert.False(t, game.Reversed, "Reversed flag should be false after double E")
}
