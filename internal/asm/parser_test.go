package asm_test

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"
	"testing"

	. "github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
)

func init() {
	log.LogLevel.Set(log.Debug)
}

// harness holds the test state and provides helpers.
type harness struct {
	*testing.T
}

func (*harness) logger() *log.Logger {
	return slog.New(
		slog.NewTextHandler(
			bufio.NewWriter(os.Stdout), log.Options, // ⍟
		),
	)
}

// Parser is a factory method for a parser under test. It creates a new parser and does an initial
// read on the input stream and returns the parser.
func (h *harness) Parse(in io.ReadCloser) Parser {
	if parser := NewParser(h.logger()); parser == nil {
		h.T.Fatal("parser: nil")
		return parser
	} else {
		parser.Read(in)
		return parser
	}
}

func (h *harness) inputString(in string) io.ReadCloser {
	reader := bytes.NewReader([]byte(in))
	return io.NopCloser(reader)
}

func (h *harness) inputError() io.ReadCloser {
	return io.NopCloser(&errorReader{})
}

func TestTokenizer(tt *testing.T) {
	t := &harness{tt}

	var emptySyms = map[string]struct {
		in io.ReadCloser
	}{
		"empty": {in: t.inputString("")},
		"error": {in: t.inputError()},
	}

	for n, tc := range emptySyms {
		t.Run(n, func(tt *testing.T) {
			t := &harness{tt}

			parser := t.Parse(tc.in)
			err := parser.Err()

			if err != nil {
				t.Errorf("unexpected parse error: got %v", err)
			}

			syms := parser.Symbols()
			if syms == nil {
				t.Error("symbol table: nil")
			} else if len(syms) != 0 {
				t.Error("len(symbols) != 0:", len(syms))
			}

		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrNoProgress
}
