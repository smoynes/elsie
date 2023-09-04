package cpu

import (
	"fmt"
)

// Execute runs a single instruction cycle.
func (cpu *LC3) Execute() error {
	// 1. FETCH INSTRUCTION: load value addressed by PC into IR and increment PC.
	addr := Word(cpu.PC)
	cpu.IR = Instruction(cpu.Mem.Load(addr))
	cpu.PC++

	// 2. DECODE
	opcode := cpu.IR.Decode()

	// 3. EVALUATE ADDRESS

	// 4. FETCH OPERANDS
	// 5. EXECUTE
	// 6. STORE RESULT
	switch opcode {
	case OpcodeNOT:
		src := GPR(cpu.IR & 0x0e00 >> 9)
		dest := GPR(cpu.IR & 0x01a0 >> 6)
		cpu.Proc.Not(src, dest)
	case OpcodeAND:
		if cpu.IR&(1<<5) != 0x0000 {
			regA := GPR(cpu.IR & 0x0e00 >> 9)
			dest := GPR(cpu.IR & 0x01a0 >> 6)
			lit := Word(cpu.IR & 0x001f)
			cpu.Proc.AndImm(regA, lit, dest)
		} else {
			regA := GPR(cpu.IR & 0x0e00 >> 9)
			dest := GPR(cpu.IR & 0x01a0 >> 6)
			regB := GPR(cpu.IR & 0x0003)
			cpu.Proc.And(regA, regB, dest)
		}
	default:
		panic("cannot decode instruction")
	}
	return nil
}

func (proc *Processor) Not(regA, regB GPR) {
	a := Word(proc.Reg[regA])
	r := a ^ 0xffff
	proc.Reg[regB] = Register(r)
	proc.Cond.Update(r)
}

func (proc *Processor) And(regA, regB, regC GPR) {
	a := Word(proc.Reg[regA])
	b := Word(proc.Reg[regB])
	r := a & b
	proc.Reg[regC] = Register(r)
	proc.Cond.Update(r)
}

func (proc *Processor) AndImm(regA GPR, lit Word, dest GPR) {
	a := Word(proc.Reg[regA])
	lit = proc.sext(lit, 5)
	r := a & lit
	proc.Reg[dest] = Register(r)
	proc.Cond.Update(r)
}

func (proc *Processor) sext(val Word, n uint8) Word {
	ans := int16(val)
	ans <<= 16 - n
	ans >>= 16 - n
	return Word(uint16(ans))
}

// An Instruction is a 16-bit value that encodes a single CPU Instruction. The
// LS-3 ISA has 15 distinct instructions (and one reserved value that is
// undefined). The top 4 bits of an instruction define the opcode; the remaining
// bits are used for operands.
type Instruction Word

func (i Instruction) String() string {
	return fmt.Sprintf("%0#4x (OP: %s)", Word(i), i.Decode())
}

func (i Instruction) Decode() Opcode {
	return Opcode(i >> 12)
}

type Opcode uint8

func (o Opcode) String() string {
	switch o {
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
	OpcodeBR  Opcode = 0b0000
	OpcodeJMP        = 0b1100
	OpcodeJSR        = 0b0100
	OpcodeADD        = 0b0001

	// AND
	// [15] 0101 [11] (RDST) [7] (RSRC1) [5] 0 [4] 00 [2] (RSRC2) [0]
	// [15] 0101 [11] (RDST) [7] (RSRC1) [5] 1 [5] (IMMED) [0]
	OpcodeAND = 0b0101

	// NOT
	// [15] 1001 [11] (RSRC) [7] (RDST)  [5] 111111 [0]
	OpcodeNOT      = 0b1001
	OpcodeLD       = 0b0010
	OpcodeLDI      = 0b1010
	OpcodeLEA      = 0b1110
	OpcodeRET      = 0b1100
	OpcodeRTI      = 0b1000
	OpcodeST       = 0b0011
	OpcodeSTI      = 0b1011
	OpcodeSTR      = 0b0111
	OpcodeTRAP     = 0b1111
	OpcodeReserved = 0b1101
)
