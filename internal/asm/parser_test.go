package asm_test

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"
	"testing/iotest"

	. "github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
)

func init() {
	log.LogLevel.Set(log.Debug)
}

// parserHarness holds the test state and provides helpers.
type parserHarness struct {
	*testing.T
}

func (h *parserHarness) logger() *log.Logger {
	buf := bufio.NewWriter(os.Stderr)

	h.T.Cleanup(func() { buf.Flush() })

	return slog.New(
		slog.NewTextHandler(buf, log.Options),
	)
}

// Parser is a factory method for a parser under test. It creates a new parser and does an initial
// read on the input stream and returns the parser. The caller should assert against Parser.Err, etc.
func (h *parserHarness) ParseStream(in io.ReadCloser) *Parser {
	h.T.Helper()

	AddOperatorForTesting("TEST", &fakeOper{})

	if parser := NewParser(h.logger()); parser == nil {
		h.T.Fatal("parser: nil")
		return parser
	} else {
		parser.Parse(in)
		return parser
	}
}

func (parserHarness) inputString(in string) io.ReadCloser {
	reader := bytes.NewReader([]byte(in))
	return io.NopCloser(reader)
}

func (h parserHarness) inputFixture(in string) io.ReadCloser {
	reader, err := os.Open(path.Join("testdata", in))
	if err != nil {
		h.Errorf("fixture: %s", err)
	}
	return reader
}

func (parserHarness) inputError() io.ReadCloser {
	return io.NopCloser(iotest.ErrReader(os.ErrInvalid))
}

type fakeOper struct{}

var _ Instruction = (*fakeOper)(nil)

func (fakeOper) String() string { return "TEST" }

func (fake *fakeOper) Parse(oper string, opers []string) (Instruction, error) {
	return &(*fake), nil // Too clever copy.
}

const ValidSyntax = (`
; Let's go!

 .ORIG 0x1000       ; origin

START:;instructions
     ;; immediate mode
     TEST R1,#1      ; decimal
     TEST R2,#0o2    ; octal
     TEST R3,#0xdada ; hex
     TEST
     TEST R1,R2
     TEST R1,R2,R3
     TEST R6, R7, R0    ; spaces

     TEST R0,[R5]
END: TEST R0, LABEL

	TEST	TABS

LOOP:TEST R3,R3,R2
     TEST R3,R3,#-1
     TEST LOOP

     .ORIG 0x3100

LABEL:
decimal:
    .DW #0
hex:
    .DW x0001
octal:
    .DW o002
binary:
    .DW b0000_0000_0000_0111
under_score:
hyphen-ate:
d1g1t1:
eof:`)

func TestParser(tt *testing.T) {
	t := parserHarness{tt}

	parser := t.ParseStream(
		io.NopCloser(strings.NewReader(ValidSyntax)),
	)

	if err := parser.Err(); err != nil {
		t.Fatal(err)
	}

	symbols := parser.Symbols()

	if len(symbols) == 0 {
		t.Fatal("no symbols")
	}

	assertSymbol(t, symbols, "START", 0x1000)
	assertSymbol(t, symbols, "END", 0x1008)
	assertSymbol(t, symbols, "LOOP", 0x1009)
	assertSymbol(t, symbols, "LABEL", 0x3100)
	assertSymbol(t, symbols, "decimal", 0x3100)
	assertSymbol(t, symbols, "hex", 0x3101)
	assertSymbol(t, symbols, "octal", 0x3102)
	assertSymbol(t, symbols, "binary", 0x3103)
	assertSymbol(t, symbols, "under_score", 0x3104)
	assertSymbol(t, symbols, "hyphen-ate", 0x3104)
	assertSymbol(t, symbols, "d1g1t1", 0x3104)
	assertSymbol(t, symbols, "eof", 0x3104)

	if len(symbols) != 12 {
		t.Errorf("unexpected symbols: want: %d, got: %d", 11, len(symbols))
		t.Log("Symbol table:")

		for k := range symbols {
			t.Log(" ", k)
		}
	}

	instructions := parser.Instructions()
	if len(instructions) == 0 {
		t.Fatal("no instructions")
	}
}

func assertSymbol(t parserHarness, symbols SymbolTable, label string, want int) {
	t.Helper()

	if got, ok := symbols[label]; !ok {
		t.Errorf("symbol: %s, missing", label)
	} else if got != want {
		t.Errorf("symbol: %s, want: %0#4x, got: %0#4x", label, want, got)
	}
}
