package cpu

// An Opcode identifies the operation to be executed by the CPU. The ISA has 15
// distinct opcodes (and one reserved value that is undefined).
type Opcode uint8

func (o Opcode) String() string {
	switch o {
	case OpcodeBR:
		return "BR"
	case OpcodeNOT:
		return "NOT"
	case OpcodeAND:
		return "AND"
	case OpcodeADD:
		return "ADD"
	case OpcodeLD:
		return "LD"
	case OpcodeLDI:
		return "LDI"
	case OpcodeLEA:
		return "LEA"
	case OpcodeST:
		return "ST"
	case OpcodeJMP:
		return "JMP"
	case OpcodeReserved:
		return "RESERVED"
	}
	return "UKNWN"
}

// BR: Conditional branch
//
// | 0000 | NZP | PCOFFSET9 |
// |------+-----+-----------|
// |15  12|11  9|8         0|
type br struct {
	nzp    Condition
	offset Word
}

const OpcodeBR = Opcode(0b0000)

var (
	_ decodable  = &br{}
	_ executable = &br{}
)

func (br *br) opcode() Opcode {
	return OpcodeBR
}

func (br *br) Decode(ins Instruction) {
	br.nzp = ins.Cond()
	br.offset = ins.Offset(PCOFFSET9)
}

func (br *br) Execute(cpu *LC3) {
	if br.nzp&cpu.Cond != 0x0 {
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

const OpcodeNOT = Opcode(0b1001)

var (
	_ decodable  = &not{}
	_ executable = &not{}
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
	r := cpu.Reg[n.src] ^ 0xffff
	cpu.Reg[n.dest] = r
	cpu.Cond.Update(r)
}

// AND: Bitwise AND binary operator (register mode)
//
// | 0101 | DR | SR1 | 0 | 00 | SR2 |
// |------+----+-----+---+----+-----|
// |15  12|11 9|8   6| 5 |4  3|2   0|
const OpcodeAND = Opcode(0b0101)

type and struct {
	dest GPR
	sr1  GPR
	sr2  GPR
}

var (
	_ decodable  = &and{}
	_ executable = &and{}
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
	r := cpu.Reg[a.sr1]
	r = r & cpu.Reg[a.sr2]
	cpu.Reg[a.dest] = r
	cpu.Cond.Update(r)
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
	_ decodable  = &andImm{}
	_ executable = &andImm{}
)

func (a *andImm) Decode(ins Instruction) {
	*a = andImm{
		dr:  ins.DR(),
		sr:  ins.SR1(),
		lit: ins.Imm5(),
	}

}

func (a *andImm) Execute(cpu *LC3) {
	r := cpu.Reg[a.sr]
	r = r & Register(a.lit)
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
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

const OpcodeADD = Opcode(0b0001)

func (a *add) opcode() Opcode {
	return OpcodeADD
}

var (
	_ decodable  = &add{}
	_ executable = &add{}
)

func (a *add) Decode(ins Instruction) {
	*a = add{
		dr:  ins.DR(),
		sr1: ins.SR1(),
		sr2: ins.SR2(),
	}
}

func (a *add) Execute(cpu *LC3) {
	r := Register(int16(cpu.Reg[a.sr1]) + int16(cpu.Reg[a.sr2]))
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
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
	_ decodable  = &addImm{}
	_ executable = &addImm{}
)

func (a *addImm) Decode(ins Instruction) {
	*a = addImm{
		dr:  ins.DR(),
		sr:  ins.SR1(),
		lit: ins.Imm5(),
	}
}

func (a *addImm) Execute(cpu *LC3) {
	r := Register(int16(cpu.Reg[a.sr]) + int16(a.lit))
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
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
	_ executable  = &ld{}
)

const OpcodeLD = Opcode(0b0010)

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
	r := Register(cpu.Mem[op.addr])
	cpu.Reg[op.dr] = r
}

func (op *ld) Execute(cpu *LC3) {
	cpu.Cond.Update(cpu.Reg[op.dr])
}

// LDI: Load indirect
//
// | 1010 | DR | PCOFFSET9 |
// |------+----------------|
// |15  12|11 9|8         0|
const OpcodeLDI = Opcode(0b1010)

type ldi struct {
	dr     GPR
	offset Word
	addr   Word
}

var (
	_ decodable   = &ldi{}
	_ addressable = &ldi{}
	_ fetchable   = &ldi{}
	_ executable  = &ldi{}
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
	a := cpu.Mem[op.addr]
	op.addr = cpu.Mem[a]
}

func (op *ldi) Execute(cpu *LC3) {
	r := Register(op.addr)
	cpu.Reg[op.dr] = r
	cpu.Cond.Update(r)
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

const OpcodeLEA = Opcode(0b1110)

var (
	_ decodable   = &lea{}
	_ addressable = &lea{}
	_ executable  = &lea{}
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
	_ executable  = &st{}
	_ storable    = &st{}
)

const OpcodeST = Opcode(0b0011)

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
	cpu.Mem[op.addr] = Word(cpu.Reg[op.sr])
}

// JMP: Unconditional branch
//
// | 1100 | 000 | SR | 00 00000 |
// |------+-----+----+----------|
// |15  12|11  9|8  6|5        0|
type jmp struct {
	sr GPR
}

const OpcodeJMP = Opcode(0b1100)

var (
	_ decodable  = &jmp{}
	_ executable = &jmp{}
)

func (jmp) opcode() Opcode {
	return OpcodeJMP
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

const OpcodeJSR = Opcode(0b0100)

var (
	_ decodable  = &jsr{}
	_ executable = &jsr{}
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
	cpu.Reg[R7] = Register(ret)
}

// JSRR: Jump to subroutine (register mode)
//
// | 0100 |  0 | SR | 00 0000 |
// |------+----+----+---------|
// |15  12| 11 |8  6|5       0|
type jsrr struct {
	sr GPR
}

var (
	_ decodable  = &jsrr{}
	_ executable = &jsrr{}
)

func (op *jsrr) opcode() Opcode { return OpcodeJSR }

func (op *jsrr) Decode(ins Instruction) {
	*op = jsrr{
		sr: ins.SR1(),
	}
}

func (op *jsrr) Execute(cpu *LC3) {
	ret := Word(cpu.PC)
	pc := ProgramCounter(cpu.Reg[op.sr])
	cpu.PC = pc
	cpu.Reg[R7] = Register(ret)
}

// RES: Reserved operator
//
// | 1101 | 0000 0000 0000 |
// |------+----------------|
// |15  12|11             0|
type reserved struct{}

func (r *reserved) opcode() Opcode {
	return OpcodeReserved
}

const (
	OpcodeRET      = Opcode(0b1100)
	OpcodeRTI      = Opcode(0b1000)
	OpcodeSTI      = Opcode(0b1011)
	OpcodeSTR      = Opcode(0b0111)
	OpcodeTRAP     = Opcode(0b1111)
	OpcodeReserved = Opcode(0b1101)
)
