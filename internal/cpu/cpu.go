// Package cpu provides an emulated CPU.
package cpu

import (
	"fmt"
	"strings"
)

// LC3 is a computer simulated in software.
type LC3 struct {

	// Program Counter. It is a pointer to the next instruction to execute.
	PC ProgramCounter

	// Instruction register.
	IR Instruction

	// Condition register.
	Cond Condition

	// General purpose registers.
	Reg RegisterFile

	// Addressable memory.
	Mem Memory
}

func New() *LC3 {
	cpu := LC3{}
	cpu.PC = 0x3000
	cpu.Cond = ConditionZero
	return &cpu
}

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC: %s IR: %s COND: %s", cpu.PC, cpu.IR, cpu.Cond)
}

// Words are the base size of data at which the CPU operates. Registers, memory
// cells, I/O and instructions all work on 16-bit values.
type Word uint16

func (w Word) String() string {
	return fmt.Sprintf("%0#4x", uint16(w))
}

// Sext sign-extends the lower n bits in place.
func (w *Word) Sext(n uint8) {
	if n > 15 {
		panic("n >= 16")
	}
	ans := int16(*w) << (16 - n) >> (16 - n)
	*w = Word(uint16(ans))
}

// Registers are used by the CPU to store values for an operation.
type Register Word

func (r Register) String() string {
	return Word(r).String()
}

// ProgramCounter is a special-purpose register that points to the next instruction in memory.
type ProgramCounter Register

func (p ProgramCounter) String() string {
	return Word(p).String()
}

// Condition is a special-purpose register that it not directly usable by
// programs. It is a 3-bit vector {Z, N, P} that stores the zero, negative and
// positive properties, respectively; only the bottom bits of the register are
// used.
type Condition Register

const (
	ConditionPositive Condition = 1 << iota
	ConditionZero
	ConditionNegative
)

func (c Condition) String() string {
	i := uint16(c)
	return fmt.Sprintf("%0#1x (P:%t N:%t Z:%t)",
		i, c.Positive(), c.Negative(), c.Zero())
}

func (c *Condition) Update(reg Register) {
	switch {
	case reg == 0:
		*c = ConditionZero
	case reg&0x8000 == 0:
		*c = ConditionPositive
	case reg&0x8000 != 0:
		*c = ConditionNegative
	default:
		panic("unreachable")
	}
}

func (c Condition) Positive() bool {
	return c&ConditionPositive != 0
}

func (c Condition) Negative() bool {
	return c&ConditionNegative != 0
}

func (c Condition) Zero() bool {
	return c&ConditionZero != 0
}

// Instruction is a 16-bit value that encodes a single CPU operation. The LS-3
// ISA has 15 distinct instructions (and one reserved value that is undefined).
// The top 4 bits of an instruction define the opcode; the remaining bits are
// used for operands and flags.
type Instruction Register

func (i Instruction) String() string {
	return fmt.Sprintf("%0#4x (OP: %s)", Word(i), i.Opcode())
}

func (i Instruction) Opcode() Opcode {
	return Opcode(i >> 12)
}

// Set of general purpose registers.
type RegisterFile [NumRegisters]Register

func (rf *RegisterFile) String() string {
	b := strings.Builder{}
	for i := 0; i < len(rf)/2; i++ {
		fmt.Fprintf(&b, "R%d: %s\tR%d: %s\n",
			i, rf[i], i+len(rf)/2, rf[i+len(rf)/2])
	}
	return b.String()
}

// GPR is the ID of a general purpose register
type GPR uint8

const (
	// General purpose registers.
	R0 GPR = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7

	// Count of general purpose registers.
	NumRegisters
)

func (r GPR) String() string {
	return Register(r).String()
}
