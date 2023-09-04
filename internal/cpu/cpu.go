// Package cpu provides an emulated CPU.
package cpu

import (
	"fmt"
	"math"
	"strings"
)

// LC3 is a computer simulated in software.
type LC3 struct {

	// Program Counter. It is a pointer to the next instruction to execute.
	PC ProgramCounter

	// Instruction register.
	IR Instruction

	// Processing unit.
	Proc Processor

	// Addressable memory.
	Mem Memory
}

func New() *LC3 {
	cpu := LC3{}
	cpu.PC = 0x3000
	cpu.Proc.Cond = ConditionZero
	return &cpu
}

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC: %s IR: %s COND: %s", cpu.PC, cpu.IR, cpu.Proc.Cond)
}

// Words are the base size of data at which the CPU operates. Registers, memory
// cells, I/O and instructions all work on 16-bit values.
type Word uint16

// Registers are used by the CPU to store values for an operation.
type Register Word

func (r Register) String() string {
	return fmt.Sprintf("%0#4x", uint16(r))
}

// ProgramCounter is a special-purpose register that points to the next instruction in memory.
type ProgramCounter Register

func (p ProgramCounter) String() string {
	return Register(p).String()
}

// Processor is the processing unit of the CPU.
type Processor struct {
	// Condition register.
	Cond Condition

	// General purpose registers.
	Reg RegisterFile
}

// Condition is a special-purpose register that it not directly usable by
// programs. It stores the positive, negative, and zero properties of computed
// values, only one of which may be true. The Condition register represents the
// value as 3-bit vector of {P, N, Z}; only the bottom three bits of the
// register are used.
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

func (c *Condition) Update(a Word) {
	switch {
	case a == 0:
		*c = ConditionZero
	case a&0x8000 == 0:
		*c = ConditionPositive
	case a&0x8000 != 0:
		*c = ConditionNegative
	default:
		panic("no")
	}
}

func (c Condition) Positive() bool {
	return c&ConditionPositive > 0x0
}

func (c Condition) Negative() bool {
	return c&ConditionNegative > 0
}

func (c Condition) Zero() bool {
	return c&ConditionZero > 0
}

// Set of general purpose registers.
type RegisterFile [NumRegisters]Register

func (rf *RegisterFile) String() string {
	b := strings.Builder{}
	for i := 0; i < len(rf)/2; i++ {
		fmt.Fprintf(&b, "R%d: %s\tR%d: %s\n", i, rf[i], i+len(rf)/2, rf[i+len(rf)/2])
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

// Number of memory addresses: 2^16 values, i.e. {0x0000, ..., 0xFFFF}
const AddressSpace = math.MaxUint16

// Addressable Memory: 16-bit words with an address space of 16 bits.
type Memory [AddressSpace]Word

func (mem *Memory) Load(addr Word) Word {
	return mem[addr]
}
