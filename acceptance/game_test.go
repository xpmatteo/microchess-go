// ABOUTME: This file contains acceptance tests for MicroChess end-to-end functionality.
// ABOUTME: It tests complete user workflows like setting up the board and playing games.

package acceptance

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/matteo/microchess-go/pkg/microchess"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
// Test data is loaded from testdata/setup-quit.yaml
func TestSetupAndQuitSequence(t *testing.T) {
	tc := loadTestCase(t, "setup-quit.yaml")
	runTestCase(t, tc)
}

// commandStep represents a single command and its expected output
type commandStep struct {
	Command         string `yaml:"command"`
	ShouldContinue  bool   `yaml:"should_continue"`
	ExpectedDisplay string `yaml:"expected_display"`
}

// testCase represents a complete test scenario with multiple steps
type testCase struct {
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	FinalReversed bool          `yaml:"final_reversed"`
	Steps         []commandStep `yaml:"steps"`
}

// loadTestCase loads a test case from a YAML file
func loadTestCase(t *testing.T, filename string) testCase {
	data, err := os.ReadFile("testdata/" + filename)
	require.NoError(t, err, "Failed to read test data file: %s", filename)

	var tc testCase
	err = yaml.Unmarshal(data, &tc)
	require.NoError(t, err, "Failed to parse YAML test data: %s", filename)

	return tc
}

// runTestCase executes a test case's steps and validates output
func runTestCase(t *testing.T, tc testCase) {
	t.Logf("Running test case: %s", tc.Name)
	t.Logf("Description: %s", tc.Description)

	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	for i, step := range tc.Steps {
		buf.Reset()

		var shouldContinue bool
		if step.Command == "DISPLAY" {
			game.Display()
			shouldContinue = true
		} else {
			shouldContinue = game.HandleCommand(step.Command)
		}

		assert.Equal(t, step.ShouldContinue, shouldContinue,
			"Step %d (%s): shouldContinue mismatch", i, step.Command)

		// Trim whitespace from both strings to avoid whitespace issues
		expected := strings.TrimSpace(step.ExpectedDisplay)
		actual := strings.TrimSpace(buf.String())
		assert.Equal(t, expected, actual,
			"Step %d (%s): display output mismatch", i, step.Command)
	}

	// Verify final state if specified
	if tc.Name != "" && strings.Contains(strings.ToLower(tc.Name), "reverse") {
		assert.Equal(t, tc.FinalReversed, game.Reversed,
			"Final reversed state should match expected: %v", tc.FinalReversed)
	}
}

// TestSetupReverseAndQuitSequence tests: start -> setup -> reverse -> reverse -> quit
// This simulates: printf "C\nE\nE\nQ\n" | go run cmd/microchess/main.go
// Test data is loaded from testdata/setup-reverse-quit.yaml
func TestSetupReverseAndQuitSequence(t *testing.T) {
	tc := loadTestCase(t, "setup-reverse-quit.yaml")
	runTestCase(t, tc)
}
