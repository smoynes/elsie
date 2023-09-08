package cpu

// exec.go defines the CPU instruction cycle.

// Cycle runs a single instruction cycle to completion.
//
// Each cycle has six steps:
//
//   - fetch instruction: using the program counter as a pointer, fetch an
//     instruction from memory and increment the program counter.
//
//   - decode operation: decode an operation and its operands from the
//     instruction.
//
//   - evaluate address: compute the memory address to be accessed.
//
//   - fetch operands: load an operand from memory using the computed address.
//
//   - execute operation: do the thing.
//
//   - store result: store operation result in memory.
//
// Operations implement a set of methods according to their semantics.
func (cpu *LC3) Cycle() error {
	cpu.Fetch()
	op := cpu.Decode()
	cpu.EvalAddress(op)
	cpu.FetchOperands(op)
	cpu.Execute(op)
	cpu.StoreResult(op)

	return nil
}

// Fetch loads the value addressed by PC into IR and increments PC.
func (cpu *LC3) Fetch() {
	cpu.Mem.MAR = Register(cpu.PC)
	cpu.Mem.Fetch()
	cpu.IR = Instruction(cpu.Mem.MDR)
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
	case OpcodeLDR:
		op = &ldr{}
	case OpcodeLEA:
		op = &lea{}
	case OpcodeST:
		op = &st{}
	case OpcodeSTI:
		op = &sti{}
	case OpcodeSTR:
		op = &str{}
	case OpcodeJMP, OpcodeRET:
		op = &jmp{}
	case OpcodeJSR, OpcodeJSRR:
		if (cpu.IR & 0x0800) == 0 {
			op = &jsrr{}
		} else {
			op = &jsr{}
		}
	case OpcodeTRAP:
		op = &trap{}
	case OpcodeRTI:
		op = &rti{}
	case OpcodeRESV:
		op = &resv{}
	}

	if op, ok := op.(decodable); ok {
		op.Decode(cpu.IR)
	}

	return op
}

// EvalAddress computes a relative memory address if the operation is
// addressable.
func (cpu *LC3) EvalAddress(op operation) {
	if op, ok := op.(addressable); ok {
		op.EvalAddress(cpu)
	}
}

// FetchOperands reads from memory into a CPU register if the operation is fetchable.
func (cpu *LC3) FetchOperands(op operation) {
	if op, ok := op.(fetchable); ok {
		cpu.Mem.Fetch()
		op.FetchOperands(cpu)
	}
}

// Execute does the operation.
func (cpu *LC3) Execute(op operation) {
	if op, ok := op.(executable); ok {
		op.Execute(cpu)
	}
}

// StoreResult writes registers to memory if the operation is storable.
func (cpu *LC3) StoreResult(op operation) {
	if op, ok := op.(storable); ok {
		op.StoreResult(cpu)
		cpu.Mem.Store()
	}
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
	FetchOperands(cpu *LC3)
}

// executable operations update CPU state. Some instructions do not, surprisingly.
type executable interface {
	Execute(cpu *LC3)
}

// storable operations store the memory data register.
type storable interface {
	addressable
	StoreResult(cpu *LC3)
}
