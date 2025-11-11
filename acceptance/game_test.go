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
			runTestCase(t, tc)
		})
	}
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
