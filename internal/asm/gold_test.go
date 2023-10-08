package asm_test

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
)

// gold_test.go contains so-called "golden tests": end-to-end tests that verify source-code input
// produces known machine-code output.

type assemblerHarness struct {
	*testing.T
}

func (t *assemblerHarness) inputStream(filename string) io.ReadCloser {
	t.Helper()

	file, err := os.Open(path.Join("testdata", filename))
	if err != nil {
		t.Fatalf("error opening %s: %s", filename, err)
	}

	return file
}

func (t *assemblerHarness) expectOutput(filename string) io.ReadCloser {
	t.Helper()

	file, err := os.Open(path.Join("testdata", filename))
	if err != nil {
		t.Fatalf("error opening %s: %s", filename, err)
	}

	return file
}

func (t *assemblerHarness) logger() *log.Logger {
	buf := bufio.NewWriter(os.Stderr)

	t.T.Cleanup(func() { buf.Flush() })

	return slog.New(
		slog.NewTextHandler(buf, log.Options),
	)
}

type goldTestCase struct {
	input    io.ReadCloser
	expected io.ReadCloser
}

func TestAssembler_Gold(tt *testing.T) {
	t := assemblerHarness{tt}

	tcs := []goldTestCase{
		{
			input:    t.inputStream("parser6.asm"),
			expected: t.expectOutput("parser6.out"),
		},
	}

	for _, tc := range tcs {
		tc := tc

		parser := asm.NewParser(t.logger())
		parser.Parse(tc.input)

		if parser.Err() != nil {
			t.Error(parser.Err())
		}

		syntax := parser.Syntax()
		symbols := parser.Symbols()

		out := new(bytes.Buffer)

		generator := asm.NewGenerator(symbols, syntax)
		count, err := generator.WriteTo(out)

		t.Logf("Wrote %d bytes", count)

		if err != nil {
			t.Error(err)
		}

		expect, err := io.ReadAll(tc.expected)
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(expect, out.Bytes()) != 0 {
			t.Error("bytes not equal")
		}
	}
}
