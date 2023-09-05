package cpu

type Opcode uint8

type Operation interface {
	opcode() Opcode
}

var _ Operation = &br{}

// BR: Conditional branch
// [15] 0000 [11] (NZP) [8] (PCOFFSET9) [0]

const OpcodeBR = Opcode(0b0000)

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
	if br.nzp&cpu.Proc.Cond != 0x0 {
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
	r := cpu.Proc.Reg[n.src] ^ 0xffff
	cpu.Proc.Reg[n.dest] = r
	cpu.Proc.Cond.Update(r)
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
	r := cpu.Proc.Reg[a.sr1]
	r = r & cpu.Proc.Reg[a.sr2]
	cpu.Proc.Reg[a.dest] = r
	cpu.Proc.Cond.Update(r)
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
	r := cpu.Proc.Reg[a.sr]
	r = r & Register(a.lit)
	cpu.Proc.Reg[a.dr] = r
	cpu.Proc.Cond.Update(r)
}

// RES: Reserved operator
// [15] 1101 [11] 0000 0000 0000 [0]
type reserved struct{}

func (r *reserved) opcode() Opcode {
	return OpcodeReserved
}

func (o Opcode) String() string {
	switch o {
	case OpcodeBR:
		return "BR"
	case OpcodeNOT:
		return "NOT"
	case OpcodeAND:
		return "AND"
	case OpcodeReserved:
		return "RESERVED"
	}
	return "UKNWN"
}

const (
	OpcodeJMP      = Opcode(0b1100)
	OpcodeJSR      = Opcode(0b0100)
	OpcodeADD      = Opcode(0b0001)
	OpcodeLD       = Opcode(0b0010)
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
