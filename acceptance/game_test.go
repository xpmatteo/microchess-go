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

// testCase represents a complete test scenario with multiple steps
type testCase struct {
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	FinalReversed bool          `yaml:"final_reversed"`
	Steps         []commandStep `yaml:"steps"`
	Skip          bool          `yaml:"skip"`
}

// commandStep represents one or more commands and the expected final output
type commandStep struct {
	Commands        string `yaml:"commands"`         // One or more commands to execute
	ShouldContinue  bool   `yaml:"should_continue"`  // Expected continue state after all commands
	ExpectedDisplay string `yaml:"expected_display"` // Expected final display after all commands
}

// TestUnknownCommand tests that unknown commands are handled gracefully
func TestUnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	game := microchess.NewGame(&buf)

	shouldContinue := game.HandleCharacter('X')
	assert.True(t, shouldContinue, "Unknown command should not quit")

	output := buf.String()

	expected := "\r\nUnknown command: X\r\nAvailable commands: C (setup), E (reverse), P (print), Q (quit), 0-7 (move)\n"

	assert.Equal(t, expected, output, "Unknown command should show error message")
}

// TestCommandSequences loads and executes all YAML test cases from testdata/
func TestCommandSequences(t *testing.T) {
	// Find all YAML files in testdata/
	files, err := os.ReadDir("testdata")
	require.NoError(t, err, "Failed to read testdata directory")

	var yamlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			yamlFiles = append(yamlFiles, file.Name())
		}
	}

	require.NotEmpty(t, yamlFiles, "No YAML test files found in testdata/")

	// Run each test case as a subtest
	for _, filename := range yamlFiles {
		tc := loadTestCase(t, filename)
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Skip {
				t.Skip()
			} else {
				runTestCase(t, tc)
			}
		})
	}
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
		// Reset buffer to capture output for this step
		buf.Reset()

		var shouldContinue bool
		// Special case: "DISPLAY" is a test-only command to show the current board
		if step.Commands == "DISPLAY" {
			game.Display()
			shouldContinue = true
		} else {
			// Execute all commands in sequence, accumulating output
			for _, ch := range step.Commands {
				shouldContinue = game.HandleCharacter(byte(ch))
				if !shouldContinue {
					break
				}
			}
		}

		assert.Equal(t, step.ShouldContinue, shouldContinue,
			"Step %d (%s): shouldContinue mismatch", i, step.Commands)

		// Extract only the LAST display from the output (for multi-char commands)
		// This matches the behavior we expect: only check the final result
		allOutput := buf.String()
		lastDisplay := extractLastDisplay(allOutput)

		// Normalize line endings (convert \r\n to \n) and trim whitespace
		// This allows tests to work in both raw terminal mode (\r\n) and normal mode (\n)
		expected := strings.TrimSpace(step.ExpectedDisplay)
		actual := strings.ReplaceAll(lastDisplay, "\r\n", "\n")
		actual = strings.TrimSpace(actual)
		// Not using testify for this assertion, so that we produce a more readable error message
		if expected != actual {
			t.Errorf("Step %d (%s): Display mismatch\n=== EXPECTED ===\n%s\n=== ACTUAL (Go) ===\n%s\n=== END ===",
				i, step.Commands, expected, actual)
		}
	}

	// Verify final state if specified
	if tc.Name != "" && strings.Contains(strings.ToLower(tc.Name), "reverse") {
		assert.Equal(t, tc.FinalReversed, game.Reversed,
			"Final reversed state should match expected: %v", tc.FinalReversed)
	}
}

// extractLastDisplay extracts the last complete board display from output
// For multi-character commands that generate multiple displays, we only want the final one
func extractLastDisplay(output string) string {
	// Split by MicroChess header to find display boundaries
	displays := strings.Split(output, "MicroChess")
	if len(displays) <= 1 {
		return output // No header found or only one display
	}

	// Get the last display and prepend the header back
	lastDisplay := displays[len(displays)-1]
	return "MicroChess" + lastDisplay
}
