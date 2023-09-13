package cpu

// exec.go defines the CPU instruction cycle.

import (
	"errors"
	"fmt"
	"log"
)

// Run starts and executes the instruction cycle until the program halts.
func (cpu *LC3) Run() error {
	var err error

	log.Printf("Initial state\n%s\n%s\n", cpu, cpu.Reg.String())

	for {
		if cpu.MCR == 0x0000 {
			break
		}

		err = cpu.Cycle()
		if err != nil {
			return err
		}

		log.Printf("Instruction complete\n%s\n%s\n", cpu, cpu.Reg)
	}

	log.Println("System HALTED")

	return nil
}

// Cycle runs a single instruction cycle to completion.
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
// Each of these steps is optional: an instruction must implement methods
// according to its semantics.
func (cpu *LC3) Cycle() error {
	var (
		err error
		op  operation
	)

	err = cpu.Fetch()

	if err == nil {
		op = cpu.IR.Decode()
	}

	if err == nil {
		err = cpu.EvalAddress(op)
	}

	if err == nil {
		err = cpu.FetchOperands(op)
	}

	if err == nil {
		err = cpu.Execute(op)
	}

	if err == nil {
		err = cpu.StoreResult(op)
	}

	if err != nil {
		log.Printf("ins: %s, error: %s\n", op, err)
		log.Printf("\n%s", cpu.String())
	}

	switch intr := errors.Unwrap(err).(type) {
	case interruptable:
		log.Printf("Handling interrupt: %s", intr)
		err = intr.Handle(cpu)
	}

	if err != nil {
		// Either the error is not interruptable or it is but invoking
		// the handler failed.
		panic(err)
	}

	return err
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() error {
	cpu.Mem.MAR = Register(cpu.PC)
	err := cpu.Mem.Fetch()
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	cpu.IR = Instruction(cpu.Mem.MDR)
	cpu.PC++
	log.Printf("fetched IR: %s", cpu.IR)

	return nil
}

// EvalAddress computes a relative memory address if the operation is
// addressable.
func (cpu *LC3) EvalAddress(op operation) error {
	if op, ok := op.(addressable); ok {
		log.Printf("evaluating: %#v", op)
		op.EvalAddress(cpu)
	}

	return nil
}

// FetchOperands reads from memory into a CPU register if the operation is fetchable.
func (cpu *LC3) FetchOperands(op operation) error {
	var err error

	if op, ok := op.(fetchable); ok {
		log.Printf("fetching: %#v", op)
		err = cpu.Mem.Fetch()
		if err != nil {
			return fmt.Errorf("operand: %w", err)
		}

		err = op.FetchOperands(cpu)
		if err != nil {
			return fmt.Errorf("operand: %w", err)
		}
	}

	return nil
}

// Execute does the operation.
func (cpu *LC3) Execute(op operation) error {
	var err error
	if op, ok := op.(executable); ok {
		log.Printf("executing: %#v", op)
		err = op.Execute(cpu)
		if err != nil {
			return fmt.Errorf("execute: %w", err)
		}
	}

	return nil
}

// StoreResult writes registers to memory if the operation is storable.
func (cpu *LC3) StoreResult(op operation) error {
	var err error
	if op, ok := op.(storable); ok {
		log.Printf("storing: %#v", op)
		op.StoreResult(cpu)
		err = cpu.Mem.Store()
		if err != nil {
			return fmt.Errorf("store: %w", err)
		}
	}

	return nil
}

// operations represents a single CPU instruction as it is being executed. The
// semantics defined by implementing optional interfaces: [decodable],
// [addressable], [fetchable], [executable], [storable].
type operation interface {
	opcode() Opcode
}

// decodable operations have operands that are decoded from the instruction
// register and stored before evaluation. Nearly all operations are decodable.
type decodable interface {
	operation
	Decode(ir Instruction)
}

// addressable operations set the memory address register.
type addressable interface {
	operation
	EvalAddress(cpu *LC3)
}

// fetchable operations load operands from the memory data registers.
type fetchable interface {
	addressable
	FetchOperands(cpu *LC3) error
}

// executable operations update CPU state. Some instructions do not, surprisingly.
type executable interface {
	Execute(cpu *LC3) error
}

// storable operations store the memory data register.
type storable interface {
	addressable
	StoreResult(cpu *LC3)
}
