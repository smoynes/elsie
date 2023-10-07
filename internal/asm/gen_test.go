package asm_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/smoynes/elsie/internal/asm"
)

type generatorHarness struct {
	*testing.T
}

func TestGenerator(tt *testing.T) {
	t := generatorHarness{tt}

	var buf bytes.Buffer

	symbols := SymbolTable{}
	syntax := make(SyntaxTable, 0)

	syntax.Add(&ORIG{LITERAL: 0x3000})
	syntax.Add(&NOT{DR: "R0", SR: "R7"})

	gen := NewGenerator(symbols, syntax)

	count, err := gen.WriteTo(&buf)
	if err != nil {
		t.Error(err)
	}

	expected := []byte{ // big endian
		0x30, 0x00,
		0x91, 0xff,
	}

	bytes := buf.Bytes()

	if count != int64(len(expected)) {
		t.Error("expected: 4 bytes")
	}

	for i := range expected {
		if expected[i] != bytes[i] {
			t.Errorf("not equal: buf[%02x]: want: %0#x, got: %0#x", i, expected[i], bytes[i])
		}
	}
}

func TestAND_Generate(t *testing.T) {
	tcs := []struct {
		oper Operation
		want uint16
	}{
		{oper: &AND{DR: "R3", SR1: "R4", SR2: "R6"}, want: 0x5706},
		{oper: &AND{DR: "R0", SR1: "R7", SYMBOL: "LABEL"}, want: 0x51e6},
		{oper: &AND{DR: "R1", SR1: "R2", OFFSET: 0x12}, want: 0x52b2},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3007,
	}

	for i := range tcs {
		oper, want := tcs[i].oper, tcs[i].want

		if mc, err := oper.Generate(symbols, pc); err != nil {
			t.Errorf("Code: %#v == error  ==> %0#4x %s", oper, mc, err)
		} else {
			t.Logf("Code: %#v == generated ==> %0#4x", oper, mc)

			if mc == nil {
				t.Error("invalid machine code")
			}

			if len(mc) != 1 {
				t.Errorf("incorrect machine code: %d bytes", len(mc))
			}

			if mc[0] != want {
				t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, mc)
			}
		}
	}
}

func TestBR_Generate(t *testing.T) {
	tcs := []struct {
		i       Operation
		mc      uint16
		wantErr *OffsetError
	}{
		{&BR{NZP: 0x7, OFFSET: 0x01}, 0x0e01, nil},
		{&BR{NZP: 0x2, OFFSET: 0xfff0}, 0x05f0, nil},
		{&BR{NZP: 0x3, SYMBOL: "LABEL"}, 0x0604, nil},
		{&BR{NZP: 0x3, SYMBOL: "BACK"}, 0x061f, nil},
		{&BR{NZP: 0x4, SYMBOL: "LONG"}, 0x061f, &OffsetError{}},
	}

	pc := uint16(0x3000)

	symbols := SymbolTable{
		"LABEL": 0x3005,
		"BACK":  0x3000,
		"LONG":  0x2fe0,
	}

	for i := range tcs {
		op, want, wantErr := tcs[i].i, tcs[i].mc, tcs[i].wantErr
		got, err := op.Generate(symbols, pc)

		if wantErr != nil && !errors.As(err, &wantErr) {
			t.Logf("err: %#v", err)
			t.Errorf("expected error: %#v, got: %#v", wantErr, err)
		} else if wantErr == nil && err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if wantErr == nil && err == nil {
			t.Logf("Code: %#v == generated ==> %0#4x", op, got)

			if got == nil {
				t.Error("invalid machine code")
			}

			if len(got) != 1 {
				t.Errorf("incorrect machine code: %d bytes", len(got))
			}

			if got[0] != want {
				t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, got)
			}
		}
	}
}

func TestLDR_Generate(t *testing.T) {
	instrs := []Operation{
		&LDR{DR: "R0", SR: "R5", OFFSET: 0x10},
		&LDR{DR: "R7", SR: "R4", SYMBOL: "LABEL"},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x300a,
	}

	if mc, err := instrs[0].Generate(symbols, pc); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[0], mc)

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}
		want := uint16(0x6150)

		if mc[0] != want {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, mc)
		}
	}

	if mc, err := instrs[1].Generate(symbols, pc); err != nil {
		t.Fatalf("Code: %#v == error    ==> %s", instrs[1], err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[1], mc)

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		want := uint16(0x6f09)
		if mc[0] != want {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, mc)
		}
	}
}

func TestLD_Generate(t *testing.T) {
	instrs := []Operation{
		&LD{DR: "R0", OFFSET: 0x10},
		&LD{DR: "R7", SYMBOL: "LABEL"},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3100,
	}

	if mc, err := instrs[0].Generate(symbols, pc); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[0], mc)

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		want := uint16(0x2010)
		if mc[0] != want {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, mc)
		}
	}

	if mc, err := instrs[1].Generate(symbols, pc); err != nil {
		t.Fatalf("Code: %#v == error    ==> %s", instrs[1], err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[1], mc)

		if mc == nil {
			t.Error("invalid machine code")
		}

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		want := uint16(0x2eff)
		if mc[0] != want {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", want, mc)
		}
	}
}

func TestADD_Generate(t *testing.T) {
	tcs := []struct {
		operation Operation
		mc        uint16
	}{
		{&ADD{DR: "R0", SR1: "R0", SR2: "R1"}, 0x1001},
		{&ADD{DR: "R1", SR1: "R1", LITERAL: 0x000f}, 0x124f},
		{&ADD{DR: "R1", SR1: "R1", SR2: "R0"}, 0x1240},
		{&ADD{DR: "R0", SR1: "R7", LITERAL: 0b0000_0000_0000_1010}, 0b0001_0001_1100_1010},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3100,
	}

	for tc := range tcs {
		op, exp := tcs[tc].operation, tcs[tc].mc

		mc, err := op.Generate(symbols, pc)
		if err != nil {
			t.Fatal(err)
		}

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		if mc[0] != exp {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", exp, mc)
		}
	}
}

func TestNOT_Generate(t *testing.T) {
	tcs := []struct {
		op   Operation
		want uint16
	}{
		{
			op:   &NOT{DR: "R1", SR: "R1"},
			want: 0x927f,
		},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{}

	for tc := range tcs {
		op, exp := tcs[tc].op, tcs[tc].want

		mc, err := op.Generate(symbols, pc)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		if mc[0] != exp {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", exp, mc)
		}
	}
}
