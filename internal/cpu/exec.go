package cpu

// Execute runs a single instruction cycle to completion.
//
// Each cycle includes six stages:
//
//   - fetch instruction: using the program counter as a pointer, fetch an
//     instruction from memory and increment the program counter.
//
//   - decode operation: decode the operation to be performed and its operands
//     from the instruction.
//
//   - evaluate address: compute the memory address for loading or storing values
//     in memory.
//
// - fetch operands: load an operand from memory using the computed address.
//
// - execute operation: do the thing.
//
// - store result: store operation result in memory.
//
// An operation may or may not implement each phase depending on its semantics.
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

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() {
	addr := Word(cpu.PC)
	word := cpu.Mem.Load(addr)
	cpu.IR = Instruction(word)
	cpu.PC++
}

// Decode returns the operation encoded in the instruction register.
func (cpu *LC3) Decode() operation {
	var op operation

	switch cpu.IR.Opcode() {
	case OpcodeBR:
		op = &br{}
	case OpcodeAND:
		if cpu.IR.Imm() {
			op = &andImm{}
		} else {
			op = &and{}
		}
	case OpcodeADD:
		if cpu.IR.Imm() {
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
	case OpcodeLEA:
		op = &lea{}
	case OpcodeST:
		op = &st{}
	case OpcodeJMP:
		op = &jmp{}
	case OpcodeJSR:
		if (cpu.IR & 0x0800) == 0 {
			op = &jsrr{}
		} else {
			op = &jsr{}
		}
	case OpcodeTRAP:
		op = &trap{}
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

// An operation represents a CPU operation as it is being executed. It contains
// its decoded operands and evaluation semantics implemented as several optional
// methods defined below.
type operation interface {
	opcode() Opcode
}

// decodable operations have operands that are decoded from the instruction
// register and stored before evaluation. Nearly all operations are decodable.
type decodable interface {
	operation
	Decode(ir Instruction)
}

// EvalAddress computes a memory address if the operation is addressable.
func (cpu *LC3) EvalAddress(op operation) {
	if op, ok := op.(addressable); ok {
		op.EvalAddress(cpu)
	}
}

// addressable operations evaluate relative addresses to be loaded or stored.
type addressable interface {
	operation
	EvalAddress(cpu *LC3)
}

// FetchOperands loads a register from memory if the operation is fetchable.
func (cpu *LC3) FetchOperands(op operation) {
	if op, ok := op.(fetchable); ok {
		op.FetchOperands(cpu)
	}
}

// fetchable operations load words of memory into registers.
type fetchable interface {
	addressable
	FetchOperands(cpu *LC3)
}

// executable operations use operands for computation to update CPU state.
// Nearly all operations are executable.
type executable interface {
	operation
	Execute(cpu *LC3)
}

// StoreResult writes registers to memory if the operation is storable.
func (cpu *LC3) StoreResult(op operation) {
	if op, ok := op.(storable); ok {
		op.StoreResult(cpu)
	}
}

// storable operations write registers to memory.
type storable interface {
	addressable
	StoreResult(cpu *LC3)
}
