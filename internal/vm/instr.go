package vm

import (
	"fmt"
)

// Instruction is special-purpose register that encodes a single CPU operation.
// The top 4 bits of an instruction define the opcode; the remaining bits are
// used for operands and flags.
type Instruction Register

func (i Instruction) String() string {
	return fmt.Sprintf("%s (OP: %s)", Word(i), i.Opcode())
}

// Opcode returns the instruction opcode which is stored in the top four bits of the instruction.
func (i Instruction) Opcode() Opcode {
	return Opcode(i&0xf000) >> 12
}

// Cond returns the condition flags from the instruction.
func (i Instruction) Cond() Condition {
	return Condition(i & 0x0e00 >> 9)
}

// DR returns the destination register ID from the instruction.
func (i Instruction) DR() GPR {
	return GPR(i & 0x0e00 >> 9)
}

// SR returns the source register ID from the instruction.
func (i Instruction) SR() GPR {
	return GPR(i & 0x0e00 >> 9)
}

// SR1 returns the first register operand from the instruction.
func (i Instruction) SR1() GPR {
	return GPR(i & 0x01d0 >> 6)
}

// SR2 returns the second register operand from the instruction.
func (i Instruction) SR2() GPR {
	return GPR(i & 0x0003)
}

// Imm returns true if the immediate-mode flag is set in the instruction
func (i Instruction) Imm() bool {
	return i&0x0020 != 0
}

// Relative returns true if the register-mode flag is set in the instruction.
func (i Instruction) Relative() bool {
	return i&0x0800 != 0
}

// Offset returns the PC-relative offset from the instruction.
func (i Instruction) Offset(n offset) Word {
	w := Word(i)
	w.Sext(uint8(n))

	return w
}

// Literal returns a literal value from the instruction.
func (i Instruction) Literal(n literal) Word {
	w := Word(i)
	w.Sext(uint8(n))

	return w
}

// Vector returns a bit vector from the instruction.
func (i Instruction) Vector(n vector) Word {
	w := Word(i)
	w.Zext(uint8(n))

	return w
}

type (
	offset  uint8
	literal uint8
	vector  uint8
)

const (
	OFFSET11 = offset(11)
	OFFSET9  = offset(9)
	OFFSET6  = offset(6)
	OFFSET5  = offset(5)
	IMM5     = literal(5)
	VECTOR8  = vector(8)
)

// Condition represents a NZP condition operand from an instruction.
type Condition Word

// Condition flags.
const (
	ConditionPositive Condition = 0x1 // P
	ConditionZero     Condition = 0x2 // Z
	ConditionNegative Condition = 0x4 // N
)

func (c Condition) String() string {
	return fmt.Sprintf(
		"%s (N:%t Z:%t P:%t)",
		Word(c).String(), c.Negative(), c.Zero(), c.Positive(),
	)
}

// Negative returns true if the N flag is set.
func (c Condition) Negative() bool {
	return c&ConditionNegative != 0
}

// Zero returns true if the Z flag is set.
func (c Condition) Zero() bool {
	return c&ConditionZero != 0
}

// Positive returns true if the P flag is set.
func (c Condition) Positive() bool {
	return c&ConditionPositive != 0
}
