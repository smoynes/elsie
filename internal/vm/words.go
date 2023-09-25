package vm

// words.go defines the basic data types of the CPU.

import (
	"fmt"
	"strings"

	"github.com/smoynes/elsie/internal/log"
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
	s := 16 - n
	i := int16(*w)
	i <<= s
	i >>= s
	*w = Word(uint16(i))
}

// Zext zero extends the lower n bits in-place.
func (w *Word) Zext(n uint8) {
	var low Word = ^(0xffff << n)
	*w &= low
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

// ProcessStatus is a special-purpose register that records important CPU flags:
// privilege, priority level, and condition flags.
//
// | PR | 000 0 | PL | 0000 0 | COND |
// +----+-------+----+--------+------+
// | 15 |14   12|11 9|8      3|2    0|
type ProcessorStatus Register

// Init configures the device at startup.
func (p *ProcessorStatus) Init(_ *LC3, _ []Word) {
	*p = ProcessorStatus(0x8080)
}

// Get reads the register for I/O.
func (p ProcessorStatus) Get() Register { return Register(p) }

// Put sets the register value for I/O.
func (p *ProcessorStatus) Put(val Register) { *p = ProcessorStatus(val) }

// Status flags in PSR vector.
const (
	StatusPositive  ProcessorStatus = 0x0001
	StatusZero      ProcessorStatus = 0x0002
	StatusNegative  ProcessorStatus = 0x0004
	StatusCondition ProcessorStatus = StatusNegative | StatusZero | StatusPositive

	StatusPriority ProcessorStatus = 0x0700
	StatusHigh     ProcessorStatus = 0x0700
	StatusNormal   ProcessorStatus = 0x0300
	StatusLow      ProcessorStatus = 0x0000

	StatusPrivilege ProcessorStatus = 0x8000
	StatusUser      ProcessorStatus = 0x8000
	StatusSystem    ProcessorStatus = 0x0000
)

func (p ProcessorStatus) String() string {
	return fmt.Sprintf(
		"%s (N:%t Z:%t P:%t PR:%d PL:%d)",
		Word(p), p.Negative(), p.Zero(), p.Positive(), p.Privilege(),
		p.Priority(),
	)
}

// Cond returns the condition codes from the status register.
func (p ProcessorStatus) Cond() Condition {
	return Condition(p & StatusCondition)
}

// Any returns true if any of the flags in the condition are set in the status
// register.
func (c ProcessorStatus) Any(cond Condition) bool {
	return c.Cond()&cond != 0
}

// Set sets the condition flags based on the zero, negative, and
// positive attributes of the register value.
func (c *ProcessorStatus) Set(reg Register) {
	// Clear condition flags.
	*c &= ^StatusCondition

	// Set condition flag from register sign.
	switch {
	case reg == 0:
		*c |= StatusZero
	case int16(reg) > 0:
		*c |= StatusPositive
	default:
		*c |= StatusNegative
	}
}

// Positive returns true if the P flag is set.
func (c ProcessorStatus) Positive() bool {
	return c&StatusPositive != 0
}

// Negative returns true if the N flag is set.
func (c ProcessorStatus) Negative() bool {
	return c&StatusNegative != 0
}

// Zero returns true if the Z flag is set.
func (c ProcessorStatus) Zero() bool {
	return c&StatusZero != 0
}

// Priority returns the priority level of the current task.
func (c ProcessorStatus) Priority() Priority {
	return Priority(c & StatusPriority >> 8)
}

func (c *ProcessorStatus) device() string { return Register(*c).String() }

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

// Privilege returns the privilege of the current task.
func (c ProcessorStatus) Privilege() Privilege {
	return Privilege(c & StatusPrivilege >> 15)
}

// Privilege represents the privilege level of a task.
type Privilege uint8

// Privilege levels.
const (
	PrivilegeSystem Privilege = iota // System
	PrivilegeUser                    // User
)

// RegisterFile is the set of general purpose registers.
type RegisterFile [NumGPR]Register

func (rf RegisterFile) String() string {
	b := strings.Builder{}
	for i := 0; i < len(rf)/2; i++ {
		fmt.Fprintf(&b, "R%d:  %s R%d: %s\n",
			i, rf[i], i+len(rf)/2, rf[i+len(rf)/2])
	}

	return b.String()
}

func (rf RegisterFile) LogValue() log.Value {
	return log.GroupValue(
		log.String("R0", rf[R0].String()),
		log.String("R1", rf[R1].String()),
		log.String("R2", rf[R2].String()),
		log.String("R3", rf[R3].String()),
		log.String("R4", rf[R4].String()),
		log.String("R5", rf[R5].String()),
		log.String("R6", rf[R6].String()),
		log.String("R7", rf[R7].String()),
	)
}

// GPR is the ID of a general purpose register
type GPR uint8

// General purpose registers.
const (
	R0 GPR = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7

	// NumGPR is the count of general purpose registers.
	NumGPR

	// Subroutine return address is in R7
	RETP GPR = R7

	// Current stack is in R6.
	SP GPR = R6
)

// ControlRegister is the master control register.
type ControlRegister Register

const (
	ControlRunning ControlRegister = 1 << 15
)

func (c ControlRegister) Running() bool {
	return c&ControlRunning != 0
}

func (c *ControlRegister) String() string {
	run := "RUN"
	if !c.Running() {
		run = "STOP"
	}

	return fmt.Sprintf("%s (%s)", Register(*c).String(), run)
}

// Init configures the device at startup.
func (p *ControlRegister) Init(_ *LC3, _ []Word) {
	*p = 0x8080
}

// Get returns the register value for I/O.
func (c *ControlRegister) Get() Register {
	return Register(*c)
}

// Put sets the register value for I/O.
func (c *ControlRegister) Put(val Register) {
	*c = ControlRegister(val)
}

func (c *ControlRegister) device() string {
	return "MCR(ELSIE LC-3 SIMULATOR)"
}
