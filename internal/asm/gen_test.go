package asm_test

import (
	"bytes"
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

	expected := []byte{ // lil endian
		0x00, 0x30,
		0xff, 0x91,
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
