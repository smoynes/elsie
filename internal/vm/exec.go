package vm

// exec.go defines the CPU instruction cycle.

import (
	"context"
	"fmt"
)

// Run starts and executes the instruction cycle until the program halts.
func (vm *LC3) Run(ctx context.Context) error {
	var err error

	vm.log.Printf("Initial state\n%s\n%s\n", vm, vm.Reg.String())

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

		vm.log.Printf("Instruction complete\n%s\n%s\n", vm.String(), vm.Reg.String())
	}

	vm.log.Println("System HALTED")

	return err
}

// Step runs a single instruction cycle to completion.
//
// Each cycle has six steps:
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
// Each of these steps is optional: an instruction implements methods according
// to its semantics.
func (vm *LC3) Step() error {
	if err :=
		vm.Fetch(); err != nil {
		return fmt.Errorf("step: %w", err)
	}

	op := vm.Decode()
	vm.EvalAddress(op)
	vm.FetchOperands(op)
	vm.Execute(op)
	vm.StoreResult(op)

	switch int := op.Err().(type) {
	case interruptable:
		vm.log.Printf("step: interrupt: %s", op.Err())

		if err := int.Handle(vm); err != nil {
			return fmt.Errorf("step: interrupt: %w", err)
		}
	case nil: // success ☺️
		vm.log.Printf("step: executed: %+v", op)
		return nil
	default:
		vm.log.Panicf("step: error: %s", op.Err())
		return nil // unreachable
	}

	return nil
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (vm *LC3) Fetch() error {
	vm.Mem.MAR = Register(vm.PC)

	if err := vm.Mem.Fetch(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	vm.IR = Instruction(vm.Mem.MDR)
	vm.PC++

	vm.log.Printf("fetched IR: %s", vm.IR)

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

	return oper
}

// EvalAddress computes a relative memory address if the operation is
// addressable.
func (vm *LC3) EvalAddress(op operation) {
	if op, ok := op.(addressable); ok && op.Err() == nil {
		op.EvalAddress()
	}
}

// FetchOperands reads from memory into a CPU register if the operation is fetchable.
func (vm *LC3) FetchOperands(op operation) {
	if op, ok := op.(fetchable); ok && op.Err() == nil {
		if err := vm.Mem.Fetch(); err != nil {
			op.Fail(fmt.Errorf("operand: %w", err))
			return
		}

		op.FetchOperands()
	}
}

// Execute does the operation.
func (vm *LC3) Execute(op operation) {
	if op, ok := op.(executable); ok && op.Err() == nil {
		op.Execute()
	}
}

// StoreResult writes registers to memory if the operation is storable.
func (vm *LC3) StoreResult(op operation) {
	if op, ok := op.(storable); ok && op.Err() == nil {
		op.StoreResult() // Can't fail.

		if err := vm.Mem.Store(); err != nil {
			op.Fail(err)
		}
	}
}

// An operation represents a single CPU instruction as it is being executed by
// the machine. The instruction's semantics are defined by implementing optional
// interfaces for each execution step: [addressable], [fetchable], [executable],
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
