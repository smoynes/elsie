package vm

import (
	"bytes"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// Interrupt represents the I/O interrupt signal to the CPU. It is as an extremely basic interrupt
// controller.
//
// There are three conditions that must be satisfied for an device to interrupt and change the CPU's
// control flow:
//
// 1. the device has raised a request;
// 2. the device's interrupt is enabled; and
// 3. the device's priority is greater than the current program (or other ISR).
type Interrupt struct {

	// Interrupt descriptor table. Each priority (PL0 to P7) references a
	// device driver and the interrupt's vector.
	idt [8]ISR

	log *log.Logger
}

// ISR is an interrupt service routine. It contains the interrupt's vector and a reference to the
// driver for the device that requests service.
type ISR struct {
	vector uint8
	driver Driver
}

func (isr ISR) String() string {
	return fmt.Sprintf("ISR{%0#2x:%s}", isr.vector, isr.driver.String())
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

// Register assigns an interrupt priority to a service routine.
func (i *Interrupt) Register(priority Priority, isr ISR) {
	if entry := i.idt[priority]; entry.driver != nil {
		// TODO: return error
		i.log.Error("intr: device priority conflict: want: %s:%s, have: %s:%s",
			priority.String(), isr.String(), priority.String(), entry.String(),
		)
	} else {
		entry.driver = isr.driver
		entry.vector = isr.vector
		i.idt[priority] = entry
	}
}

func (i Interrupt) Requested(curr Priority) (uint8, bool) {
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
	return fmt.Sprintf("INT: (%s:%0#2x)", i.table, uint16(i.vec))
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
