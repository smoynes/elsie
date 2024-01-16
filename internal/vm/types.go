package vm

// types.go defines the basic data types of the CPU.

import (
	"fmt"
)

// Word is the base data type on which the CPU operates. Registers, memory
// cells, I/O and instructions all work on 16-bit values.
type Word uint16

func (w Word) String() string {
	return fmt.Sprintf("%0#4x", uint16(w))
}

// Sext sign-extends the lower n bits in-place.
func (w *Word) Sext(n uint8) {
	// Maybe this deserves an explanation. 😬
	//
	// Tersely, given
	//
	//    i := int16(0b0000_0000_0000_1010)
	//    n := uint8(4)
	//
	// Then
	//
	//    s := 16 - 4 // 12
	//    i = i << s  // 0b1010_0000_0000_0000
	///   i = i >> s  // 0b1111_1111_1111_1010
	//
	// So the bottom n bits in i are sign extended.
	//
	// More verbosely, to sign extend the bottom n bits of an integer, we use two
	// shift operations. First, the left shift operator moves the n-th bit
	// left to the most significant of the word. That puts the sign bit of
	// the initial lower n bits in the sign position of the word. Next, the
	// word is shifted rightwards, back to the original position. The right
	// shift extends the sign bit across the top bits and gives us our
	// result. The int64 and uint64 conversions are needed because Go's
	// right shift operator only extends signed integers.
	i := int16(*w)
	i <<= 16 - n
	i >>= 16 - n
	*w = Word(i)
}

// Zext zero extends the lower n bits in-place.
func (w *Word) Zext(n uint8) {
	low := Word(^(int16(-1) << n))
	*w &= low
}

// Registers are used by the CPU to store values for an operation.
type Register Word

func (r Register) String() string {
	return Word(r).String()
}

// Offset adds an offset value to a register. The offset is taken as a
func (r *Register) Offset(offset Word) {
	val := Word(*r) + offset
	*r = Register(val)
}

// Instruction is a value that encodes a single CPU operation and is stored in a special purpose
// register. The top 4 bits of an instruction define the opcode; the remaining bits are used for
// operands and flags.
type Instruction Word

// NewInstruction creates a new instruction value for the given opcode.
func NewInstruction(opcode Opcode, operands uint16) Instruction {
	val := uint16(opcode) << 12
	val |= operands & 0x0fff

	return Instruction(val)
}

func (i Instruction) String() string {
	return fmt.Sprintf("%s (OP: %s)", Word(i), i.Opcode())
}

// Operand applies
func (i *Instruction) Operand(operand uint16) {
	*i |= Instruction(operand) & 0x0fff
}

// Encode returns the instruction as a word.
func (i Instruction) Encode() Word {
	return Word(i)
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

// Literal returns a literal n-bit, sign-extended value from the instruction.
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

// Priority represents the priority level of a task.
type Priority uint8

// Task and interrupt priorities.
const (
	PL0 Priority = iota
	PL1
	PL2
	PL3
	PL4
	PL5
	PL6
	PL7
	NumPL

	PriorityLOW    Priority = 0x00 // NORM
	PriorityNormal Priority = 0x03 // NORM
	PriorityHigh   Priority = 0x07 // HIGH
)

// Privilege represents the privilege level of a task.
type Privilege uint8

// Privilege levels.
const (
	PrivilegeSystem Privilege = iota // System
	PrivilegeUser                    // User
)

// GPR is the ID of a general purpose register
type GPR uint8

// General purpose registers.
const (
	R0 = GPR(iota)
	R1
	R2
	R3
	R4
	R5
	R6
	R7

	NumGPR             // Count of general purpose registers.
	SP     = R6        // Current stack is in R6.
	RETP   = R7        // Subroutine return address is in R7.
	BadGPR = GPR(0xff) // Invalid sentinel value.

)

// ControlRegister is the master control register.
type ControlRegister Register

const (
	// ControlRunning is the bit in the control register which, if true, lets the machine continue
	// computing; if false, the machine before executing the next instruction.
	ControlRunning ControlRegister = 1 << 15
)

func (cr ControlRegister) Running() bool {
	return cr&ControlRunning != 0
}

func (cr *ControlRegister) String() string {
	run := "RUN"
	if !cr.Running() {
		run = "STOP"
	}

	return fmt.Sprintf("%s (%s)", Register(*cr).String(), run)
}

// Init configures the device at startup.
func (cr *ControlRegister) Init(_ *LC3, _ []Word) {
	*cr = 0x8080
}

// Get returns the register value for I/O.
func (cr *ControlRegister) Get() Register {
	return Register(*cr)
}

// Put sets the register value for I/O.
func (cr *ControlRegister) Put(val Register) {
	*cr = ControlRegister(val)
}

func (cr *ControlRegister) device() string {
	return "MCR(𝔼𝕃𝕊𝕀𝔼 LC-3 SIMULATOR)"
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
type Condition uint8

// Condition flags.
const (
	ConditionPositive = Condition(1 << iota) // P
	ConditionZero                            // Z
	ConditionNegative                        // N
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
