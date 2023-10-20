package vm

import (
	"bytes"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// Interrupt represents the I/O interrupt signal to the CPU. It is an extremely basic interrupt
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
	idt [NumPL]ISR

	log *log.Logger
}

func (i Interrupt) LogValue() log.Value {
	var as []log.Attr

	for i, isr := range i.idt {
		pl := Priority(i)
		if isr.driver != nil {
			as = append(as, log.String(pl.String(), isr.String()))
		}
	}

	return log.GroupValue(as...)
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

// An interruptableError is returned from an instruction cycle to signal the CPU to jump to an
// interrupt service routine.
type interruptableError interface {
	error

	// Handle changes the execution context to the interrupt's service routine.
	Handle(cpu *LC3) error

	fmt.Stringer // For the sake of debugging.
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

func (intr *interrupt) Handle(cpu *LC3) error {
	err := cpu.PushStack(Word(intr.psr))
	if err != nil {
		return err
	}

	err = cpu.PushStack(Word(intr.pc))
	if err != nil {
		return err
	}

	cpu.Mem.MAR = Register(intr.table | intr.vec)
	err = cpu.Mem.Fetch()

	if err != nil {
		return err
	}

	cpu.PC = ProgramCounter(Word(cpu.Mem.MDR))

	return nil
}

func (intr *interrupt) Is(err any) bool {
	if _, ok := err.(*interrupt); ok {
		return true
	}

	return false
}

func (intr *interrupt) As(err any) bool {
	if err, ok := err.(**interrupt); ok {
		if err != nil {
			*err = intr
		}

		return true
	}

	return false
}

func (*interrupt) Error() string {
	return "interrupt"
}

func (intr *interrupt) String() string {
	return fmt.Sprintf("INT: (%s:%0#2x)", intr.table, uint16(intr.vec))
}

// acv is a memory access control violation exception.
type acv struct {
	*interrupt
}

func (ae *acv) Is(target error) bool {
	switch target.(type) {
	case *acv, *interrupt:
		return true
	default:
		return false
	}
}

func (ae *acv) As(target any) bool {
	switch err := target.(type) {
	case **acv:
		if *err != nil {
			*err = ae
		}

		return true
	case **interrupt:
		if *err != nil {
			*err = ae.interrupt
		}

		return true
	default:
		return false
	}
}

func (*acv) Error() string {
	return "acv error"
}

func (ae *acv) String() string {
	return fmt.Sprintf("EXC: ACV (%s:%0#2x)", ae.table, ae.vec)
}

// Trap handler table and defined vectors in the table.
const (
	TrapTable = Word(0x0000) // TRAPs (0x0000:0x00ff)
	TrapOUT   = Word(0x21)
	TrapHALT  = Word(0x25)
)

// Interrupt service routine table and defined service routines.
const (
	ISRTable    = Word(0x0100) // IVT (0x0100:0x01ff)
	ISRKeyboard = Word(0x80)   // KBD
)

// Exception vector table and defined vectors in the table.
const (
	// 0x0100:0x017f
	ExceptionServiceRoutines = Word(0x0100) // EXC
	ExceptionPMV             = Word(0x00)   // PMV
	ExceptionXOP             = Word(0x01)   // XOP
	ExceptionACV             = Word(0x02)   // ACV
)
