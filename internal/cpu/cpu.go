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
	Mem Memory          // All the memory you'll ever need
}

func New() *LC3 {
	cpu := LC3{
		PC:  0x0300,
		IR:  0x0000,
		PSR: initialStatus,
		USP: Register(IOPageAddr),    // User stack grows down from the top of user space.
		SSP: Register(UserSpaceAddr), // Similarly, system stack starts where users end.
	}
	cpu.Mem = NewMemory(&cpu.PSR)
	cpu.Reg[SP] = Register(UserSpaceAddr)

	return &cpu
}

// initial value of PSR at boot is undefined. At least, I haven't found it in the ISA reference.
// We'll start the machine with system privileges, normal priority, and with all condition flags
// set.
const initialStatus = ProcessorStatus(StatusSystem | StatusNormal | StatusCondition)

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC: %s IR: %s PSR: %s USP: %s SSP: %s",
		cpu.PC, cpu.IR, cpu.PSR, cpu.USP, cpu.SSP)
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
