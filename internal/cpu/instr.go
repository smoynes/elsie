package cpu

import (
	"fmt"
)

// Instruction is a 16-bit value that encodes a single CPU operation. The LS-3
// ISA has 15 distinct instructions (and one reserved value that is undefined).
// The top 4 bits of an instruction define the opcode; the remaining bits are
// used for operands and flags.
type Instruction Register

func (i Instruction) String() string {
	return fmt.Sprintf("%s (OP: %s)", Word(i), i.Opcode())
}

func (i Instruction) Opcode() Opcode {
	return Opcode(i >> 12)
}

func (i Instruction) Cond() Condition {
	return Condition(i & 0x0e00 >> 9)
}

func (i Instruction) DR() GPR {
	return GPR(i & 0x0e00 >> 9)
}

func (i Instruction) SR1() GPR {
	return GPR(i & 0x01a0 >> 6)
}

func (i Instruction) SR2() GPR {
	return GPR(i & 0x0003)
}

type Offset uint8

const (
	PCOFFSET11 Offset = iota
	PCOFFSET9
	PCOFFSET5
)

func (i Instruction) Offset(n Offset) Word {
	var w Word
	switch n {
	case PCOFFSET11:
		w = Word(i & 0x03ff)
		w.Sext(11)
	case PCOFFSET9:
		w = Word(i & 0x01ff)
		w.Sext(9)
	case PCOFFSET5:
		w = Word(i & 0x001f)
		w.Sext(5)
	default:
		panic("unexpected offset")
	}
	return w
}

func (i Instruction) Imm() Word {
	w := Word(i & 0x001f)
	w.Sext(5)
	return w
}

// Execute runs a single instruction cycle.
func (cpu *LC3) Execute() error {
	cpu.Fetch()
	op := cpu.Decode()
	cpu.EvalAddress(op)
	cpu.FetchOperands(op)

	if op, ok := op.(executable); ok {
		op.Execute(cpu)
	}
	cpu.StoreResult(op)
	return nil
}

// executable operations use operands for computation may update CPU state.
// Nearly all operations are executable.
type executable interface {
	Operation
	Execute(cpu *LC3)
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() {
	addr := Word(cpu.PC)
	cpu.IR = Instruction(cpu.Mem.Load(addr))
	cpu.PC++
}

// Decode returns an operation from the instruction register.
func (cpu *LC3) Decode() Operation {
	var op Operation

	switch cpu.IR.Opcode() {
	case OpcodeBR:
		op = &br{}
	case OpcodeAND:
		if (cpu.IR & 0x0020) != 0 {
			op = &andImm{}
		} else {
			op = &and{}
		}
	case OpcodeADD:
		if (cpu.IR & 0x0020) != 0 {
			op = &addImm{}
		} else {
			op = &add{}
		}
	case OpcodeNOT:
		op = &not{}
	case OpcodeLD:
		op = &ld{}
	case OpcodeLDI:
		op = &ldi{}
	case OpcodeJMP:
		op = &jmp{}
	case OpcodeJSR:
		if (cpu.IR & 0x0800) == 0 {
			op = &jsrr{}
		} else {
			op = &jsr{}
		}
	case OpcodeReserved:
		op = &reserved{}
	default:
		panic("decode error")
	}

	if op, ok := op.(decodable); ok {
		op.Decode(cpu.IR)
	}

	return op
}

// decodable operations have operands that are decoded from the instruction
// register. Nearly all operations are decodable.
type decodable interface {
	Operation
	Decode(ir Instruction)
}

// EvalAddress computes a memory address if the operation is addressable.
func (cpu *LC3) EvalAddress(op Operation) {
	if op, ok := op.(addressable); ok {
		op.EvalAddress(cpu)
	}
}

type addressable interface {
	Operation
	EvalAddress(cpu *LC3)
}

// FetchOperands loads registers from memory if the operation is fetchable.
func (cpu *LC3) FetchOperands(op Operation) {
	if op, ok := op.(fetchable); ok {
		op.FetchOperands(cpu)
	}
}

type fetchable interface {
	Operation
	EvalAddress(cpu *LC3)
	FetchOperands(cpu *LC3)
}

// StoreResult writes registers to memory if the operation is storable.
func (cpu *LC3) StoreResult(op Operation) {
	if op, ok := op.(storable); ok {
		op.StoreResult(cpu)
	}
}

type storable interface {
	Operation
	EvalAddress(cpu *LC3)
	StoreResult(cpu *LC3)
}
