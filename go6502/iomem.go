// ABOUTME: Custom memory implementation with simple console I/O support
// ABOUTME: Intercepts reads/writes to $FFF0 (output) and $FFF1 (input)

package main

import (
	"fmt"
	"os"

	"github.com/beevik/go6502/cpu"
	"golang.org/x/term"
)

// IoMemory wraps FlatMemory and adds I/O at specific addresses
type IoMemory struct {
	mem       *cpu.FlatMemory
	oldState  *term.State
	lastInput byte
}

// NewIoMemory creates a new memory with I/O support and switches to raw terminal mode
func NewIoMemory() *IoMemory {
	// Switch stdin to raw mode for character-by-character input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not set raw terminal mode: %v\n", err)
	}

	return &IoMemory{
		mem:      cpu.NewFlatMemory(),
		oldState: oldState,
	}
}

// Restore returns the terminal to its original state
func (m *IoMemory) Restore() {
	if m.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), m.oldState)
	}
}

// LoadByte loads a byte, with I/O at $FFF1
func (m *IoMemory) LoadByte(addr uint16) byte {
	// $FFF1 - Character input (blocking read from stdin, one byte at a time)
	if addr == 0xFFF1 {
		if m.lastInput == 0 {
			buf := make([]byte, 1)
			_, err := os.Stdin.Read(buf)
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
func (m *IoMemory) LoadBytes(addr uint16, b []byte) {
	m.mem.LoadBytes(addr, b)
}

// LoadAddress passes through to underlying memory
func (m *IoMemory) LoadAddress(addr uint16) uint16 {
	return m.mem.LoadAddress(addr)
}

// StoreByte stores a byte, with I/O at $FFF0
func (m *IoMemory) StoreByte(addr uint16, v byte) {
	// $FFF0 - Character output (write to stdout)
	if addr == 0xFFF0 {
		fmt.Printf("%c", v)
		return
	}
	m.mem.StoreByte(addr, v)
}

// StoreBytes passes through to underlying memory
func (m *IoMemory) StoreBytes(addr uint16, b []byte) {
	m.mem.StoreBytes(addr, b)
}

// StoreAddress passes through to underlying memory
func (m *IoMemory) StoreAddress(addr uint16, v uint16) {
	m.mem.StoreAddress(addr, v)
}
