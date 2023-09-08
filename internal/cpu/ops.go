package cpu

// ops.go defines the CPU operations and their semantics.

// An Opcode identifies the operation to be executed by the CPU. The ISA has 15
// distinct opcodes and one reserved value that is undefined and causes an
// exception.
type Opcode uint8

// BR: Conditional branch
//
// | 0000 | NZP | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type br struct {
	cond   Condition
	offset Word
}

const OpcodeBR = Opcode(0b0000) // BR

var (
	_ decodable  = &br{}
	_ executable = &br{}
)

func (br *br) opcode() Opcode {
	return OpcodeBR
}

func (br *br) Decode(ins Instruction) {
	br.cond = ins.Cond()
	br.offset = ins.Offset(OFFSET9)
}

func (br *br) Execute(cpu *LC3) {
	if cpu.PSR.Any(br.cond) {
		cpu.PC = ProgramCounter(int16(cpu.PC) + int16(br.offset))
	}
}

// NOT: Bitwise complement operation
//
// | 1001 | DR | SR | 1 | 1 1111 |
// |------+----+----+---+--------|
// |15  12|11 9|8  6| 5 |4      0|
type not struct {
	src  GPR
	dest GPR
}

const OpcodeNOT = Opcode(0b1001) // NOT

var (
	_ decodable = &not{}
)

func (n *not) opcode() Opcode {
	return OpcodeNOT
}

func (n *not) Decode(ins Instruction) {
	*n = not{
		src:  ins.SR1(),
		dest: ins.DR(),
	}
}

func (n *not) Execute(cpu *LC3) {
	cpu.Temp = cpu.Reg[n.src] ^ 0xffff
	cpu.Reg[n.dest] = cpu.Temp
	cpu.PSR.Set(cpu.Temp)
}

// AND: Bitwise AND binary operator (register mode)
//
// | 0101 | DR | SR1 | 0 | 00 | SR2 |
// |------+----+-----+---+----+-----|
// |15  12|11 9|8   6| 5 |4  3|2   0|

type and struct {
	dest GPR
	sr1  GPR
	sr2  GPR
}

const OpcodeAND = Opcode(0b0101) // AND

var (
	_ decodable = &and{}
)

func (a *and) opcode() Opcode {
	return OpcodeAND
}

func (a *and) Decode(ins Instruction) {
	a.dest = ins.DR()
	a.sr1 = ins.SR1()
	a.sr2 = ins.SR2()
}

func (a *and) Execute(cpu *LC3) {
	cpu.Temp = cpu.Reg[a.sr1]
	cpu.Temp &= cpu.Reg[a.sr2]
	cpu.Reg[a.dest] = cpu.Temp
	cpu.PSR.Set(cpu.Temp)
}

// AND: Bitwise AND binary operator (immediate mode)
//
// | 0101 | DR  | SR | 1 | IMM5 |
// |------+-----+----+---+------|
// |15  12|11  9|8  6| 5 |4    0|
type andImm struct {
	dr  GPR
	sr  GPR
	lit Word
}

func (a *andImm) opcode() Opcode {
	return OpcodeAND
}

var (
	_ decodable = &andImm{}
)

func (a *andImm) Decode(ins Instruction) {
	*a = andImm{
		dr:  ins.DR(),
		sr:  ins.SR1(),
		lit: ins.Literal(IMM5),
	}
}

func (a *andImm) Execute(cpu *LC3) {
	cpu.Temp = cpu.Reg[a.sr] & Register(a.lit)
	cpu.Reg[a.dr] = cpu.Temp
	cpu.PSR.Set(cpu.Temp)
}

// ADD: Arithmetic addition operator (register mode)
//
// | 0001 | DR | SR1 | 000 | SR2 |
// |------+----+-----+-----+-----|
// |15  12|11 9|8   6| 5  3|2   0|
type add struct {
	dr  GPR
	sr1 GPR
	sr2 GPR
}

const OpcodeADD = Opcode(0b0001) // ADD

func (a *add) opcode() Opcode {
	return OpcodeADD
}

var (
	_ decodable = &add{}
)

func (a *add) Decode(ins Instruction) {
	*a = add{
		dr:  ins.DR(),
		sr1: ins.SR1(),
		sr2: ins.SR2(),
	}
}

func (a *add) Execute(cpu *LC3) {
	cpu.Temp = Register(int16(cpu.Reg[a.sr1]) + int16(cpu.Reg[a.sr2]))
	cpu.Reg[a.dr] = cpu.Temp
	cpu.PSR.Set(cpu.Temp)
}

// ADD: Arithmetic addition operator (immediate mode)
//
// | 0001 | DR  | SR | 1 | 11111 |
// |------+-----+----+---+-------|
// |15  12|11  9|8  6| 5 |4     0|
type addImm struct {
	dr  GPR
	sr  GPR
	lit Word
}

func (a *addImm) opcode() Opcode {
	return OpcodeADD
}

var (
	_ decodable = &addImm{}
)

func (a *addImm) Decode(ins Instruction) {
	*a = addImm{
		dr:  ins.DR(),
		sr:  ins.SR1(),
		lit: ins.Literal(IMM5),
	}
}

func (a *addImm) Execute(cpu *LC3) {
	cpu.Temp = Register(int16(cpu.Reg[a.sr]) + int16(a.lit))
	cpu.Reg[a.dr] = cpu.Temp
	cpu.PSR.Set(cpu.Temp)
}

// LD: Load word from memory.
//
// | 0010 | DR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type ld struct {
	dr     GPR
	offset Word
}

var (
	_ decodable   = &ld{}
	_ addressable = &ld{}
	_ fetchable   = &ld{}
)

const OpcodeLD = Opcode(0b0010) // LD

func (ld) opcode() Opcode {
	return OpcodeLD
}

func (op *ld) Decode(ins Instruction) {
	*op = ld{
		dr:     ins.DR(),
		offset: ins.Offset(OFFSET9),
	}
}

func (op *ld) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.PC) + int16(op.offset))
}

func (op *ld) FetchOperands(cpu *LC3) {
	cpu.Reg[op.dr] = cpu.Mem.MDR
}

func (op *ld) Execute(cpu *LC3) {
	cpu.PSR.Set(cpu.Reg[op.dr])
}

// LDI: Load indirect
//
// | 1010 | DR | OFFSET9 |
// |------+--------------|
// |15  12|11 9|8       0|
type ldi struct {
	dr     GPR
	offset Word
}

const OpcodeLDI = Opcode(0b1010) // LDI

var (
	_ decodable   = &ldi{}
	_ addressable = &ldi{}
	_ fetchable   = &ldi{}
)

func (op *ldi) opcode() Opcode { return OpcodeLDI }

func (op *ldi) Decode(ins Instruction) {
	*op = ldi{
		dr:     ins.DR(),
		offset: ins.Offset(OFFSET9),
	}
}

func (op *ldi) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.PC) + int16(op.offset))
}

func (op *ldi) FetchOperands(cpu *LC3) {
	cpu.Mem.MAR = cpu.Mem.MDR
	cpu.Mem.Fetch()
	cpu.Reg[op.dr] = cpu.Mem.MDR
}

func (op *ldi) Execute(cpu *LC3) {
	cpu.PSR.Set(cpu.Mem.MDR)
}

// LDR: Load Relative
//
// | 0110 | DR | BASE | OFFSET6 |
// |------+----+------+---------|
// |15  12|11 9|8    6|5       0|
type ldr struct {
	dr     GPR
	base   GPR
	offset Word
}

const OpcodeLDR = Opcode(0b0110) // LDR

var (
	_ decodable   = &ldr{}
	_ addressable = &ldr{}
	_ fetchable   = &ldr{}
)

func (op *ldr) opcode() Opcode { return OpcodeLDR }

func (op *ldr) Decode(ins Instruction) {
	*op = ldr{
		dr:     ins.DR(),
		base:   ins.SR1(),
		offset: ins.Offset(OFFSET6),
	}
}

func (op *ldr) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.Reg[op.base]) + int16(op.offset))
}

func (op *ldr) FetchOperands(cpu *LC3) {
	cpu.Reg[op.dr] = cpu.Mem.MDR
}

func (op *ldr) Execute(cpu *LC3) {
	cpu.PSR.Set(cpu.Reg[op.dr])
}

// LEA: Load effective address
//
// | 1110 | DR | OFFSET9 |
// |------+--------------|
// |15  12|11 9|8       0|
type lea struct {
	dr     GPR
	offset Word
}

const OpcodeLEA = Opcode(0b1110) // LEA

var (
	_ decodable = &lea{}
	_ fetchable = &lea{}
)

func (op *lea) opcode() Opcode { return OpcodeLEA }

func (op *lea) Decode(ins Instruction) {
	*op = lea{
		dr:     ins.DR(),
		offset: ins.Offset(OFFSET9),
	}
}

func (op *lea) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.PC) + int16(op.offset))
	println(cpu.Mem.MAR.String())
}

func (op *lea) FetchOperands(cpu *LC3) {
	cpu.Reg[op.dr] = cpu.Mem.MDR
	println(cpu.Mem.MDR.String())
	println(op.dr.String())

}

// ST: Store word in memory.
//
// | 0011 | SR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type st struct {
	sr     GPR
	offset Word
}

var (
	_ decodable   = &st{}
	_ addressable = &st{}
	_ storable    = &st{}
)

const OpcodeST = Opcode(0b0011) // ST

func (st) opcode() Opcode {
	return OpcodeST
}

func (op *st) Decode(ins Instruction) {
	*op = st{
		sr:     ins.SR(),
		offset: ins.Offset(OFFSET9),
	}
}

func (op *st) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.PC) + int16(op.offset))
}

func (op *st) Execute(cpu *LC3) {
	cpu.Mem.MDR = cpu.Reg[op.sr]
}

func (op *st) StoreResult(cpu *LC3) {} // ?

// STI: Store Indirect.
//
// | 1011 | SR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type sti struct {
	sr     GPR
	offset Word
}

var (
	_ decodable   = &sti{}
	_ addressable = &sti{}
	_ fetchable   = &sti{}
	_ storable    = &sti{}
)

const OpcodeSTI = Opcode(0b1011) // STI

func (sti) opcode() Opcode {
	return OpcodeSTI
}

func (op *sti) Decode(ins Instruction) {
	*op = sti{
		sr:     ins.SR(),
		offset: ins.Offset(OFFSET9),
	}
}

func (op *sti) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.PC) + int16(op.offset))
}

func (op *sti) FetchOperands(cpu *LC3) {
	cpu.Mem.MAR = cpu.Mem.MDR
}

func (op *sti) Execute(cpu *LC3) {
	cpu.Mem.MDR = cpu.Reg[op.sr]
}

func (op *sti) StoreResult(cpu *LC3) {}

// STR: Store Relative.
//
// | 0111 | SR  | GPR | OFFSET6 |
// |------+-----+-----+---------|
// |15  12|11  9|8   6|5       0|
type str struct {
	sr     GPR
	base   GPR
	offset Word
}

var (
	_ decodable   = &str{}
	_ addressable = &str{}
	_ storable    = &str{}
)

const OpcodeSTR = Opcode(0b0111) // STR

func (str) opcode() Opcode {
	return OpcodeSTR
}

func (op *str) Decode(ins Instruction) {
	*op = str{
		sr:     ins.SR(),
		base:   ins.SR1(),
		offset: ins.Offset(OFFSET6),
	}
}

func (op *str) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(int16(cpu.Reg[op.base]) + int16(op.offset))
}

func (op *str) Execute(cpu *LC3) {
	cpu.Mem.MDR = cpu.Reg[op.sr]
}

func (op *str) StoreResult(cpu *LC3) {}

// JMP: Unconditional branch
//
// | 1100 | 000 | SR | 00 00000 |
// |------+-----+----+----------|
// |15  12|11  9|8  6|5        0|
//
// RET: Return from subroutine
// | 1100 | 111 | SR | 00 00000 |
// |------+-----+----+----------|
// |15  12|11  9|8  6|5        0|
type jmp struct {
	sr GPR
}

const (
	OpcodeJMP = Opcode(0b1100) // JMP
	OpcodeRET = Opcode(0xff)   // RET
)

var (
	_ decodable = &jmp{}
)

func (j jmp) opcode() Opcode {
	if j.sr == R7 {
		return OpcodeRET
	} else {
		return OpcodeJMP
	}
}

func (op *jmp) Decode(ins Instruction) {
	*op = jmp{
		sr: GPR(ins & 0x01e0 >> 6),
	}
}

func (op *jmp) Execute(cpu *LC3) {
	pc := ProgramCounter(cpu.Reg[op.sr])
	cpu.PC = pc
}

// JSR: Jump to subroutine (relative mode)
//
// | 0100 |  1 | OFFSET11 |
// |------+----+----------|
// |15  12| 11 |10       0|
type jsr struct {
	offset Word
}

const OpcodeJSR = Opcode(0b0100) // JSR

var (
	_ decodable = &jsr{}
)

func (op *jsr) opcode() Opcode { return OpcodeJSR }

func (op *jsr) Decode(ins Instruction) {
	*op = jsr{
		offset: Word(ins & 0x07ff),
	}
	op.offset.Sext(11)
}

func (op *jsr) Execute(cpu *LC3) {
	ret := Word(cpu.PC)
	pc := ProgramCounter(int16(cpu.PC) + int16(op.offset))
	cpu.PC = pc
	cpu.Reg[RET] = Register(ret)
}

// JSRR: Jump to subroutine (register mode)
//
// | 0100 |  0 | SR | 00 0000 |
// |------+----+----+---------|
// |15  12| 11 |8  6|5       0|
type jsrr struct {
	sr GPR
}

const OpcodeJSRR = Opcode(0xfe) // JSRR

var (
	_ decodable = &jsrr{}
)

func (op *jsrr) opcode() Opcode { return OpcodeJSRR }

func (op *jsrr) Decode(ins Instruction) {
	*op = jsrr{
		sr: ins.SR1(),
	}
}

func (op *jsrr) Execute(cpu *LC3) {
	ret := Word(cpu.PC)
	pc := ProgramCounter(cpu.Reg[op.sr])
	cpu.PC = pc
	cpu.Reg[RET] = Register(ret)
}

// TRAP: System call or software interrupt.
//
// | 1111 | 0000 | VECTOR8 |
// |------+------+---------|
// |15  12|11   8|7       0|
type trap struct {
	vec Word
	isr Word
}

const OpcodeTRAP = Opcode(0b1111) // TRAP

func (op *trap) opcode() Opcode {
	return OpcodeTRAP
}

var (
	_ decodable = &trap{}
	_ fetchable = &trap{}
)

func (op *trap) Decode(ins Instruction) {
	*op = trap{
		vec: ins.Vector(VECTOR8),
		isr: 0x0000,
	}
}

func (op *trap) EvalAddress(cpu *LC3) {
	cpu.Mem.MAR = Register(op.vec)
}

func (op *trap) FetchOperands(cpu *LC3) {
	op.isr = Word(cpu.Mem.MDR)
}

func (op *trap) Execute(cpu *LC3) {
	cpu.Temp = Register(cpu.PSR)

	// Switch from the user to the system stack and elevate to system
	// privilege level.
	if cpu.PSR.Privilege() == PrivilegeUser {
		cpu.USP = cpu.Reg[SP]
		cpu.Reg[SP] = cpu.SSP
		cpu.PSR ^= StatusPrivilege
	}

	// Push the old status register and program counter onto the stack.
	cpu.PushStack(Word(cpu.Temp))
	cpu.PushStack(Word(cpu.PC))

	// Finally, jump to the ISR using the interrupt vector.
	cpu.PC = ProgramCounter(op.isr)
}

// RTI: Return from trap or interrupt
//
// | 1000 | 0000 0000 0000 |
// |------+----------------|
// |15  12|11             0|
type rti struct{}

const OpcodeRTI = Opcode(0b1000) // RTI

func (op *rti) opcode() Opcode {
	return OpcodeRTI
}

var (
	_ executable = &trap{}
)

func (op *rti) Execute(cpu *LC3) {
	if cpu.PSR.Privilege() == PrivilegeUser {
		// TODO: raise privilege exception
		return
	}

	// Restore program counter and status register.
	cpu.PC = ProgramCounter(cpu.PopStack())
	cpu.PSR = ProcessorStatus(cpu.PopStack())

	if cpu.PSR.Privilege() == PrivilegeUser {
		// When changing back to user privileges, swap the system and
		// user stack pointers.
		cpu.SSP = cpu.Reg[SP]
		cpu.Reg[SP] = cpu.USP
	}
}

// RESV: Reserved operator
//
// | 1101 | 0000 0000 0000 |
// |------+----------------|
// |15  12|11             0|
type resv struct{}

const OpcodeRESV = Opcode(0b1101) // RESV

var _ executable = &resv{}

func (r *resv) opcode() Opcode {
	return OpcodeRESV
}

func (*resv) Execute(cpu *LC3) {
	// TODO: raise exception
}
