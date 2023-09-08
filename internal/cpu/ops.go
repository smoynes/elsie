package cpu

// An Opcode identifies the operation to be executed by the CPU. The ISA has 15
// distinct opcodes and one reserved value that is undefined and causes an
// exception.
type Opcode uint8

// BR: Conditional branch
//
// | 0000 | NZP | PCOFFSET9 |
// |------+-----+-----------|
// |15  12|11  9|8         0|
type br struct {
	cond   Condition
	offset Word
}

const OpcodeBR = Opcode(0b0000) // BR

var (
	_ decodable = &br{}
)

func (br *br) opcode() Opcode {
	return OpcodeBR
}

func (br *br) Decode(ins Instruction) {
	br.cond = ins.Cond()
	br.offset = ins.Offset(PCOFFSET9)
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
// | 0010 | DR  | PCOFFSET9 |
// |------+-----+---+-------|
// |15  12|11  9|8         0|
type ld struct {
	dr     GPR
	offset Word
	addr   Word
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
		offset: ins.Offset(PCOFFSET9),
	}
}

func (op *ld) EvalAddress(cpu *LC3) {
	op.addr = Word(int16(cpu.PC) + int16(op.offset))
}

func (op *ld) FetchOperands(cpu *LC3) {
	r := Register(cpu.Mem.Load(op.addr))
	cpu.Reg[op.dr] = r
}

func (op *ld) Execute(cpu *LC3) {
	cpu.PSR.Set(cpu.Reg[op.dr])
}

// LDI: Load indirect
//
// | 1010 | DR | PCOFFSET9 |
// |------+----------------|
// |15  12|11 9|8         0|
type ldi struct {
	dr     GPR
	offset Word
	addr   Word
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
		offset: ins.Offset(PCOFFSET9),
	}
}

func (op *ldi) EvalAddress(cpu *LC3) {
	op.addr = Word(int16(cpu.PC) + int16(op.offset))
}

func (op *ldi) FetchOperands(cpu *LC3) {
	op.addr = cpu.Mem.Load(op.addr)
}

func (op *ldi) Execute(cpu *LC3) {
	a := cpu.Mem.Load(op.addr)
	r := Register(a)
	cpu.Reg[op.dr] = r
	cpu.PSR.Set(r)
}

// LEA: Load effective address
//
// | 1110 | DR | PCOFFSET9 |
// |------+----------------|
// |15  12|11 9|8         0|
type lea struct {
	dr     GPR
	offset Word
	addr   Word
}

const OpcodeLEA = Opcode(0b1110) // LEA

var (
	_ decodable   = &lea{}
	_ addressable = &lea{}
)

func (op *lea) opcode() Opcode { return OpcodeLEA }

func (op *lea) Decode(ins Instruction) {
	*op = lea{
		dr:     ins.DR(),
		offset: ins.Offset(PCOFFSET9),
	}
}

func (op *lea) EvalAddress(cpu *LC3) {
	op.addr = Word(int16(cpu.PC) + int16(op.offset))
}

func (op *lea) Execute(cpu *LC3) {
	r := Register(op.addr)
	cpu.Reg[op.dr] = r
}

// ST: Store word in memory.
//
// | 0011 | SR  | PCOFFSET9 |
// |------+-----+---+-------|
// |15  12|11  9|8         0|
type st struct {
	sr     GPR
	offset Word
	addr   Word
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
		offset: ins.Offset(PCOFFSET9),
	}
}

func (op *st) EvalAddress(cpu *LC3) {
	op.addr = Word(int16(cpu.PC) + int16(op.offset))
}

func (op *st) Execute(cpu *LC3) {
	// TODO: check PSR and raise ACV
}

func (op *st) StoreResult(cpu *LC3) {
	cpu.Mem.Store(op.addr, Word(cpu.Reg[op.sr]))
}

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
// | 0100 |  1 | PCOFFSET11 |
// |------+----+------------|
// |15  12| 11 |10         0|
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

func (op *trap) EvalAddress(*LC3) {
	// NOP: the vector is already the address.
}

func (op *trap) FetchOperands(cpu *LC3) {
	op.isr = cpu.Mem.Load(op.vec)
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
	_ operation = &trap{}
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
type reserved struct{}

const OpcodeReserved = Opcode(0b1101) // RESV

var _ operation = &reserved{}

func (r *reserved) opcode() Opcode {
	return OpcodeReserved
}

func (reserved) Execute(cpu *LC3) {
	// TODO: raise exception
}

const (
	OpcodeSTI = Opcode(0b1011) // STI
	OpcodeSTR = Opcode(0b0111) // STR
)
