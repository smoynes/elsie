package cpu

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
	case OpcodeReserved:
		return "RESERVED"
	}
	return "UKNWN"
}

type Operation interface {
	opcode() Opcode
}

var _ Operation = &br{}

const OpcodeBR = Opcode(0b0000)

// BR: Conditional branch
// [15] 0000 [11] (NZP) [8] (PCOFFSET9) [0]
type br struct {
	nzp    Condition
	offset Word
}

func (br *br) opcode() Opcode {
	return OpcodeBR
}

func (br *br) Decode(ins Instruction) {
	br.nzp = Condition(ins & 0x0e00 >> 9)
	br.offset = Word(ins & 0x01ff)
	br.offset.Sext(9)
}

func (br *br) Execute(cpu *LC3) {
	if br.nzp&cpu.Cond != 0x0 {
		cpu.PC = ProgramCounter(int16(cpu.PC) + int16(br.offset))
	}
}

// NOT: Bitwise complement operation
// [15] 1001 [11] (RSRC) [7] (RDST)  [5] 111111 [0]
const OpcodeNOT = Opcode(0b1001)

type not struct {
	src  GPR
	dest GPR
}

var _ Operation = &not{}

func (n *not) opcode() Opcode {
	return OpcodeNOT
}

func (n *not) Decode(ins Instruction) {
	n.src = GPR(ins & 0x0e00 >> 9)
	n.dest = GPR(ins & 0x01a0 >> 6)
}

func (n *not) Execute(cpu *LC3) {
	r := cpu.Reg[n.src] ^ 0xffff
	cpu.Reg[n.dest] = r
	cpu.Cond.Update(r)
}

// AND: Bitwise AND binary operator (register mode)
// [15] 0101 [11] (RDST) [7] (RSRC1) [5] 0 [4] 00 [2] (RSRC2) [0]
const OpcodeAND = Opcode(0b0101)

type and struct {
	dest GPR
	sr1  GPR
	sr2  GPR
}

var _ Operation = &and{}

func (a *and) opcode() Opcode {
	return OpcodeAND
}

func (a *and) Decode(ins Instruction) {
	a.dest = GPR(ins & 0x0e00 >> 9)
	a.sr1 = GPR(ins & 0x01d0 >> 6)
	a.sr2 = GPR(ins & 0x0007)
}

func (a *and) Execute(cpu *LC3) {
	r := cpu.Reg[a.sr1]
	r = r & cpu.Reg[a.sr2]
	cpu.Reg[a.dest] = r
	cpu.Cond.Update(r)
}

// AND: Bitwise AND binary operator (immediate mode)
// [15] 0101 [11] (DR) [7] (SR1) [5] 1 [5] (IMM5) [0]
type andImm struct {
	dr  GPR
	sr  GPR
	lit Word
}

var _ Operation = &andImm{}

func (a *andImm) opcode() Opcode {
	return OpcodeAND
}

func (a *andImm) Decode(ins Instruction) {
	*a = andImm{
		dr:  GPR(ins & 0x0e00 >> 9),
		sr:  GPR(ins & 0x01d0 >> 6),
		lit: Word(ins & 0x001f),
	}
	a.lit.Sext(5)
}

func (a *andImm) Execute(cpu *LC3) {
	r := cpu.Reg[a.sr]
	r = r & Register(a.lit)
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
}

// RES: Reserved operator
// [15] 1101 [11] 0000 0000 0000 [0]
type reserved struct{}

func (r *reserved) opcode() Opcode {
	return OpcodeReserved
}

const OpcodeADD = Opcode(0b0001)

// ADD: Arithmetic addition operator (register mode)
// [15] 0001 [11] (DR) [7] (SR1) [5] 0 [5] 00 [2] (SR2) [0]
type add struct {
	dr  GPR
	sr1 GPR
	sr2 GPR
}

func (a *add) opcode() Opcode {
	return OpcodeADD
}

var (
	_ decodable  = &add{}
	_ executable = &add{}
)

func (a *add) Decode(ins Instruction) {
	*a = add{
		dr:  GPR(ins & 0x0e00 >> 9),
		sr1: GPR(ins & 0x1d0 >> 6),
		sr2: GPR(ins & 0x007),
	}
}

func (a *add) Execute(cpu *LC3) {
	r := Register(int16(cpu.Reg[a.sr1]) + int16(cpu.Reg[a.sr2]))
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
}

// ADD: Arithmetic addition operator (immediate mode)
// [15] 0001 [11] (DR) [7] (SR1) [5] 1 [5] IMM5 [0]
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
		dr:  GPR(ins & 0x0e00 >> 9),
		sr:  GPR(ins & 0x1d0 >> 6),
		lit: Word(ins & 0x01f),
	}
	a.lit.Sext(5)
}

func (a *addImm) Execute(cpu *LC3) {
	r := Register(int16(cpu.Reg[a.sr]) + int16(a.lit))
	cpu.Reg[a.dr] = r
	cpu.Cond.Update(r)
}

const OpcodeLD = Opcode(0b0010)

// LD: Load word from memory
// [15] 0010 [12] (DR) [9] PCOFFSET9 [0]
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

func (ld) opcode() Opcode {
	return OpcodeLD
}

func (op *ld) Decode(ins Instruction) {
	*op = ld{
		dr:     GPR(ins & 0x0e00 >> 9),
		offset: Word(ins & 0x01ff),
	}
	op.offset.Sext(9)
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

const (
	OpcodeJMP      = Opcode(0b1100)
	OpcodeJSR      = Opcode(0b0100)
	OpcodeLDI      = Opcode(0b1010)
	OpcodeLEA      = Opcode(0b1110)
	OpcodeRET      = Opcode(0b1100)
	OpcodeRTI      = Opcode(0b1000)
	OpcodeST       = Opcode(0b0011)
	OpcodeSTI      = Opcode(0b1011)
	OpcodeSTR      = Opcode(0b0111)
	OpcodeTRAP     = Opcode(0b1111)
	OpcodeReserved = Opcode(0b1101)
)
