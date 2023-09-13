// Package cpu provides an emulated CPU.
package cpu

import (
	"fmt"
)

// LC3 is a computer simulated in software.
type LC3 struct {
	PC  ProgramCounter  // Instruction Pointer
	IR  Instruction     // Instruction Register
	PSR ProcessorStatus // Processor Status Register
	Reg RegisterFile    // General-purpose register file
	USP Register        // User stack pointer
	SSP Register        // System stack pointer
	MCR Register        // Master control register
	Mem Memory          // All the memory you'll ever need.
}

// New initializes a minimal virtual machine state.
func New() *LC3 {
	// Set CPU registers to known values.
	cpu := LC3{
		PC:  0x0300,
		IR:  0x0000,
		PSR: initialStatus,
		USP: Register(IOPageAddr),    // User stack grows down from the top of user space.
		SSP: Register(UserSpaceAddr), // Similarly, system stack starts where user's end.
		MCR: Register(0x8000),        // Set the RUN flag.
	}

	// Initialize general purpose registers to a pleasing pattern.
	copy(cpu.Reg[:], []Register{
		0xffff, 0x0000,
		0xfff0, 0xf000,
		0xff00, 0x0f00,
		0xf000, 0x00f0,
	})
	cpu.Reg[SP] = cpu.USP

	// Configure MMU.
	cpu.Mem = NewMemory(&cpu.PSR)

	kbd := Keyboard{}
	disp := Display{}

	// Map CPU registers into address space.
	err := cpu.Mem.device.Map(MMIO{
		MCRAddr:  &cpu.MCR,
		PSRAddr:  &cpu.PSR,
		KBSRAddr: &kbd.status,
		KBDRAddr: &kbd.data,
		DSRAddr:  &disp.status,
		DDRAddr:  &disp.data,
	})
	if err != nil {
		panic(err)
	}

	return &cpu
}

// initial value of PSR at boot is undefined. At least, I haven't found it in the ISA reference.
// We'll start the machine with system privileges, normal priority, and with all condition flags
// set.
const initialStatus = ProcessorStatus(StatusSystem | StatusNormal | StatusCondition)

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC:  %s IR: %s \nPSR: %s\nUSP: %s SSP: %s MCR: %s\n"+
		"MAR: %s MDR: %s\n",
		cpu.PC, cpu.IR, cpu.PSR, cpu.USP, cpu.SSP, cpu.MCR,
		cpu.Mem.MAR, cpu.Mem.MDR)
}

// PushStack pushes a word onto the current stack.
func (cpu *LC3) PushStack(w Word) error {
	cpu.Reg[SP]--
	cpu.Mem.MAR = cpu.Reg[SP]
	cpu.Mem.MDR = Register(w)
	return cpu.Mem.Store()
}

// PopStack pops a word from the current stack into MDR.
func (cpu *LC3) PopStack() error {
	cpu.Reg[SP]++
	cpu.Mem.MAR = cpu.Reg[SP] - 1
	return cpu.Mem.Fetch()
}
