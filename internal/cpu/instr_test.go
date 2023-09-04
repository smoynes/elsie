package cpu

import (
	"fmt"
	"testing"
)

func TestInstructions(t *testing.T) {
	t.Run("Reserved", func(t *testing.T) {
		t.Parallel()
		var instr Instruction = 0b11010011_10100111
		var op Opcode = instr.Decode()

		if op != OpcodeReserved {
			t.Errorf("instr: %s, want: %b, got: %b", instr, OpcodeNOT, op)
		}
	})

	t.Run("BR", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0000_010_0_0000_0111
		cpu.Proc.Cond = ConditionZero

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeBR {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeBR, op)
		}

		if cpu.PC != 0x3000+0x0008 {
			t.Errorf("PC incorrect, want: %0b, got: %0b", 0x3000+0x0008, cpu.PC)
		}

		if cpu.Proc.Cond != ConditionZero {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionZero, cpu.Proc.Cond)
		}
	})

	t.Run("BRnzp", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0000_111_1_1111_0111
		cpu.Proc.Cond = ConditionZero

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeBR {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeBR, op)
		}

		if cpu.PC != 0x3000+1-0x0009 {
			t.Errorf("PC incorrect, want: %0b, got: %0b", 0x3000+1-0x0009, cpu.PC)
		}

		if cpu.Proc.Cond != ConditionZero {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionZero, cpu.Proc.Cond)
		}
	})

	t.Run("NOT", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Proc.Reg[R0] = 0b0101_1010_1111_0000
		cpu.Mem[cpu.PC] = 0b1001_000_000_111111

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeNOT {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeNOT, op)
		}

		if cpu.Proc.Reg[R0] != 0b1010_0101_0000_1111 {
			t.Errorf("r0 incorrect, want: %0b, got: %0b", 0b1010_0101_0000_1111, cpu.Proc.Reg[R0])
		}

		if cpu.Proc.Cond != ConditionNegative {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionNegative, cpu.Proc.Cond)
		}
	})

	t.Run("AND", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0101_000_000_0_00_001
		cpu.Proc.Reg[R0] = 0b0101_1010_1111_0000
		cpu.Proc.Reg[R1] = 0x0000

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeAND {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, OpcodeAND, op)
		}

		if cpu.Proc.Reg[R0] != 0x0000 {
			t.Errorf("r0 incorrect, want: %0b, got: %0b", 0b1010_0101_0000_1111, cpu.Proc.Reg[R0])
		}

		if cpu.Proc.Cond != ConditionZero {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionZero, cpu.Proc.Cond)
		}
	})

	t.Run("ANDIMM", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0101_000_000_1_00101
		cpu.Proc.Reg[R0] = 0b0101_1010_1111_1111

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeAND {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, OpcodeAND, op)
		}

		if cpu.Proc.Reg[R0] != 0x0005 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 0x0003, cpu.Proc.Reg[R0])
		}

		if !cpu.Proc.Cond.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionPositive, cpu.Proc.Cond)
		}
	})

	t.Run("ANDIMM SEXT", func(t *testing.T) {
		t.Parallel()
		var cpu *LC3 = New()
		cpu.Mem[cpu.PC] = 0b0101_000_000_1_10101
		cpu.Proc.Reg[R0] = 0b0101_1010_1111_1111

		err := cpu.Execute()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Decode(); op != OpcodeAND {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, OpcodeAND, op)
		}

		if cpu.Proc.Reg[R0] != 0x5af5 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 0x5af5, cpu.Proc.Reg[R0])
		}

		if !cpu.Proc.Cond.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", ConditionPositive, cpu.Proc.Cond)
		}
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
			p := Processor{}
			got := p.sext(Word(tc.have), tc.bits)

			if got != Word(tc.want) {
				t.Errorf("got: %016b want: %016b", got, tc.want)
			}
		})
	}
}
