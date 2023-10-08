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

type generateCase struct {
	oper    Operation
	want    uint16
	wantErr error
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

func TestAND_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &AND{DR: "R3", SR1: "R4", SR2: "R6"}, want: 0x5706},
		{oper: &AND{DR: "R0", SR1: "R7", SYMBOL: "LABEL"}, want: 0x51e7},
		{oper: &AND{DR: "R1", SR1: "R2", OFFSET: 0x12}, want: 0x52b2},
		{oper: &AND{DR: "BAD", SR1: "R0", OFFSET: 0x12}, wantErr: &RegisterError{Reg: "BAD"}},
		{oper: &AND{DR: "R7", SR1: "BAD", OFFSET: 0x12}, wantErr: &RegisterError{Reg: "BAD"}},
		{oper: &AND{DR: "R0", SR1: "R0", SR2: "R9"}, wantErr: &RegisterError{Reg: "R9"}},
		{oper: &AND{DR: "R0", SR1: "R0", SYMBOL: "BACK"}, want: 0x503e},
		{oper: &AND{DR: "R0", SR1: "R0", SYMBOL: "FAR"}, wantErr: &OffsetError{Offset: 0xffd0}},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3007,
		"BACK":  0x2ffe,
		"FAR":   0x2fd0,
	}

	t.Run(pc, symbols, tcs)
}

func TestBR_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{&BR{NZP: 0x7, OFFSET: 0x01}, 0x0e01, nil},
		{&BR{NZP: 0x2, OFFSET: 0xfff0}, 0x05f0, nil},
		{&BR{NZP: 0x3, SYMBOL: "LABEL"}, 0x0605, nil},
		{&BR{NZP: 0x3, SYMBOL: "BACK"}, 0x0600, nil},
		{&BR{NZP: 0x4, SYMBOL: "LONG"}, 0x061f, &OffsetError{Offset: 0xff00}},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"LONG":   0x2f00,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

func TestLDR_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{&LDR{DR: "R0", SR: "R5", OFFSET: 0x10}, 0x6150, nil},
		{&LDR{DR: "R7", SR: "R4", SYMBOL: "LABEL"}, 0x6f05, nil},
		{&LDR{DR: "R5", SR: "R1", SYMBOL: "BACK"}, 0x6a40, nil},
		{&LDR{DR: "R3", SR: "R2", SYMBOL: "GONE"}, 0, &SymbolError{0x3000, "GONE"}},
		{&LDR{DR: "R1", SR: "R3", SYMBOL: "FAR"}, 0, &OffsetError{Offset: 0xff00}},
		{&LDR{DR: "R2", SR: "R4", SYMBOL: "YONDER"}, 0, &OffsetError{Offset: 0x1000}},
		{&LDR{DR: "R8", SR: "R2", SYMBOL: "LABEL"}, 0, &RegisterError{Reg: "R8"}},
		{&LDR{DR: "R0", SR: "DR", SYMBOL: "LABEL"}, 0, &RegisterError{Reg: "DR"}},
	}
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"FAR":    0x2f00,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

// Run tests a collection of generator tests cases.
func (t *generatorHarness) Run(pc uint16, symbols SymbolTable, tcs []generateCase) {
	t.Helper()

	for i := range tcs {
		oper, want, expErr := tcs[i].oper, tcs[i].want, tcs[i].wantErr

		if mc, err := oper.Generate(symbols, pc); expErr == nil && err != nil {
			t.Errorf("Code: %#v == error  ==> %s", oper, err)
		} else if expErr != nil {
			switch wantErr := expErr.(type) {
			case *RegisterError:
				if !errors.As(err, &wantErr) {
					// 5 indents is 2 too many
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, err)
				}
				if wantErr.Reg != expErr.(*RegisterError).Reg {
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, expErr)
				}
			case *OffsetError:
				if !errors.As(err, &wantErr) {
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, err)
				}
				if wantErr.Offset != expErr.(*OffsetError).Offset {
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, expErr)
				}
			}
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

			if err != nil {
				t.Error(err)
			}
		}
	}

}

func TestLD_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{&LD{DR: "R0", OFFSET: 0x10}, 0x2010, nil},
		{&LD{DR: "R7", SYMBOL: "LABEL"}, 0x2e05, nil},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"FAR":    0x2f00,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

func TestADD_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{&ADD{DR: "R0", SR1: "R0", SR2: "R1"}, 0x1001, nil},
		{&ADD{DR: "R1", SR1: "R1", LITERAL: 0x000f}, 0x124f, nil},
		{&ADD{DR: "R1", SR1: "R1", SR2: "R0"}, 0x1240, nil},
		{&ADD{DR: "R0", SR1: "R7", LITERAL: 0b0000_0000_0000_1010}, 0b0001_0001_1100_1010, nil},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"FAR":    0x2f00,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

func TestNOT_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &NOT{DR: "R1", SR: "R1"}, want: 0x927f, wantErr: nil},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{}

	for tc := range tcs {
		op, exp := tcs[tc].oper, tcs[tc].want

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
