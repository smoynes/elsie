// Package cpu provides an emulated CPU.
package cpu

import (
	"fmt"
)

// LC3 is a computer simulated in software.
type LC3 struct {
	PC   ProgramCounter  // Instruction Pointer
	IR   Instruction     // Instruction Register
	PSR  ProcessorStatus // Processor Status Register
	Reg  RegisterFile    // General-purpose register file
	Mem  Memory          // All the memory you'll ever need
	USP  Register        // User stack pointer
	SSP  Register        // System stack pointer
	Temp Register        // Temporary value
}

func New() *LC3 {
	cpu := LC3{
		PC:  0x0300,
		PSR: initialStatus,
	}

	return &cpu
}

// initial value of PSR at boot is seemingly undefined. At least, I haven't
// found it in the ISA reference. Set all condition flags, an invalid value, to
// make it a bit more obvious that the conditions are uninitialized when
// debugging.
const initialStatus = ProcessorStatus(
	Word(PrivilegeSystem) | Word(PriorityNormal) | Word(StatusCondition),
)

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC: %s IR: %s PSR: %s", cpu.PC, cpu.IR, cpu.PSR)
}

// PushStack pushes a word onto the current stack.
func (cpu *LC3) PushStack(w Word) {
	cpu.Reg[SP]--
	cpu.Mem.MAR = cpu.Reg[SP]
	cpu.Mem.MDR = Register(w)
	cpu.Mem.Store()
}

// PopStack pops a word from the current stack into a register.
func (cpu *LC3) PopStack() Word {
	cpu.Reg[SP]++
	cpu.Mem.MAR = cpu.Reg[SP] - 1
	cpu.Mem.Fetch()
	return Word(cpu.Mem.MDR)
}
