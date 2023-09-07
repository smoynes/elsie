package cpu

import (
	"fmt"
	"testing"
)

func TestInstructions(t *testing.T) {
	t.Run("Reserved", func(t *testing.T) {
		var instr Instruction = 0b11010011_10100111
		cpu := New()
		cpu.IR = instr

		var op operation = cpu.Decode()
		if op.opcode() != OpcodeReserved {
			t.Errorf("instr: %s, want: %b, got: %b", instr, OpcodeReserved, op)
		}
	})

	t.Run("BR", func(t *testing.T) {
		cpu := New()
		cpu.Mem[cpu.PC] = 0b0000_010_0_0000_0111
		cpu.Cond = ConditionZero

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeBR {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeBR, op)
		}

		if cpu.PC != 0x3000+0x0008 {
			t.Errorf("PC incorrect, want: %0b, got: %0b", 0x3000+0x0008, cpu.PC)
		}

		if cpu.Cond != ConditionZero {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionZero, cpu.Cond)
		}
	})

	t.Run("BRnzp", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0000_111_1_1111_0111
		cpu.Cond = ConditionZero

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeBR {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeBR, op)
		}

		if cpu.PC != 0x3000+1-0x0009 {
			t.Errorf("PC incorrect, want: %0b, got: %0b", 0x3000+1-0x0009, cpu.PC)
		}

		if cpu.Cond != ConditionZero {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionZero, cpu.Cond)
		}
	})

	t.Run("NOT", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Reg[R0] = 0b0101_1010_1111_0000
		cpu.Mem[cpu.PC] = 0b1001_000_000_111111

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeNOT {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeNOT, op)
		}

		if cpu.Reg[R0] != 0b1010_0101_0000_1111 {
			t.Errorf("r0 incorrect, want: %0b, got: %0b", 0b1010_0101_0000_1111, cpu.Reg[R0])
		}

		if cpu.Cond != ConditionNegative {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionNegative, cpu.Cond)
		}
	})

	t.Run("AND", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0101_000_000_0_00_001
		cpu.Reg[R0] = 0b0101_1010_1111_1111
		cpu.Reg[R1] = 0x00f0

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if cpu.Reg[R0] != 0x00f0 {
			t.Errorf("R0 incorrect, want: %0#16b, got: %0#b", 0x0000, cpu.Reg[R0])
		}

		if cpu.Cond != ConditionPositive {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionPositive, cpu.Cond)
		}
	})

	t.Run("ANDIMM", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0101_000_000_1_10101
		cpu.Reg[R0] = 0b0101_1010_1111_1111

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeAND {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, OpcodeAND, op)
		}

		if cpu.Reg[R0] != 0x5af5 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 0x5af5, cpu.Reg[R0])
		}

		if !cpu.Cond.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionPositive, cpu.Cond)
		}
	})

	t.Run("ADD", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0001_000_000_0_00001
		cpu.Reg[R0] = 0
		cpu.Reg[R1] = 1

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeADD {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, OpcodeAND, op)
		}

		oper := cpu.Decode()
		t.Logf("oper: %#+v", oper)

		if cpu.Reg[R0] != 1 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 1, cpu.Reg[R0])
		}

		if !cpu.Cond.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionPositive, cpu.Cond)
		}
	})

	t.Run("ADDIMM", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0001_000_000_1_10000
		cpu.Reg[R0] = 0

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeADD {
			t.Errorf("instr: %s, want: %04b, got: %04b",
				cpu.IR, OpcodeAND, op)
		}

		oper := cpu.Decode()
		t.Logf("oper: %#+v", oper)

		if cpu.Reg[R0] != 0xfff0 {
			t.Errorf("r0 incorrect, want: %d (%s), got: %d (%s)",
				Register(0xfff0), Register(0xfff0),
				cpu.Reg[R0], cpu.Reg[R0])
		}

		if !cpu.Cond.Negative() {
			t.Errorf("cond incorrect, want: %s, got: %s",
				ConditionNegative, cpu.Cond)
		}
	})

	t.Run("LD", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.PC = 0x00ff
		cpu.Reg[R2] = 0xcafe
		cpu.Mem[cpu.PC] = 0b0010_010_011000110
		cpu.Mem[0x0100+0x00c6] = 0x0f00

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeLD {
			t.Errorf("instr: %s, want: %04b, got: %04b",
				cpu.IR, OpcodeLD, op)
		}

		if cpu.Reg[R2] != 0x0f00 {
			t.Errorf("R2 incorrect, want: %d (%s), got: %d (%s)",
				Register(0x0f00), Register(0x0f00),
				cpu.Reg[R2], cpu.Reg[R2])
		}

		if !cpu.Cond.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s",
				ConditionPositive, cpu.Cond)
		}

		oper := cpu.Decode().(*ld)
		oper.EvalAddress(cpu)
		oper.FetchOperands(cpu)
		t.Logf("oper: %#+v", oper)
	})

	t.Run("JMP", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.PC = 0x00ff
		cpu.Mem[cpu.PC] = 0b1100_000_111_000000
		cpu.Reg[R7] = 0x0010

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeJMP {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, OpcodeJMP, op)
		}

		if cpu.PC != 0x0010 {
			t.Errorf("PC incorrect, want: %d (%s), got: %d (%s)",
				Register(0x0010), Register(0x0010),
				cpu.PC, cpu.PC)
		}

		oper := cpu.Decode().(*jmp)
		oper.Execute(cpu)
		t.Logf("oper: %#+v", oper)
	})

	t.Run("JSRR", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.PC = 0x0400
		cpu.Mem[cpu.PC] = 0b0100_0_00_100_000000
		cpu.Reg[R4] = 0x0300
		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeJSR {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, OpcodeJSR, op)
		}

		if cpu.PC != 0x0300 {
			t.Errorf("PC incorrect, want: %d (%s), got: %d (%s)",
				Register(0x0300), Register(0x0300),
				cpu.PC, cpu.PC)
		}

		if cpu.Reg[R7] != 0x0401 {
			t.Errorf("R7 incorrect, want: %d (%s), got: %d (%s)",
				Register(0x0401), Register(0x0401),
				cpu.PC, cpu.PC)
		}

		oper := cpu.Decode().(*jsrr)
		t.Logf("oper: %#+v", oper)
	})

	t.Run("LDI", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.PC = 0x0400
		cpu.Mem[cpu.PC] = 0xa001
		addr := Word(0x0402)
		cpu.Mem[addr] = 0xdad0
		cpu.Reg[R0] = 0xffff
		cpu.Mem[0xdad0] = 0xcafe

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		t.Logf("mem: %0#v", cpu.Mem[0x0402])

		if op := cpu.IR.Opcode(); op != OpcodeLDI {
			t.Errorf("IR: %s, want: %s, got: %s",
				cpu.IR.String(), OpcodeLDI, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("PC: want: %d (%s), got: %d (%s)",
				Register(0x0401), Register(0x0401),
				cpu.PC, cpu.PC)
		}

		if cpu.Reg[R0] != 0xcafe {
			t.Errorf("R0 incorrect, want: %d (%s), got: %d (%s)",
				Register(0xdad0), Register(0xdad0),
				cpu.Reg[R0], cpu.Reg[R0])
		}

		if !cpu.Cond.Negative() {
			t.Errorf("COND incorrect, want: %s, got: %s",
				ConditionNegative, cpu.Cond)
		}
	})

	t.Run("ST", func(t *testing.T) {
		var cpu *LC3 = New()
		cpu.PC = 0x0400
		cpu.Reg[R7] = 0xcafe
		cpu.Mem[cpu.PC] = 0b0011_111_0_1000_0000
		cpu.Mem[0x0481] = 0x0f00

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != OpcodeST {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, OpcodeST, op)
		}

		val := cpu.Mem[0x0481]
		if val != 0xcafe {
			t.Errorf("Mem[%s] want: %s, got: %s",
				Word(0x0481), Word(0xcafe), val)
		}

		if cpu.Reg[R7] != 0xcafe {
			t.Errorf("R7 incorrect, want: %d (%s), got: %d (%s)",
				Register(0xcafe), Register(0xcafe),
				cpu.Reg[R7], cpu.Reg[R7])
		}

		if !cpu.Cond.Zero() {
			t.Errorf("cond incorrect, want: %s, got: %s",
				ConditionZero, cpu.Cond)
		}

		oper := cpu.Decode().(*st)
		oper.EvalAddress(cpu)
		oper.StoreResult(cpu)
		t.Logf("oper: %#+v", oper)
	})

}

func TestSext(t *testing.T) {
	tcs := []struct {
		have uint16
		bits uint8
		want uint16
	}{
		{
			have: 0x000e,
			bits: 4,
			want: 0xfffe,
		},
		{
			have: 0x0000,
			bits: 1,
			want: 0x0000,
		},
		{
			have: 0x8000,
			bits: 1,
			want: 0x0000,
		},
		{
			have: 0x0001,
			bits: 1,
			want: 0xffff,
		},
		{
			have: 0x0001,
			bits: 2,
			want: 0x0001,
		},
		{
			have: 0x0003,
			bits: 1,
			want: 0xffff,
		},
		{
			have: 0xf00f,
			bits: 6,
			want: 0x000f,
		},
		{
			have: 0xf01e,
			bits: 6,
			want: 0x001e,
		},
		{
			have: 0xf03e,
			bits: 6,
			want: 0xfffe,
		},
		{
			have: 0xf02e,
			bits: 6,
			want: 0xffee,
		},
		{
			have: 0xf070,
			bits: 6,
			want: 0xfff0,
		},
		{
			have: 0x0001,
			bits: 0,
			want: 0x0000,
		},
		{
			have: 0xffff,
			bits: 0,
			want: 0x0000,
		},
	}

	for _, tc := range tcs {
		tc := tc
		name := fmt.Sprintf("%0#4x %d", tc.have, tc.bits)
		t.Run(name, func(t *testing.T) {
			got := Word(tc.have)
			got.Sext(tc.bits)

			if got != Word(tc.want) {
				t.Errorf("got: %016b want: %016b", got, tc.want)
			}
		})
	}
}
