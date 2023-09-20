package vm

import (
	"bytes"
	"fmt"
)

// Interrupt represents the I/O interrupt signal to the CPU. It is as an extremely basic interrupt
// controller.
//
// # Each
//
// There are three conditions that must be satisfied for an device to interrupt and
// change the CPU's control flow:
//
// 1. the device has raised a request;
// 2. the device's interrupt is enabled; and
// 3. the device's priority is greater than the current program (or other ISR).
type Interrupt struct {

	// Interrupt descriptor table. Each priority (PL0 to P7) references a
	// device driver and an 8 bit vector in the interrupt vector table.
	idt [8]struct {
		driver Driver
		vector uint8
	}

	log logger
}

func (i Interrupt) String() string {
	var b bytes.Buffer

	_, _ = b.WriteString("IDT(\n")

	for pl := len(i.idt) - 1; pl >= 0; pl-- {
		id := i.idt[pl]

		fmt.Fprintf(&b, "\t%s:%s", Priority(pl), Word(id.vector))

		if id.driver != nil {
			fmt.Fprintf(&b, ":%s", id.driver)
		}

		fmt.Fprintln(&b)
	}

	fmt.Fprint(&b, ")")

	return b.String()
}

func (i *Interrupt) Register(priority Priority, driver Driver, vector uint8) {
	if descriptor := i.idt[priority]; descriptor.driver != nil {
		// TODO: return error
		i.log.Printf("intr: device priority conflict: want: %s:%s, have: %s:%s",
			priority.String(), driver.String(), priority, descriptor.driver.String(),
		)
	} else {
		descriptor.driver = driver
		descriptor.vector = vector
		i.idt[priority] = descriptor
	}
}

func (i Interrupt) Request(curr Priority) (uint8, bool) {
	for pl := len(i.idt) - 1; pl > int(curr); pl-- {
		idt := i.idt[pl]
		if idt.driver == nil {
			continue
		} else if idt.driver.InterruptRequested() {
			return idt.vector, true
		}
	}

	return 0, false
}

// interruptable errors are returned from instruction cycle steps to signal the CPU to jump to a
// service routine.
type interruptable interface {
	error

	// For the sake of debugging.
	fmt.Stringer

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
//   - computing the address in the interrupt table
//   - fetching the address of a service routine from a table
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

// Trap handler table and defined vectors in the table.
const (
	TrapTable = Word(0x0000)
	TrapHALT  = Word(0x0025)
)

// Interrupt vector table address.
const (
	InterruptVectorTable = Word(0x0100) // IVT
)

// Exception vector table and defined vectors in the table.
const (
	ExceptionServiceRoutines = Word(0x0100) // EXC
	ExceptionPMV             = Word(0x00)   // PMV
	ExceptionXOP             = Word(0x01)   // XOP
	ExceptionACV             = Word(0x02)   // ACV
	// 0x0100:0x017f
)

// Interrupt service routines and defined vectors.
const (
	ISRTable    = Word(0x0180) // ISR
	ISRKeyboard = Word(0x80)   // KBD
)
