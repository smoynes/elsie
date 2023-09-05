package cpu

import (
	"fmt"
)

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() {
	addr := Word(cpu.PC)
	cpu.IR = Instruction(cpu.Mem.Load(addr))
	cpu.PC++
}

func (cpu *LC3) Decode() Operation {
	var op Operation

	switch cpu.IR.Opcode() {
	case OpcodeReserved:
		op = &reserved{}
	case OpcodeAND:
		if (cpu.IR & 0x0020) != 0 {
			op = &andImm{}
		} else {
			op = &and{}
		}
	case OpcodeNOT:
		op = &not{}
	case OpcodeBR:
		op = &br{}
	default:
		panic("decode error")
	}

	if op, ok := op.(interface{ Decode(Instruction) }); ok {
		op.Decode(cpu.IR)
	}

	return op
}

// Execute runs a single instruction cycle.
func (cpu *LC3) Execute() error {
	cpu.Fetch()

	operation := cpu.Decode()

	if operation, ok := operation.(interface{ EvalAddress(*LC3) }); ok {
		operation.EvalAddress(cpu)
	}
	if operation, ok := operation.(interface{ FetchOperands(*LC3) }); ok {
		operation.FetchOperands(cpu)
	}
	if operation, ok := operation.(interface{ Execute(*LC3) }); ok {
		operation.Execute(cpu)
	}
	if operation, ok := operation.(interface{ StoreResult(*LC3) }); ok {
		operation.StoreResult(cpu)
	}

	return nil
}

// An Instruction is a 16-bit value that encodes a single CPU Instruction. The
// LS-3 ISA has 15 distinct instructions (and one reserved value that is
// undefined). The top 4 bits of an instruction define the opcode; the remaining
// bits are used for operands.
type Instruction Register

func (i Instruction) String() string {
	return fmt.Sprintf("%0#4x (OP: %s)", Word(i), i.Opcode())
}

func (i Instruction) Opcode() Opcode {
	return Opcode(i >> 12)
}
