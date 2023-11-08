package vm

import (
	"errors"
	"fmt"
	"testing"
)

func TestRESV(tt *testing.T) {
	tt.Parallel()

	tt.Run("user-mode", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x3000
		cpu.PSR = StatusUser | StatusNormal | StatusNegative

		t.Log(cpu.PSR.String())
		cpu.REG[SP] = 0x2ff0
		cpu.SSP = 0x1200

		_ = cpu.Mem.store(Word(cpu.PC), 0b1101_0000_0000_0000)
		_ = cpu.Mem.store(Word(0x0101), 0x1100)
		_ = cpu.Mem.store(Word(0x1100), 0x1110)

		err := cpu.Step()

		t.Logf("err %T %#v", err, err)

		if errors.Is(err, ErrAccessControl) {
			t.Errorf("acv: %#v", err)
		}

		if err != nil {
			t.Errorf("err: %#v", err)
		}

		if cpu.IR.Opcode() != RESV {
			t.Errorf("instr: %s(%0#x), want: %s, got: %s(%v)",
				cpu.IR, cpu.IR, RESV, cpu.IR.Opcode(), cpu.IR.Opcode())
		}

		if cpu.PC != 0x1100 {
			t.Errorf("PC want: %0#x, got: %s", 0x1100, cpu.PC)
		}

		if cpu.PSR != (^StatusUser&StatusPrivilege)|StatusNormal|StatusNegative {
			t.Errorf("PSR want: %s, got: %s",
				(^StatusUser&StatusPrivilege)|StatusNormal|StatusNegative, cpu.PSR)
		}

		if cpu.REG[SP] != cpu.SSP-2 {
			t.Errorf("SP want: %s, got: %s", cpu.SSP, cpu.REG[SP])
		}

		if cpu.USP != 0x2ff0 {
			t.Errorf("USP want: %s, got: %s", Word(0x2ff0), cpu.USP)
		}
	})

	tt.Run("system-mode", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PSR = (StatusSystem & StatusPrivilege) | StatusNormal | StatusNegative
		cpu.REG[SP] = 0x2ff0
		cpu.SSP = 0x1200

		_ = cpu.Mem.store(Word(cpu.PC), 0b1101_0000_0000_0000)
		_ = cpu.Mem.store(Word(0x0101), 0x1100)
		_ = cpu.Mem.store(Word(0x1100), 0x1110)

		err := cpu.Step()

		t.Logf("err %T %#v", err, err)

		if err != nil {
			t.Errorf("err: %#v", err)
		}

		if cpu.IR.Opcode() != RESV {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, RESV, cpu.IR.Opcode())
		}

		if cpu.PC != 0x1100 {
			t.Errorf("PC want: %0#x, got: %s", 0x1000, cpu.PC)
		}

		if cpu.REG[SP] != Register(0x2ff0-2) {
			t.Errorf("SP want: %s, got: %s", Word(0x2fee)-2, cpu.REG[SP])
		}

		if cpu.USP != 0xfe00 {
			t.Errorf("USP want: %s, got: %s", Word(0xfe00), cpu.USP)
		}

		if cpu.PSR != StatusSystem|StatusNormal|StatusNegative {
			t.Errorf("PSR want: %s, got: %s",
				StatusSystem|StatusNormal|StatusNegative, cpu.PSR)
		}
	})
}

func TestInstructions(tt *testing.T) {
	tt.Parallel()

	tt.Run("BR", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		_ = cpu.Mem.store(Word(cpu.PC), 0b0000_010_0_0000_0111)
		cpu.PSR = StatusZero

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != BR {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, BR, op)
		}

		if cpu.PC != 0x3000+0x0008 {
			t.Errorf("PC want: %0#x, got: %s", 0x3008, cpu.PC)
		}

		if cpu.PSR != StatusZero {
			t.Errorf("cond incorrect, want: %s, got: %s",
				StatusZero, cpu.PSR)
		}
	})

	tt.Run("BRnzp", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0100
		_ = cpu.Mem.store(Word(cpu.PC), 0b0000_111_1_1111_0111)
		cpu.PSR.Set(0xf000)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != BR {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, BR, op)
		}

		if cpu.PC != 0x00f8 {
			t.Errorf("PC incorrect, want: %0#x, got: %s",
				0x00f8, cpu.PC)
		}

		if !cpu.PSR.Negative() {
			t.Errorf("cond incorrect, want: %v, got: %s",
				ConditionNegative, cpu.PSR.Cond())
		}
	})

	tt.Run("NOT", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.REG[R0] = 0b0101_1010_1111_0000
		_ = cpu.Mem.store(Word(cpu.PC), 0b1001_000_000_111111)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != NOT {
			t.Errorf("instr: %s, want: %b, got: %b", cpu.IR, NOT, op)
		}

		if cpu.REG[R0] != 0b1010_0101_0000_1111 {
			t.Errorf("r0 incorrect, want: %0b, got: %0b", 0b1010_0101_0000_1111, cpu.REG[R0])
		}

		if cpu.PSR.Cond() != ConditionNegative {
			t.Errorf("COND incorrect, want: %v, got: %s", ConditionNegative, cpu.PSR.Cond())
		}
	})

	tt.Run("AND", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		_ = cpu.Mem.store(Word(cpu.PC), 0b0101_000_000_0_00_001)
		cpu.REG[R0] = 0x5aff
		cpu.REG[R1] = 0x00f0

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if cpu.REG[R0] != 0x00f0 {
			t.Errorf("R0 incorrect, want: %0#16b, got: %0#b",
				0x0000, cpu.REG[R0])
		}

		if cpu.PSR.Cond() != ConditionPositive {
			t.Errorf("COND incorrect, want: %s, got: %s",
				ConditionPositive, cpu.PSR.Cond())
		}
	})

	tt.Run("ANDIMM", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		_ = cpu.Mem.store(Word(cpu.PC), 0b0101_000_000_1_10101)
		cpu.REG[R0] = 0b0101_1010_1111_1111

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != AND {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, AND, op)
		}

		if cpu.REG[R0] != 0x5af5 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 0x5af5, cpu.REG[R0])
		}

		if !cpu.PSR.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", StatusPositive, cpu.PSR)
		}
	})

	tt.Run("ADD", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		_ = cpu.Mem.store(Word(cpu.PC), 0b0001_000_000_0_00001)
		cpu.REG[R0] = 0
		cpu.REG[R1] = 1

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != ADD {
			t.Errorf("instr: %s, want: %04b, got: %04b", cpu.IR, AND, op)
		}

		oper := cpu.Decode()
		t.Logf("oper: %#+v", oper)

		if cpu.REG[R0] != 1 {
			t.Errorf("r0 incorrect, want: %016b, got: %016b", 1, cpu.REG[R0])
		}

		if !cpu.PSR.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s", StatusPositive, cpu.PSR)
		}
	})

	tt.Run("ADDIMM", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		_ = cpu.Mem.store(Word(cpu.PC), 0b0001_000_000_1_10000)
		cpu.REG[R0] = 0

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != ADD {
			t.Errorf("instr: %s, want: %04b, got: %04b",
				cpu.IR, AND, op)
		}

		oper := cpu.Decode()
		t.Logf("oper: %#+v", oper)

		if cpu.REG[R0] != 0xfff0 {
			t.Errorf("r0 incorrect, want: %s, got: %s",
				Register(0xfff0), cpu.REG[R0])
		}

		if !cpu.PSR.Negative() {
			t.Errorf("cond incorrect, want: %s, got: %s",
				StatusNegative, cpu.PSR)
		}
	})

	tt.Run("LD", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x00ff
		cpu.REG[R2] = 0xcafe
		_ = cpu.Mem.store(Word(cpu.PC), 0b0010_010_011000110)
		_ = cpu.Mem.store(Word(0x0100+0x00c6), 0x0f00)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != LD {
			t.Errorf("instr: %s, want: %04b, got: %04b",
				cpu.IR, LD, op)
		}

		if cpu.REG[R2] != 0x0f00 {
			t.Errorf("R2 incorrect, want: %s, got: %s",
				Register(0x0f00), cpu.REG[R2])
		}

		if !cpu.PSR.Positive() {
			t.Errorf("cond incorrect, want: %s, got: %s",
				StatusPositive, cpu.PSR)
		}
	})

	tt.Run("JMP", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x00ff
		_ = cpu.Mem.store(Word(cpu.PC), 0b1100_000_111_000000)
		cpu.REG[R7] = 0x0010

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != JMP {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, JMP, op)
		}

		if cpu.PC != 0x0010 {
			t.Errorf("PC incorrect, want: %s, got: %s",
				Register(0x0010), cpu.PC)
		}
	})

	tt.Run("JSRR", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0400
		_ = cpu.Mem.store(Word(cpu.PC), 0b0100_0_00_100_000000)
		cpu.REG[R4] = 0x0300
		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != JSR {
			t.Errorf("instr: %s, want: %s, got: %s",
				cpu.IR, JSR, op)
		}

		if cpu.PC != 0x0300 {
			t.Errorf("PC incorrect, want: %s, got: %s",
				Register(0x0300), cpu.PC)
		}

		if cpu.REG[R7] != 0x0401 {
			t.Errorf("R7 incorrect, want: %s, got: %s",
				Register(0x0401), cpu.PC)
		}
	})

	tt.Run("LDI", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0400
		_ = cpu.Mem.store(Word(cpu.PC), 0xa001)
		addr := Word(0x0402)
		_ = cpu.Mem.store(addr, 0xdad0)
		cpu.REG[R0] = 0xffff
		_ = cpu.Mem.store(0xdad0, 0xcafe)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		var r Register
		err = cpu.Mem.load(Word(0x0402), &r)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", r)

		if op := cpu.IR.Opcode(); op != LDI {
			t.Errorf("IR: %s, want: %s, got: %s",
				cpu.IR.String(), LDI, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("PC: want: %s, got: %s",
				Register(0x0401), cpu.PC)
		}

		if cpu.REG[R0] != 0xcafe {
			t.Errorf("R0 incorrect, want: %s, got: %s",
				Register(0xdad0), cpu.REG[R0])
		}

		if !cpu.PSR.Negative() {
			t.Errorf("COND incorrect, want: %s, got: %s",
				StatusNegative, cpu.PSR)
		}
	})

	tt.Run("LDR", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0400
		cpu.REG[R0] = 0xf0f0
		cpu.REG[R4] = 0x8000
		_ = cpu.Mem.store(Word(cpu.PC), 0b0110_000_100_00_0010)
		addr := Word(0x8000 + 0x0002)
		_ = cpu.Mem.store(addr, 0xdad0)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		var r Register
		err = cpu.Mem.load(addr, &r)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", r)

		if op := cpu.IR.Opcode(); op != LDR {
			t.Errorf("IR: %s, want: %s, got: %s",
				cpu.IR.String(), LDR, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("PC: want: %s, got: %s",
				Register(0x0401), cpu.PC)
		}

		if cpu.REG[R0] != 0xdad0 {
			t.Errorf("R0 incorrect, want: %s, got: %s",
				Register(0xdad0), cpu.REG[R0])
		}

		if !cpu.PSR.Negative() {
			t.Errorf("COND incorrect, want: %s, got: %s",
				StatusNegative, cpu.PSR)
		}
	})

	tt.Run("LEA", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		initialStatus := cpu.PSR

		cpu.PC = 0x0400
		_ = cpu.Mem.store(Word(cpu.PC), 0b1110_000_1_00000000)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != LEA {
			t.Errorf("IR: %s, want: %s, got: %s", cpu.IR.String(), LEA, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("R0 incorrect, want: %s, got: %s",
				Register(0x0401), cpu.PC)
		}

		if initialStatus != cpu.PSR {
			t.Errorf("condition codes should not chance, want: %s, got: %s", initialStatus, cpu.PSR)
		}
	})

	tt.Run("ST", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0400
		cpu.REG[R7] = 0xcafe

		_ = cpu.Mem.store(Word(cpu.PC), 0b0011_111_0_1000_0000)
		_ = cpu.Mem.store(Word(0x0481), 0x0f00)

		initialStatus := cpu.PSR

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != ST {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, ST, op)
		}

		var val Register
		err = cpu.Mem.load(Word(0x0481), &val)
		if err != nil {
			t.Error(err)
		}

		if val != 0xcafe {
			t.Errorf("Mem[%s] want: %s, got: %s",
				Word(0x0481), Word(0xcafe), val)
		}

		if cpu.REG[R7] != 0xcafe {
			t.Errorf("R7 incorrect, want: %s, got: %s",
				Register(0xcafe), cpu.REG[R7])
		}

		if initialStatus != cpu.PSR {
			t.Errorf("condition codes should not change, want: %s, got: %s", initialStatus, cpu.PSR)
		}
	})

	tt.Run("STI", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x0400
		cpu.REG[RETP] = 0xcafe

		_ = cpu.Mem.store(Word(cpu.PC), 0b1011_111_0_0000_0001)
		_ = cpu.Mem.store(Word(cpu.PC)+2, 0x0f00)
		_ = cpu.Mem.store(Word(0x0f00), 0x0fff)

		initialPSR := cpu.PSR

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != STI {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, STI, op)
		}

		var val Register
		err = cpu.Mem.load(Word(0x0f00), &val)

		if err != nil {
			t.Error(err)
		}

		if val != 0xcafe {
			t.Errorf("Mem[%s] want: %s, got: %s", Word(0x0f00), Word(0xcafe), val)
		}

		// Condition codes should not change
		if initialPSR != cpu.PSR {
			t.Errorf("cond incorrect, want: %s, got: %s", initialPSR, cpu.PSR)
		}
	})

	tt.Run("TRAP USER", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x4050
		cpu.PSR = StatusUser | StatusZero
		cpu.SSP = 0x3000
		cpu.USP = 0xface
		cpu.REG[SP] = 0xfe00
		_ = cpu.Mem.store(Word(cpu.PC), 0b1111_0000_1000_0000)
		_ = cpu.Mem.store(Word(0x0080), 0xadad)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != TRAP {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, TRAP, op)
		}

		if cpu.PC != 0xadad {
			t.Errorf("PC want: %s, got: %s",
				ProgramCounter(0xadad), cpu.PC)
		}

		if cpu.USP != 0xfe00 {
			t.Errorf("USP want: R6 = %s, got: %s",
				Register(0xfe00), cpu.USP)
		}

		if cpu.SSP != 0x3000 {
			t.Errorf("SSP want: %s, got: %s",
				Word(0x3000), cpu.SSP)
		}

		if cpu.REG[SP] != 0x3000-2 {
			t.Errorf("SP want: %s, got: %s",
				Word(0x2ffe), cpu.REG[SP])
		}

		var val Register
		err = cpu.Mem.load(Word(cpu.REG[SP]), &val)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", val)

		if val != 0x4051 {
			t.Errorf("SP top want: %s <= PC, got: %s",
				Register(0x4051), val)
		}

		err = cpu.Mem.load(Word(cpu.REG[SP]+1), &val)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", val)

		if val != 0x8002 {
			t.Errorf("SP bottom want: %s <= PSR, got: %s",
				StatusZero&^StatusPrivilege, ProcessorStatus(val))
		}

		if cpu.PSR.Privilege() != PrivilegeSystem {
			t.Errorf("PSR want: %s, got: %s",
				ProcessorStatus(0x0000), cpu.PSR)
		}
	})

	tt.Run("TRAP SYSTEM", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x20ff
		cpu.PSR = (^StatusUser & StatusPrivilege) | StatusNormal | StatusZero
		t.Log(cpu.PSR.String())
		cpu.USP = 0xffff
		cpu.SSP = 0x1f00
		cpu.REG[SP] = 0x1e00
		_ = cpu.Mem.store(Word(cpu.PC), 0b1111_0000_1000_0000)
		_ = cpu.Mem.store(Word(0x0080), 0xadad)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != TRAP {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, TRAP, op)
		}

		if cpu.PC != 0xadad {
			t.Errorf("PC want: %s, got: %s",
				ProgramCounter(0xadad), cpu.PC)
		}

		if cpu.USP != 0xffff {
			t.Errorf("USP want: %s, got: %s",
				Register(0xffff), cpu.USP)
		}

		if cpu.SSP != 0x1f00 {
			t.Errorf("SSP want: %s, got: %s",
				Word(0x2f00), cpu.SSP)
		}

		if cpu.REG[SP] != 0x1e00-2 {
			t.Errorf("SP want: %s, got: %s",
				Word(0x1dfe), cpu.REG[SP])
		}

		var val Register
		err = cpu.Mem.load(Word(cpu.REG[SP]), &val)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", val)

		if val != 0x2100 {
			t.Errorf("SP top want: %s <= PC, got: %s",
				Register(0x2100), val)
		}

		err = cpu.Mem.load(Word(cpu.REG[SP]+1), &val)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", val)

		if Word(val) != Word((^StatusUser&StatusPrivilege)|StatusNormal|StatusZero) {
			t.Errorf("SP bottom want PSR: %s, got: %s",
				(^StatusUser&StatusPrivilege)|StatusNormal|StatusZero,
				val)
		}

		if cpu.PSR.Privilege() != PrivilegeSystem {
			t.Errorf("PSR want: %s, got: %s",
				ProcessorStatus(0x0000), cpu.PSR)
		}
	})

	tt.Run("RTI to USER", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0xadaf
		_ = cpu.Mem.store(Word(cpu.PC), 0b1000_0000_0000_0000)
		cpu.PSR = ^StatusUser | StatusNegative
		cpu.SSP = 0xffff

		cpu.REG[SP] = 0x3000 - 2 // user PC, PSR on system stack
		_ = cpu.Mem.store(Word(cpu.REG[SP]), 0x0401)
		_ = cpu.Mem.store(Word(cpu.REG[SP]+1),
			Word(StatusUser|StatusNegative))

		cpu.USP = 0xfade // previous stored user stack
		_ = cpu.Mem.store(Word(cpu.USP), 0xff00)
		_ = cpu.Mem.store(Word(cpu.USP+1), 0x0ff0)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != RTI {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, RTI, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("PC want: %s, got: %s",
				ProgramCounter(0x0401), cpu.PC)
		}

		if cpu.PSR.Privilege() != PrivilegeUser {
			t.Errorf("PSR want: %s, got: %s",
				ProcessorStatus(0x8004), cpu.PSR)
		}

		if cpu.USP != 0xfade {
			t.Errorf("USP want: %s, got: %s",
				Register(0xfade), cpu.USP)
		}
		if cpu.SSP != 0x3000 {
			t.Errorf("SSP want: %s, got: %s",
				Word(0x3000), cpu.SSP)
		}

		if cpu.REG[SP] != 0xfade {
			t.Errorf("SP want: USP=%s, got: %s",
				Word(0xfade), cpu.REG[SP])
		}

		var top Register
		err = cpu.Mem.load(Word(cpu.REG[SP]), &top)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", top)

		if top != 0xff00 {
			t.Errorf("SP top want: %s, got: %s",
				Word(0xff00), top)
		}

		var bottom Register
		err = cpu.Mem.load(Word(cpu.REG[SP]+1), &bottom)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", bottom)

		if bottom != 0x0ff0 {
			t.Errorf("SP bottom want: %s, got: %s",
				Word(0x0ff0), bottom)
		}
	})

	tt.Run("RTI to SYSTEM", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0xadaf
		_ = cpu.Mem.store(Word(cpu.PC), 0b1000_0000_0000_0000)
		_ = cpu.Mem.store(Word(0x0080), 0xcafe)
		cpu.PSR = ^StatusUser | StatusNegative
		cpu.SSP = 0xffff

		cpu.REG[SP] = 0x2f00 - 2 // system PC, PSR on system stack
		_ = cpu.Mem.store(Word(cpu.REG[SP]), 0x0401)
		_ = cpu.Mem.store(Word(cpu.REG[SP]+1),
			Word(StatusSystem|StatusZero))

		// Values on old system stack.
		_ = cpu.Mem.store(Word(cpu.REG[SP]+2), 0x1111)
		_ = cpu.Mem.store(Word(cpu.REG[SP]+3), 0x2222)

		cpu.USP = 0x4200 // Previous stored user stack.
		_ = cpu.Mem.store(Word(cpu.USP), 0xff00)
		_ = cpu.Mem.store(Word(cpu.USP+1), 0x0ff0)

		err := cpu.Step()
		if err != nil {
			t.Error(err)
		}

		if op := cpu.IR.Opcode(); op != RTI {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, RTI, op)
		}

		if cpu.PC != 0x0401 {
			t.Errorf("PC want: %s, got: %s",
				ProgramCounter(0x0401), cpu.PC)
		}

		if cpu.PSR != StatusSystem|StatusZero {
			t.Errorf("PSR want: %s, got: %s",
				ProcessorStatus(0x0002), cpu.PSR)
		}

		if cpu.USP != 0x4200 {
			t.Errorf("USP want: %s, got: %s",
				Register(0x4200), cpu.USP)
		}

		if cpu.SSP != 0xffff {
			t.Errorf("SSP want: %s, got: %s",
				Word(0xffff), cpu.SSP)
		}

		if cpu.REG[SP] != 0x2f00 {
			t.Errorf("SP want: %s, got: %s",
				Word(0x3000), cpu.REG[SP])
		}

		var top Register
		err = cpu.Mem.load(Word(cpu.REG[SP]), &top)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", top)

		if top != 0x1111 {
			t.Errorf("SP top want: %s, got: %s",
				Word(0x1111), top)
		}

		var bottom Register
		err = cpu.Mem.load(Word(cpu.REG[SP]+1), &bottom)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", top)

		if bottom != 0x2222 {
			t.Errorf("SP bottom want: %s, got: %s",
				Word(0x2222), bottom)
		}
	})

	tt.Run("RTI as USER", func(tt *testing.T) {
		var (
			t   = NewTestHarness(tt)
			cpu = t.Make()
		)

		cpu.PC = 0x3300 // User space PC
		_ = cpu.Mem.store(Word(cpu.PC), 0b1000_0000_0000_0000)
		cpu.PSR = StatusUser | StatusNormal | StatusNegative
		cpu.SSP = 0x1a1a

		cpu.REG[SP] = 0x2f00 - 2 // some data on stack
		_ = cpu.Mem.store(Word(cpu.REG[SP]), 0x0001)
		_ = cpu.Mem.store(Word(cpu.REG[SP]+1), 0xface)

		cpu.USP = 0xffff                        // Invalid user stack pointer
		_ = cpu.Mem.store(Word(0x0100), 0x1234) // PMV table points to handler

		err := cpu.Step()
		if err != nil {
			t.Errorf("unhandled instruction error: %v", err)
		}

		if op := cpu.IR.Opcode(); op != RTI {
			t.Errorf("IR: %s, want: %0#4b, got: %0#4b",
				cpu.IR, RTI, op)
		}

		if cpu.PC != 0x1234 {
			t.Errorf("PC want: %s, got: %s",
				ProgramCounter(0x1234), cpu.PC)
		}

		if cpu.PSR != (StatusPrivilege^StatusUser)|StatusNormal|StatusNegative {
			t.Errorf("PSR want: %s, got: %s",
				ProcessorStatus(0x0004), cpu.PSR)
		}

		if cpu.USP != 0x2f00-2 {
			t.Errorf("USP want: %s, got: %s",
				Register(0x2efe), cpu.USP)
		}

		if cpu.SSP != 0x1a1a {
			t.Errorf("SSP want: %s, got: %s",
				Word(0x1a1a), cpu.SSP)
		}

		if cpu.REG[SP] != 0x1a1a-2 {
			t.Errorf("SP want: %s, got: %s",
				Word(0x1a18), cpu.REG[SP])
		}

		var top Register
		err = cpu.Mem.load(Word(cpu.REG[SP]), &top)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", top)

		if top != 0x3301 {
			t.Errorf("SP top want: %s, got: %s",
				Word(0x3301), top)
		}

		var bottom Register
		err = cpu.Mem.load(Word(cpu.REG[SP]+1), &bottom)
		if err != nil {
			t.Error(err)
		}
		t.Logf("mem: %0#v", bottom)

		if Word(bottom) != Word((StatusPrivilege&StatusUser)|StatusNormal|StatusNegative) {
			t.Errorf("SP bottom want: %s, got: %s",
				(StatusPrivilege&StatusUser)|StatusLow|StatusNegative, bottom)
		}
	})
}

func TestSext(tt *testing.T) {
	tt.Parallel()

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
		tt.Run(name, func(tt *testing.T) {
			t := NewTestHarness(tt)

			got := Word(tc.have)
			got.Sext(tc.bits)

			if got != Word(tc.want) {
				t.Errorf("got: %016b want: %016b", got, tc.want)
			}
		})
	}
}
