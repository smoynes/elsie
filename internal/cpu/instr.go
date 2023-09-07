package cpu

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

// Opcode returns the instruction opcode.
func (i Instruction) Opcode() Opcode {
	return Opcode(i >> 12)
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
	// Some operations have only a source register with no destination,
	// e.g., storage operations, and use the same instruction bits as
	// operations with a destination. The duplication is intentional: we
	// trust the compiler to optimize the indirection and it makes
	// implementations clearer to read.
	return i.DR()
}

// SR1 returns the first register operand from the instruction.
func (i Instruction) SR1() GPR {
	return GPR(i & 0x01a0 >> 6)
}

// SR2 returns the second register operand from the instruction.
func (i Instruction) SR2() GPR {
	return GPR(i & 0x0003)
}

// Imm returns true if the immediate-mode flag is set in the instruction
func (i Instruction) Imm() bool {
	return i&0x020 != 0
}

// Imm5 returns the immediate-mode literal from the instruction. It is a 5-bit
// value sign-extended to a word.
func (i Instruction) Imm5() Word {
	w := Word(i & 0x001f)
	w.Sext(5)

	return w
}

// Offset returns the PC-relative offset from the instruction. It may be a 5-,
// 9-, or 11-bit value sign-extended to a word.
func (i Instruction) Offset(n Offset) Word {
	var w Word

	switch n {
	case PCOFFSET11:
		w = Word(i & 0x03ff)
		w.Sext(11)
	case PCOFFSET9:
		w = Word(i & 0x01ff)
		w.Sext(9)
	case PCOFFSET5:
		w = Word(i & 0x001f)
		w.Sext(5)
	default:
		panic("unexpected offset")
	}

	return w
}

// Vector returns a bit vector from the instruction.
func (i Instruction) Vector(n uint8) Word {
	w := Word(i)
	w.Zext(n)

	return w
}

// Offset identifies the length of a PC-relative offset.
type Offset uint8

// Condition represents a ZNP condition operand from an instruction.
type Condition Word

// Condition flags.
const (
	ConditionPositive Condition = 1 << iota
	ConditionZero
	ConditionNegative
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

// Offset identifier constants.
const (
	PCOFFSET11 Offset = iota
	PCOFFSET9
	PCOFFSET5
)
