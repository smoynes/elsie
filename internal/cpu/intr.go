package cpu

import (
	"fmt"
)

// interruptable errors are returned from instruction cycle steps to signal the
// CPU to jump to a service routine.
type interruptable interface {
	error

	// Handle changes the execution context to the interrupt's service
	// routine.
	Handle(cpu *LC3) error
}

// interrupts change the control flow of the CPU.
//
// Includes:
//   - access control (ACV)
//   - privilege mode (PMV)
//   - illegal opcode (XOP)
//   - trap           (TRAP)
//   - I/O interrupt  (IO)
//
// In each case, execution continues after:
//
//   - pushing the caller's PC and PSR to the stack
//
//   - computing the address in the interrupt table
//
//   - fetching the address of a service routine from a table
//
//   - jumping to the service routine address
type interrupt struct {
	table Word            // Address of vector table.
	vec   Word            // Vector in interrupt vector table.
	pc    ProgramCounter  // Program counter of the caller.
	psr   ProcessorStatus // Status register of the caller.
}

func (i *interrupt) Handle(cpu *LC3) error {
	err := cpu.PushStack(Word(i.psr))
	if err != nil {
		return err
	}

	err = cpu.PushStack(Word(i.pc))
	if err != nil {
		return err
	}

	cpu.Mem.MAR = Register(i.table | i.vec)
	err = cpu.Mem.Fetch()
	if err != nil {
		return err
	}

	cpu.PC = ProgramCounter(Word(cpu.Mem.MDR))

	return nil
}

func (i *interrupt) Error() string {
	return "interrupted: " + i.String()
}

func (i *interrupt) String() string {
	return fmt.Sprintf("INT: (%s:%s)", i.table, i.vec)
}

// Exception vector table and defined vectors in the table.
const (
	ExceptionTable = Word(0x0100)
	ExceptionPMV   = Word(0x00)
	ExceptionXOP   = Word(0x01)
	ExceptionACV   = Word(0x02)
)

// Trap handler table and defined vectors in the table.
const (
	TrapTable = Word(0x0000)
	TrapHALT  = Word(0x0025)
)
