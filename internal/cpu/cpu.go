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
		PSR: initialStatus,
	}
	cpu.Mem = NewMemory(&cpu.PSR)

	return &cpu
}

// initial value of PSR at boot is seemingly undefined. At least, I haven't
// found it in the ISA reference. Starts with system privileges, normal
// priority, and with condition flags set.
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
