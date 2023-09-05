package cpu

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
