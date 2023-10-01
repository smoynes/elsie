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

	AddOperatorForTesting("TEST", &fakeInstruction{})

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

type fakeInstruction struct{}

func (fakeInstruction) String() string { return "TEST" }

func (fake *fakeInstruction) Parse(oper string, opers []string) (Instruction, error) {
	return fake, nil
}

const ValidSyntax = (`
; Let's go!
     .ORIG 0x1000       ; origin
START:;instructions
     ;; immediate mode
     TEST R1,#1      ; decimal ; 0x1000
     TEST R2,#0o2    ; octal   ; 0x1001
     TEST R3,#0xdada ; hex     ; 0x1002
     TEST                      ; 0x1003
     TEST R1,R2                ; 0x1004
     TEST R1,R2,R3             ; 0x1005
     TEST R6, R7, R0 ; spaces  ; 0x1006
     TEST R0,[R5]              ; 0x1007

END: TEST R0, LABEL            ; 0x1008
	TEST	TABS               ; 0x1009

; Label in the style of the text.
LOOP  TEST R3,R3,R2            ; 0x100a
      TEST R3,R3,#-1           ; 0x100b
      TEST LOOP                ; 0x100c

LOOP0 TEST                     ; 0x100d
LOOP1 TEST R1                  ; 0x100e
LOOP2 TEST R1,R2               ; 0x100f

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
eof:
.END
`)

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

	assertSymbol(t, symbols, "LOOP", 0x100a)
	assertSymbol(t, symbols, "LOOP0", 0x100d)
	assertSymbol(t, symbols, "LOOP1", 0x100e)
	assertSymbol(t, symbols, "LOOP2", 0x100f)

	assertSymbol(t, symbols, "LABEL", 0x3100)
	assertSymbol(t, symbols, "DECIMAL", 0x3100)
	assertSymbol(t, symbols, "HEX", 0x3101)
	assertSymbol(t, symbols, "OCTAL", 0x3102)
	assertSymbol(t, symbols, "BINARY", 0x3103)
	assertSymbol(t, symbols, "UNDER_SCORE", 0x3104)
	assertSymbol(t, symbols, "HYPHEN-ATE", 0x3104)
	assertSymbol(t, symbols, "D1G1T1", 0x3104)
	assertSymbol(t, symbols, "EOF", 0x3104)

	if len(symbols) != 15 {
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

func assertSymbol(t parserHarness, symbols SymbolTable, label string, want uint16) {
	t.Helper()

	if got, ok := symbols[label]; !ok {
		t.Errorf("symbol: %s, missing", label)
	} else if got != want {
		t.Errorf("symbol: %s, want: %0#4x, got: %0#4x", label, want, got)
	}
}