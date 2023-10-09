package asm_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"unicode/utf16"

	. "github.com/smoynes/elsie/internal/asm"
)

type generatorHarness struct {
	*testing.T
}

type generateCase struct {
	oper     Operation
	want     uint16   // A single code point.
	wantCode []uint16 // Multiple code points.
	wantErr  error
}

// Run tests a collection of generator tests cases.
func (t *generatorHarness) Run(pc uint16, symbols SymbolTable, tcs []generateCase) {
	t.Helper()

	for i := range tcs {
		oper, want, expErr := tcs[i].oper, tcs[i].want, tcs[i].wantErr

		if mc, err := oper.Generate(symbols, pc); expErr == nil && err != nil {
			t.Errorf("unexpected error: %#v %s", oper, err)
		} else if expErr != nil {
			switch wantErr := expErr.(type) { //nolint:errorlint
			case *RegisterError:
				if !errors.As(err, &wantErr) {
					// 5 indents is 2 too many
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, err)
				}
				if wantErr.Reg != expErr.(*RegisterError).Reg { //nolint:errorlint
					t.Errorf("unexpected error: want: %#v, got: %#v", wantErr, expErr)
				}
			case *OffsetError:
				if !errors.As(err, &wantErr) {
					t.Errorf("unexpected error: want: %#v, got: %#v", expErr, err)
				}
				if wantErr.Offset != expErr.(*OffsetError).Offset { //nolint:errorlint
					t.Errorf("unexpected error: want: %#v, got: %#v", expErr, err)
				}
			case *SymbolError:
				if !errors.As(err, &wantErr) {
					t.Errorf("unexpected error: want: %#v, got: %#v", expErr, err)
				}
				if wantErr.Symbol != expErr.(*SymbolError).Symbol { //nolint:errorlint
					t.Errorf("unexpected error: want: %#v, got: %#v", expErr, err)
				}
			default:
				t.Errorf("expected error: want: %#v, got: %#v", expErr, err)
			}
		} else {
			t.Logf("Code: %#v == generated ==> %0#4x", oper, mc)

			if mc == nil {
				t.Error("invalid machine code")
			}

			if len(mc) != 1 {
				t.Errorf("incorrect machine code: %d bytes", len(mc))
				return
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

// TestGenerator is something like an integration test for the code generator.
func TestGenerator(tt *testing.T) {
	t := generatorHarness{tt}

	var buf bytes.Buffer

	syntax := make(SyntaxTable, 0)
	syntax.Add(&ORIG{LITERAL: 0x3000})
	syntax.Add(&NOT{DR: "R0", SR: "R7"})
	syntax.Add(&AND{DR: "R3", SR1: "R4", SR2: "R6"})

	symbols := SymbolTable{}
	symbols.Add("LABEL", 0x2ff0)

	gen := NewGenerator(symbols, syntax)

	count, err := gen.WriteTo(&buf)
	if err != nil {
		t.Error(err)
	}

	expected := []byte{ // big endian
		0x30, 0x00,
		0x91, 0xff,
		0x57, 0x06,
	}

	bytes := buf.Bytes()

	if want, got := len(expected), count; want != int(got) {
		t.Errorf("expected: %d, got: %d bytes", want, got)
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
		{oper: &BR{NZP: 0x7, OFFSET: 0x01}, want: 0x0e01, wantErr: nil},
		{oper: &BR{NZP: 0x2, OFFSET: 0xfff0}, want: 0x05f0, wantErr: nil},
		{oper: &BR{NZP: 0x3, SYMBOL: "LABEL"}, want: 0x0605, wantErr: nil},
		{oper: &BR{NZP: 0x3, SYMBOL: "BACK"}, want: 0x0600, wantErr: nil},
		{oper: &BR{NZP: 0x4, SYMBOL: "LONG"}, want: 0x061f, wantErr: &OffsetError{Offset: 0xd000}},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"LONG":   0x0,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

func TestLDR_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &LDR{DR: "R0", SR: "R5", OFFSET: 0x10}, want: 0x6150, wantErr: nil},
		{oper: &LDR{DR: "R7", SR: "R7", OFFSET: 0xffff}, want: 0x6fff, wantErr: nil},
		{oper: &LDR{DR: "R7", SR: "R4", SYMBOL: "LABEL"}, want: 0x6f05, wantErr: nil},
		{oper: &LDR{DR: "R5", SR: "R1", SYMBOL: "BACK"}, want: 0x6a40, wantErr: nil},
		{oper: &LDR{DR: "R3", SR: "R2", SYMBOL: "GONE"}, want: 0, wantErr: &SymbolError{0x3000, "GONE"}},
		{oper: &LDR{DR: "R1", SR: "R3", SYMBOL: "FAR"}, want: 0, wantErr: &OffsetError{Offset: 0xbf00}},
		{oper: &LDR{DR: "R2", SR: "R4", SYMBOL: "YONDER"}, want: 0, wantErr: &OffsetError{Offset: 0x1000}},
		{oper: &LDR{DR: "R8", SR: "R2", SYMBOL: "LABEL"}, want: 0, wantErr: &RegisterError{Reg: "R8"}},
		{oper: &LDR{DR: "R0", SR: "DR", SYMBOL: "LABEL"}, want: 0, wantErr: &RegisterError{Reg: "DR"}},
	}
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":  0x3005,
		"BACK":   0x3000,
		"FAR":    0xef00,
		"YONDER": 0x4000,
	}

	t.Run(pc, symbols, tcs)
}

func TestLD_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &LD{DR: "R0", OFFSET: 0x2f}, want: 0x202f},
		{oper: &LD{DR: "R0", OFFSET: 0xffff}, want: 0x20ff},
		{oper: &LD{DR: "R0", OFFSET: 0x10}, want: 0x2010, wantErr: nil},
		{oper: &LD{DR: "R7", SYMBOL: "LABEL"}, want: 0x2e05, wantErr: nil},
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

func TestLEA_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff,
		"THERE":     0x3080,
		"WAYBACK":   0x2c00,
		"OVERTHERE": 0x3200,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &LEA{DR: "R0", OFFSET: 0x003f}, want: 0xe03f},
		{oper: &LEA{DR: "R1", OFFSET: 0x01ff}, want: 0xe3ff},
		{oper: &LEA{DR: "R2", SYMBOL: "THERE"}, want: 0xe480},
		{oper: &LEA{DR: "R3", SYMBOL: "WAYBACK"}, wantErr: &OffsetError{0xfc00}},
		{oper: &LEA{DR: "R4", SYMBOL: "OVERTHERE"}, wantErr: &OffsetError{0x0200}},
		{oper: &LEA{DR: "R5", SYMBOL: "DNE"}, wantErr: &SymbolError{Loc: 0x3000, Symbol: "DNE"}},
	}

	t.Run(pc, symbols, tcs)
}

func TestLDI_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x31ff,
		"THERE":     0x3080,
		"WAYBACK":   0x2dff,
		"OVERTHERE": 0x3200,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &LDI{SR: "R0", OFFSET: 0x10}, want: 0xa010, wantErr: nil},
		{oper: &LDI{SR: "R0", OFFSET: 0xffff}, want: 0xa1ff, wantErr: nil},
		{oper: &LDI{SR: "R7", SYMBOL: "LABEL"}, want: 0xafff, wantErr: nil},
		{oper: &LDI{SR: "R2", SYMBOL: "THERE"}, want: 0xa480},
		{oper: &LDI{SR: "R3", SYMBOL: "WAYBACK"}, wantErr: &OffsetError{0xfdff}},
		{oper: &LDI{SR: "R4", SYMBOL: "OVERTHERE"}, wantErr: &OffsetError{0x0200}},
		{oper: &LDI{SR: "R5", SYMBOL: "DNE"}, wantErr: &SymbolError{Loc: 0x3000, Symbol: "DNE"}},
	}

	t.Run(pc, symbols, tcs)
}

func TestST_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff,
		"THERE":     0x3080,
		"WAYBACK":   0x2c00,
		"OVERTHERE": 0x3200,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &ST{SR: "R0", OFFSET: 0x1ff}, want: 0x31ff},
		{oper: &ST{SR: "R0", OFFSET: 0x00}, want: 0x3000},
		{oper: &ST{SR: "R0", OFFSET: 0xffff}, want: 0x31ff},
		{oper: &ST{SR: "R0", OFFSET: 0x10}, want: 0x3010, wantErr: nil},
		{oper: &ST{SR: "R7", SYMBOL: "LABEL"}, want: 0x3fff, wantErr: nil},
		{oper: &ST{SR: "R2", SYMBOL: "THERE"}, want: 0x3480},
		{oper: &ST{SR: "R3", SYMBOL: "WAYBACK"}, wantErr: &OffsetError{0xfc00}},
		{oper: &ST{SR: "R4", SYMBOL: "OVERTHERE"}, wantErr: &OffsetError{0x0200}},
		{oper: &ST{SR: "R5", SYMBOL: "DNE"}, wantErr: &SymbolError{Loc: 0x3000, Symbol: "DNE"}},
	}

	t.Run(pc, symbols, tcs)
}

func TestSTI_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff,
		"THERE":     0x3080,
		"WAYBACK":   0x2c00,
		"OVERTHERE": 0x3200,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &STI{SR: "R0", OFFSET: 0x10}, want: 0xb010, wantErr: nil},
		{oper: &STI{SR: "R7", SYMBOL: "LABEL"}, want: 0xbfff, wantErr: nil},
		{oper: &STI{SR: "R2", SYMBOL: "THERE"}, want: 0xb480},
		{oper: &STI{SR: "R3", SYMBOL: "WAYBACK"}, wantErr: &OffsetError{0xfc00}},
		{oper: &STI{SR: "R4", SYMBOL: "OVERTHERE"}, wantErr: &OffsetError{0x0200}},
		{oper: &STI{SR: "R5", SYMBOL: "DNE"}, wantErr: &SymbolError{Loc: 0x3000, Symbol: "DNE"}},
	}

	t.Run(pc, symbols, tcs)
}

func TestSTR_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff, // -1
		"THERE":     0x301f, // 64
		"BACK":      0x2fe0, // -64
		"WAYBACK":   0x2fd0,
		"OVERTHERE": 0x3040,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &STR{SR1: "R0", SR2: "R1", OFFSET: 0x2f}, want: 0x706f},
		{oper: &STR{SR1: "R0", SR2: "R0", OFFSET: 0x00}, want: 0x7000},
		{oper: &STR{SR1: "R0", SR2: "R1", OFFSET: 0xffff}, want: 0x707f},
		{oper: &STR{SR1: "R7", SR2: "R2", SYMBOL: "LABEL"}, want: 0x7e9f},
		{oper: &STR{SR1: "R2", SR2: "R3", SYMBOL: "THERE"}, want: 0x74df},
		{oper: &STR{SR1: "R3", SR2: "R4", SYMBOL: "WAYBACK"}, wantErr: &OffsetError{0xffd0}},
		{oper: &STR{SR1: "R4", SR2: "R5", SYMBOL: "OVERTHERE"}, wantErr: &OffsetError{0x0040}},
		{oper: &STR{SR1: "R5", SR2: "R6", SYMBOL: "DNE"}, wantErr: &SymbolError{Loc: 0x3000, Symbol: "DNE"}},
	}

	t.Run(pc, symbols, tcs)
}

func TestJMP_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff, // -1
		"THERE":     0x301f, // 64
		"BACK":      0x2fe0, // -64
		"WAYBACK":   0x2fd0,
		"OVERTHERE": 0x3040,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &JMP{SR: "R0"}, want: 0xc000},
		{oper: &JMP{SR: "R7"}, want: 0xc1c0},
		{oper: &JMP{SR: "R2"}, want: 0xc080},
	}

	t.Run(pc, symbols, tcs)
}

func TestRET_Generate(tt *testing.T) {
	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL":     0x2fff, // -1
		"THERE":     0x301f, // 64
		"BACK":      0x2fe0, // -64
		"WAYBACK":   0x2fd0,
		"OVERTHERE": 0x3040,
	}

	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &RET{}, want: 0xc1c0},
	}

	t.Run(pc, symbols, tcs)
}
func TestADD_Generate(tt *testing.T) {
	t := generatorHarness{tt}
	tcs := []generateCase{
		{oper: &ADD{DR: "R0", SR1: "R0", SR2: "RR"}, want: 0, wantErr: &RegisterError{Reg: "RR"}},
		{oper: &ADD{DR: "R4", SR1: "R1", LITERAL: ^uint16(0x0004)}, want: 0x187b, wantErr: nil},
		{oper: &ADD{DR: "R1", SR1: "R1", LITERAL: 0x000f}, want: 0x126f, wantErr: nil},
		{oper: &ADD{DR: "R1", SR1: "R1", SR2: "R0"}, want: 0x1240, wantErr: nil},
		{oper: &ADD{DR: "R0", SR1: "R7", LITERAL: 0b0000_0000_0000_1010}, want: 0b0001_0001_1110_1010, wantErr: nil},
		{oper: &ADD{DR: "R2", SR1: "R6", LITERAL: 0x15cf}, want: 0x15af, wantErr: nil},
		{oper: &ADD{DR: "R1", SR1: "R1", LITERAL: 0x21c0}, want: 0x1260, wantErr: nil},
		{oper: &ADD{DR: "R1", SR1: "R1", LITERAL: 0}, want: 0x1260, wantErr: nil},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{}

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

func TestSTRINGZ_Generate(tt *testing.T) {
	t := generatorHarness{tt}

	tcs := []generateCase{
		{
			oper:     &STRINGZ{LITERAL: "Hello, there!"},
			wantCode: utf16.Encode([]rune("Hello, there!\x00")),
		},
		{
			oper:     &STRINGZ{LITERAL: ""},
			wantCode: []uint16{0x0000},
		},
		{
			oper:     &STRINGZ{LITERAL: "⍤"},
			wantCode: append(utf16.Encode([]rune{'⍤'}), 0x0000),
		},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{}

	for tc := range tcs {
		op := tcs[tc].oper
		wantCode := tcs[tc].wantCode

		code, err := op.Generate(symbols, pc)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if code == nil {
			t.Error("invalid machine code")
		}

		// Convert []uint16 to []byte...
		codeBuffer := new(bytes.Buffer)
		err = binary.Write(codeBuffer, binary.BigEndian, code)

		if err != nil {
			t.Error(err)
			return
		}

		wantBytes := new(bytes.Buffer)
		err = binary.Write(wantBytes, binary.BigEndian, wantCode)

		if err != nil {
			t.Error(err)
			return
		}

		if bytes.Compare(codeBuffer.Bytes(), wantBytes.Bytes()) != 0 {
			t.Error("code differs")
			t.Errorf("%s", codeBuffer.Bytes())
			t.Errorf("%s", wantBytes.Bytes())
		}
	}
}
