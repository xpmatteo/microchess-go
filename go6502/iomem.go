// ABOUTME: Custom memory implementation with simple console I/O support
// ABOUTME: Intercepts reads/writes to $FFF0 (output) and $FFF1 (input)

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/beevik/go6502/cpu"
)

// IoMemory wraps FlatMemory and adds I/O at specific addresses
type IoMemory struct {
	mem       *cpu.FlatMemory
	reader    *bufio.Reader
	lastInput byte
}

// NewIoMemory creates a new memory with I/O support
func NewIoMemory() *IoMemory {
	return &IoMemory{
		mem:    cpu.NewFlatMemory(),
		reader: bufio.NewReader(os.Stdin),
	}
}

// LoadByte loads a byte, with I/O at $FFF1
func (m *IoMemory) LoadByte(addr uint16) byte {
	// $FFF1 - Character input (blocking read from stdin)
	if addr == 0xFFF1 {
		if m.lastInput == 0 {
			char, err := m.reader.ReadByte()
			if err != nil {
				return 0
			}
			m.lastInput = char
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
