// ABOUTME: This file tests the original 6502 MicroChess against YAML test cases.
// ABOUTME: It runs the emulator programmatically and compares output with expectations.

package acceptance

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beevik/go6502/cpu"
	"github.com/stretchr/testify/require"
)

// TestIoMemory implements cpu.Memory with buffered I/O for testing
type TestIoMemory struct {
	mem       *cpu.FlatMemory
	input     io.Reader
	output    io.Writer
	lastInput byte
}

// NewTestIoMemory creates a new memory with buffered I/O
func NewTestIoMemory(input io.Reader, output io.Writer) *TestIoMemory {
	return &TestIoMemory{
		mem:    cpu.NewFlatMemory(),
		input:  input,
		output: output,
	}
}

// LoadByte loads a byte, with I/O at $FFF1
func (m *TestIoMemory) LoadByte(addr uint16) byte {
	// $FFF1 - Character input (blocking read from input buffer)
	if addr == 0xFFF1 {
		if m.lastInput == 0 {
			buf := make([]byte, 1)
			_, err := m.input.Read(buf)
			if err != nil {
				return 0
			}
			m.lastInput = buf[0]
		}
		result := m.lastInput
		m.lastInput = 0
		return result
	}
	return m.mem.LoadByte(addr)
}

// LoadBytes passes through to underlying memory
func (m *TestIoMemory) LoadBytes(addr uint16, b []byte) {
	m.mem.LoadBytes(addr, b)
}

// LoadAddress passes through to underlying memory
func (m *TestIoMemory) LoadAddress(addr uint16) uint16 {
	return m.mem.LoadAddress(addr)
}

// StoreByte stores a byte, with I/O at $FFF0
func (m *TestIoMemory) StoreByte(addr uint16, v byte) {
	// $FFF0 - Character output (write to output buffer)
	if addr == 0xFFF0 {
		_, _ = m.output.Write([]byte{v})
		return
	}
	m.mem.StoreByte(addr, v)
}

// StoreBytes passes through to underlying memory
func (m *TestIoMemory) StoreBytes(addr uint16, b []byte) {
	m.mem.StoreBytes(addr, b)
}

// StoreAddress passes through to underlying memory
func (m *TestIoMemory) StoreAddress(addr uint16, v uint16) {
	m.mem.StoreAddress(addr, v)
}

// Test6502CommandSequences runs YAML test cases against the original 6502 MicroChess
func Test6502CommandSequences(t *testing.T) {
	// Check if microchess.bin exists
	_, err := os.Stat("../go6502/microchess.bin")
	require.NoError(t, err)

	// Find all YAML test files recursively
	yamlFiles, err := filepath.Glob("testdata/**/*.yaml")
	require.NoError(t, err, "Failed to glob YAML files")
	require.NotEmpty(t, yamlFiles, "No YAML test files found in testdata/")

	// Run each test case as a subtest
	for _, filepath := range yamlFiles {
		tc := loadTestCase(t, filepath)
		t.Run(tc.Name+" (6502)", func(t *testing.T) {
			if tc.Skip {
				t.Skip()
			} else if tc.Skip6502 {
				t.Skip("Skipped on 6502: test uses Go port extensions not available in original 1976 program")
			} else {
				run6502TestCase(t, tc)
			}
		})
	}
}

// run6502TestCase executes a test case against the 6502 emulator
func run6502TestCase(t *testing.T, tc testCase) {
	t.Logf("Running 6502 test case: %s", tc.Name)
	t.Logf("Description: %s", tc.Description)

	// Load the binary
	data, err := os.ReadFile("../go6502/microchess.bin")
	require.NoError(t, err, "Failed to load microchess.bin")

	// Create buffers for I/O
	var inputBuf bytes.Buffer
	var outputBuf bytes.Buffer

	// Build input string from commands
	// Skip "DISPLAY" as it's a test-only command, not a real 6502 command
	for _, step := range tc.Steps {
		if step.Commands != "DISPLAY" {
			inputBuf.WriteString(step.Commands)
		}
	}

	// Create CPU with I/O memory that uses our buffers
	mem := NewTestIoMemory(&inputBuf, &outputBuf)
	c := cpu.NewCPU(cpu.CMOS, mem)

	// Load program at $1000
	mem.StoreBytes(0x1000, data)
	c.SetPC(0x1000)

	// Run until program halts or max cycles
	maxCycles := uint64(10000000) // Safety limit (increased for complex operations)
	for c.Cycles < maxCycles {
		// Check for BRK or infinite loop
		if mem.LoadByte(c.Reg.PC) == 0x00 {
			break
		}
		c.Step()
	}

	// Get the full output
	output := outputBuf.String()
	t.Logf("6502 output length: %d bytes", len(output))

	// Split output by displays
	displays := splitDisplays(output)
	t.Logf("Found %d displays in 6502 output", len(displays))

	// Debug: log each display
	for i, disp := range displays {
		normalized := normalizeDisplay(disp)
		// Extract status code from display
		lines := strings.Split(normalized, "\n")
		statusLine := lines[len(lines)-1]
		t.Logf("Display %d status: %s", i, statusLine)
	}

	// Compare each step's expected display against the corresponding 6502 display
	// For multi-character commands, we skip intermediate displays and only check the final one

	// The 6502 always outputs an initial display (display 0).
	// If the first step is "DISPLAY", we check it. Otherwise, we skip it.
	displayIdx := 0
	if len(tc.Steps) > 0 && tc.Steps[0].Commands != "DISPLAY" {
		displayIdx = 1 // Skip the initial auto-display from 6502
	}

	for i, step := range tc.Steps {
		if step.ExpectedDisplay == "" {
			continue // Skip steps with no expected display (like Q command)
		}

		// Count how many character inputs this step will generate
		// This helps us skip intermediate displays for multi-character commands
		numChars := 0
		if step.Commands != "DISPLAY" {
			numChars = len(step.Commands)
		}

		// Skip ahead to the final display for this step
		// For single-char commands, this is just the next display
		// For multi-char commands, we need to skip the intermediate ones
		if numChars > 1 {
			// Skip intermediate displays (numChars - 1)
			displayIdx += numChars - 1
		}

		if displayIdx >= len(displays) {
			t.Errorf("Step %d (%s): No 6502 display available (expected %d displays, got %d)",
				i, step.Commands, displayIdx+1, len(displays))
			continue
		}

		// Normalize both outputs for comparison
		// The 6502 version has different formatting (extra separators, ? prompt)
		expected := normalizeDisplay(step.ExpectedDisplay)
		actual := normalizeDisplay(displays[displayIdx])

		if expected != actual {
			t.Errorf("Step %d (%s): Display mismatch\n=== EXPECTED ===\n%s\n=== ACTUAL (6502) ===\n%s\n=== END ===",
				i, step.Commands, expected, actual)
		} else {
			t.Logf("Step %d (%s): Display matches ✓", i, step.Commands)
		}

		displayIdx++
	}
}

// normalizeDisplay normalizes a display for comparison
// Strips extra separators, prompts, and focuses on actual content
func normalizeDisplay(display string) string {
	var result strings.Builder
	lines := strings.Split(strings.TrimSpace(display), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, separator lines, and prompt lines
		if line == "" || line == "-------------------------" || line == "?" {
			continue
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}

// splitDisplays splits the output into individual display sections
func splitDisplays(output string) []string {
	// Look for the MicroChess header as display boundaries
	var displays []string
	lines := strings.Split(output, "\n")
	var currentDisplay strings.Builder
	inDisplay := false

	for _, line := range lines {
		if strings.Contains(line, "MicroChess") {
			if inDisplay && currentDisplay.Len() > 0 {
				displays = append(displays, currentDisplay.String())
			}
			currentDisplay.Reset()
			inDisplay = true
		}
		if inDisplay {
			currentDisplay.WriteString(line)
			currentDisplay.WriteString("\n")
		}
	}

	if currentDisplay.Len() > 0 {
		displays = append(displays, currentDisplay.String())
	}

	return displays
}
