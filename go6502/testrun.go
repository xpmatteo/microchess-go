// ABOUTME: Test harness to run 6502 programs with I/O support
// ABOUTME: Usage: go run testrun.go iomem.go <program.bin>

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/beevik/go6502/cpu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run testrun.go iomem.go <program.bin>")
		os.Exit(1)
	}

	// Create CPU with I/O-enabled memory
	mem := NewIoMemory()
	c := cpu.NewCPU(cpu.CMOS, mem)

	// Set up signal handling to restore terminal on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Ensure terminal is restored on exit
	defer mem.Restore()

	// Handle signals in goroutine
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, restoring terminal...")
		mem.Restore()
		os.Exit(0)
	}()

	// Load the program
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error loading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	// Load at $1000
	mem.StoreBytes(0x1000, data)

	// Set PC to start
	c.SetPC(0x1000)

	fmt.Printf("Loaded %d bytes at $1000\n", len(data))
	fmt.Println("Running...")
	fmt.Println("---")

	// Run until BRK
	for {
		// Check for BRK instruction
		if mem.LoadByte(c.Reg.PC) == 0x00 {
			break
		}
		c.Step()
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Printf("Program halted at $%04X after %d cycles\n", c.Reg.PC, c.Cycles)
}
