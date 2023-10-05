package vm

// exec.go defines the CPU instruction cycle.

import (
	"context"
	"errors"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// Run starts and executes the instruction cycle until the program halts.
func (vm *LC3) Run(ctx context.Context) error {
	var err error

	vm.log.Info("START", log.Group("STATE", vm))

	for {
		if err := ctx.Err(); err != nil {
			return err
		} else if !vm.MCR.Running() {
			break
		}

		err = vm.Step()
		if err != nil {
			break
		}

		vm.log.Info("EXEC", log.Group("STATE", vm))

		err = vm.serviceInterrupts()
		if err != nil {
			break
		}
	}

	vm.log.Info("HALTED (HCF)", log.Group("STATE", vm))

	return err
}

// serviceInterrupts invokes the highest priority interrupt service routine, if any.
func (vm *LC3) serviceInterrupts() error {
	if vec, intr := vm.INT.Requested(vm.PSR.Priority()); intr {
		isr := &interrupt{
			table: ISRTable,
			vec:   Word(vec), // TODO: change type to uint8?
			pc:    vm.PC,
			psr:   vm.PSR,
		}

		vm.log.Debug("interrupt requested", "ISR", isr)

		if err := isr.Handle(vm); err != nil {
			// TODO: Handle error.
			return fmt.Errorf("int: %w", err)
		}
	}

	return nil
}

// Step runs a single instruction to completion.
//
// Each operation has as many as six steps:
//
//   - fetch instruction: using the program counter as a pointer, fetch an
//     instruction from memory into the instruction register and increment the
//     program counter.
//
//   - decode operation: get instruction from the instruction register.
//
//   - evaluate address: compute the memory address to be accessed.
//
//   - fetch operands: load an operand from memory using the computed address.
//
//   - execute operation: do the thing.
//
//   - store result: store operation result in memory using the computed
//     address.
//
// An instruction implements methods according to its operational semantics; see [operation].
func (vm *LC3) Step() error {
	if err := vm.Fetch(); err != nil {
		return fmt.Errorf("ins: %w", err)
	}

	op := vm.Decode()
	vm.EvalAddress(op)
	vm.FetchOperands(op)
	vm.Execute(op)
	vm.StoreResult(op)

	err := op.Err()

	if err == nil {
		vm.log.Debug("executed instruction", "OP", op)

		return nil
	} else if errors.Is(err, &interrupt{}) {
		handler := err.(interruptableError) //nolint:errorlint

		vm.log.Error("instruction raised interrupt", "OP", op, "INT", err, "HANDLE", handler)

		if err := handler.Handle(vm); err != nil {
			vm.log.Error("interrupt service routine error", "ERR", err)
			return fmt.Errorf("step: %w", err)
		}

		return nil
	} else { // err != nil
		vm.log.Error("instruction error", "OP", op, "ERR", err)

		return fmt.Errorf("ins: %w", err)
	}
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (vm *LC3) Fetch() error {
	vm.Mem.MAR = Register(vm.PC)

	if err := vm.Mem.Fetch(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	vm.IR = Instruction(vm.Mem.MDR)
	vm.PC++

	vm.log.Debug("fetched", "IR", vm.IR)

	return nil
}

// Decode the instruction from IR.
func (vm *LC3) Decode() operation {
	var oper operation

	switch vm.IR.Opcode() {
	case BR:
		oper = &br{}
	case AND:
		if vm.IR.Imm() {
			oper = &andImm{}
		} else {
			oper = &and{}
		}
	case ADD:
		if vm.IR.Imm() {
			oper = &addImm{}
		} else {
			oper = &add{}
		}
	case NOT:
		oper = &not{}
	case LD:
		oper = &ld{}
	case LDI:
		oper = &ldi{}
	case LDR:
		oper = &ldr{}
	case LEA:
		oper = &lea{}
	case ST:
		oper = &st{}
	case STI:
		oper = &sti{}
	case STR:
		oper = &str{}
	case JMP, RET:
		oper = &jmp{}
	case JSR, JSRR:
		if vm.IR.Relative() {
			oper = &jsr{}
		} else {
			oper = &jsrr{}
		}
	case TRAP:
		oper = &trap{}
	case RTI:
		oper = &rti{}
	case RESV:
		oper = &resv{}
	}

	oper.Decode(vm)

	vm.log.Debug("decoded", "OP", oper)

	return oper
}

// EvalAddress computes a relative memory address if the operation is
// addressable.
func (vm *LC3) EvalAddress(op operation) {
	if op, ok := op.(addressable); ok && op.Err() == nil {
		op.EvalAddress()
		vm.log.Debug("eval", "OP", op)
	}
}

// FetchOperands reads from memory into a CPU register if the operation is fetchable.
func (vm *LC3) FetchOperands(op operation) {
	if op, ok := op.(fetchable); ok && op.Err() == nil {
		if err := vm.Mem.Fetch(); err != nil {
			err = &acv{
				&interrupt{
					table: 0x01,
					vec:   0x02,
					pc:    vm.PC,
					psr:   vm.PSR,
				},
			}

			op.Fail(err)

			return
		}

		op.FetchOperands()
		vm.log.Debug("fetched", "OP", op.String())
	}
}

// Execute does the operation.
func (vm *LC3) Execute(op operation) {
	if op, ok := op.(executable); ok && op.Err() == nil {
		op.Execute()
		vm.log.Debug("exec", log.String("OP", op.String()))
	}
}

// StoreResult writes registers to memory if the operation is storable.
func (vm *LC3) StoreResult(op operation) {
	if op, ok := op.(storable); ok && op.Err() == nil {
		op.StoreResult() // Can't fail.

		if err := vm.Mem.Store(); err != nil {
			err = &acv{
				&interrupt{
					table: 0x01,
					vec:   0x02,
					pc:    vm.PC,
					psr:   vm.PSR,
				},
			}

			op.Fail(err)

			return
		}

		vm.log.Debug("stored", "OP", op.String())
	}
}

// An operation represents a single CPU instruction as it is being executed by
// the machine. The instruction's semantics are defined by implementing optional
// interfaces for each execution stage: [addressable], [fetchable], [executable],
// [storable].
type operation interface {
	// Decode initializes the operation from the machine's instruction
	// pointer.
	Decode(vm *LC3) // TODO: remove argument

	// Fail signals that an error occurred during execution. After it is
	// called with an error, the remaining steps of the operation are
	// skipped.
	Fail(err error)

	// Err returns the error when the instruction cannot continue execution.
	Err() error

	// Stringer for dabugs.
	fmt.Stringer
}

// addressable operations set the memory address register.
type addressable interface {
	operation
	EvalAddress()
}

// fetchable operations load operands from the memory data registers.
type fetchable interface {
	addressable
	FetchOperands()
}

// executable operations update CPU state. Some instructions do not, surprisingly.
type executable interface {
	operation
	Execute()
}

// storable operations store values to memory.
type storable interface {
	addressable

	// StoreResult is called before writing the memory data register to the
	// address pointed to by the address register.
	StoreResult()
}
