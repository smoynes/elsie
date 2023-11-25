package asm_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

type asmTestCase struct {
	input         io.Reader
	inputBytes    []byte
	expected      io.Reader
	expectedHex   io.Reader
	expectedSlice []byte
	expectedErr   error
}

func (tc *asmTestCase) Run(t *assemblerHarness) {
	t.Helper()

	parser := asm.NewParser(t.logger())

	if tc.input != nil {
		parser.Parse(tc.input)
	} else {
		parser.Parse(bytes.NewReader(tc.inputBytes))
	}

	if parser.Err() != nil {
		t.Error(parser.Err())
	}

	syntax := parser.Syntax()
	symbols := parser.Symbols()
	generator := asm.NewGenerator(symbols, syntax)

	var (
		out   bytes.Buffer
		count int64
		err   error
	)

	if tc.expectedHex == nil {
		count, err = generator.WriteTo(&out)
	} else {
		bs, err := generator.Encode()
		if err != nil {
			t.Error(err.Error())
			return
		}
		c, err := out.Write(bs)
		if err != nil {
			t.Error(err.Error())
			return
		}
		count = int64(c)
	}

	t.Logf("Wrote %d bytes", count)

	if err != nil {
		t.Error(err)
	}

	var expected []byte

	if tc.expected != nil {
		expected, err = io.ReadAll(tc.expected)
		if err != nil {
			t.Error(err)
			return
		}
	} else if tc.expectedHex != nil {
		expected, err = io.ReadAll(tc.expectedHex)
		if err != nil {
			t.Error(nil)
			return
		}
	} else if tc.expectedSlice != nil {
		expected = tc.expectedSlice
	}

	if bytes.Compare(expected, out.Bytes()) != 0 {
		t.Error("bytes not equal:")

		b := out.Bytes()

		for i := 0; i < len(b) && i < len(expected); i++ {
			if b[i] != expected[i] {
				t.Errorf("\tindex: %d: %0#2x != %0#2x (%[2]q != %[3]q)", i, b[i], expected[i])
			}
		}
	}

	if tc.expectedErr != nil {
		if !errors.Is(err, tc.expectedErr) {
			t.Errorf("expected err: %[1]s (%+[1]v), got: %[2]s (%+[2]v)", tc.expectedErr, err)
		}
	}
}

func TestAssembler_Gold(tt *testing.T) {
	t := assemblerHarness{tt}

	tcs := []asmTestCase{
		{
			input:    t.inputStream("parser6.asm"),
			expected: t.expectOutput("parser6.out"),
		},
		{
			input:    t.inputStream("parser7.asm"),
			expected: t.expectOutput("parser7.out"),
		},
		{
			input:       t.inputStream("parser6.asm"),
			expectedHex: t.expectOutput("parser6.hex"),
		},
		{
			input:       t.inputStream("parser7.asm"),
			expectedHex: t.expectOutput("parser7.hex"),
		},
		{
			input:       t.inputStream("parser9.asm"),
			expectedHex: t.expectOutput("parser9.hex"),
		},
	}

	for i, tc := range tcs {
		tc := tc

		var name string
		if fn, ok := tc.expected.(interface{ Name() string }); ok {
			name = fmt.Sprintf("Test %s #%d", fn.Name(), i)
		} else if fn, ok := tc.expectedHex.(interface{ Name() string }); ok {
			name = fmt.Sprintf("Test %s #%d", fn.Name(), i)
		} else {
			name = fmt.Sprintf("Test #%d", i)
		}

		t.Run(name, func(tt *testing.T) {
			t := assemblerHarness{tt}
			tc.Run(&t)
		})
	}
}

func TestAssembler_EdgeCases(tt *testing.T) {
	tcs := map[string]asmTestCase{
		"nil": {
			input:         nil,
			expectedSlice: nil,
			expectedErr:   nil,
		},
	}

	for name, tc := range tcs {
		tc := tc

		tt.Run(name, func(tt *testing.T) {
			t := assemblerHarness{tt}
			tc.Run(&t)
		})
	}
}
