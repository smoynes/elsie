package asm_test

import (
	"bufio"
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
// read on the input stream and returns the parser. The caller should assert against Parser.Err,
// etc.
func (h *parserHarness) ParseStream(in io.ReadCloser) *Parser {
	h.T.Helper()

	parser := NewParser(h.logger())

	if parser == nil {
		h.T.Fatal("parser: nil")
		return parser
	}

	parser.Probe("TEST", &fakeInstruction{})
	parser.Parse(in)

	return parser
}

func (h parserHarness) inputString(in string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(in))
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

func (fake *fakeInstruction) Parse(oper string, opers []string) error {
	return nil
}

func (fake *fakeInstruction) Generate(sym SymbolTable, loc uint16) ([]uint16, error) {
	return nil, nil
}

const ValidSyntax = (`
; Let's go!
     .ORIG x1000       ; origin
START:;instructions
     ;; immediate mode
     TEST R1,#1      ; decimal ; 0x1000
     TEST R2,#o2     ; octal   ; 0x1001
     TEST R3,#xdada  ; hex     ; 0x1002
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
    .DW 123                    ; 0x3100
hex:
    .DW x0001                  ; 0x3101
octal:
    .DW o002                   ; 0x3102
binary:
    .DW b0000_0000_0000_0111   ; 0x3103
under_score:
hyphen-ate:
d1g1t1:


.ORIG x3200
AND R1,R1,#0
AND R1,R2,#x0
AND R1,R2,LOOP

BRNP  LOOP
BRz   #x0

  LD  DR,#x0
  LD  DR,#o777
  LD  R1, LOOP
  LD  R1,[LOOP] ; ???
  LD  DR,#-1
  LD  DR,#012
  LD  R9,#x0123

eof:
.END
`)

// A rough integration test for the parser. It is quite brittle and, yet, has proven valuable during
// design and development.
func TestParser(tt *testing.T) {
	t := parserHarness{tt}

	in := t.inputString(ValidSyntax)
	parser := t.ParseStream(in)

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
	assertSymbol(t, symbols, "EOF", 0x320c)

	if len(symbols) != 15 {
		t.Errorf("unexpected symbols: want: %d, got: %d", 11, len(symbols))
		t.Log("Symbol table:")

		for k := range symbols {
			t.Log(" ", k)
		}
	}

	err := parser.Err()
	if err != nil {
		t.Error(err)
	}

	instructions := parser.Syntax()
	if len(instructions) == 0 {
		t.Error("no instructions")
	}
}

// Test the parser using source code from ./testdata.
func TestParser_Fixtures(tt *testing.T) {
	tests := []string{
		//"parser2.asm", TODO: .STRINGZ
		"parser3.asm",
		"parser4.asm",
		"parser5.asm",
		"parser6.asm",
	}

	for _, fn := range tests {
		fn := fn

		tt.Run(fn, func(tt *testing.T) {
			var (
				t  = parserHarness{tt}
				fs = t.inputFixture(fn)
			)

			parser := t.ParseStream(fs)

			if parser.Err() != nil {
				t.Error(parser.Err())
			}

			t.Logf("%#v", parser.Symbols())
		})
	}
}

func TestParser_Errors(tt *testing.T) {
	t := parserHarness{tt}

	parser := t.ParseStream(t.inputError())

	err := parser.Err()

	if err == nil {
		t.Error("expected error")
	}
}

func TestParser_FILL(tt *testing.T) {
	t := parserHarness{tt}
	in := t.inputString(`
.ORIG x1234
.FILL xdada
`)
	parser := t.ParseStream(in)

	if err := parser.Err(); err != nil {
		t.Error(err)
	}

	syntax := parser.Syntax()

	if len(syntax) == 0 {
		t.Errorf("no data")
	}

	code := syntax[0x1234]

	if fill, ok := code.(*FILL); !ok || fill.LITERAL != 0xdada {
		t.Errorf("data: 0x1234 %#v != %#v", syntax[0x1234], 0xdada)
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
