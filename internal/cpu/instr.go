package cpu

// Execute runs a single instruction cycle.
func (cpu *LC3) Execute() error {
	cpu.Fetch()

	operation := cpu.Decode()

	if operation, ok := operation.(addressable); ok {
		operation.EvalAddress(cpu)
	}
	if operation, ok := operation.(interface{ FetchOperands(*LC3) }); ok {
		operation.FetchOperands(cpu)
	}
	if operation, ok := operation.(executable); ok {
		operation.Execute(cpu)
	}
	if operation, ok := operation.(interface{ StoreResult(*LC3) }); ok {
		operation.StoreResult(cpu)
	}

	return nil
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() {
	addr := Word(cpu.PC)
	cpu.IR = Instruction(cpu.Mem.Load(addr))
	cpu.PC++
}

// Decodes returns an operation from the instruction register.
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
	case OpcodeADD:
		if (cpu.IR & 0x0020) != 0 {
			op = &addImm{}
		} else {
			op = &add{}
		}
	case OpcodeLD:
		op = &ld{}
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

// executable operations use operands for computation may update CPU state. Nearly
// all operations are executable.
type executable interface {
	Operation
	Execute(cpu *LC3)
}

// fetchable operations use operands to load words from memory into registers.
type fetchable interface {
	Operation
	EvalAddress(cpu *LC3)
	FetchOperands(cpu *LC3)
}

type addressable interface {
	Operation
	EvalAddress(cpu *LC3)
}
