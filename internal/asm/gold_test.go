package asm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

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
	input       io.ReadCloser
	expected    io.ReadCloser
	expectedHex io.ReadCloser
}

func TestAssembler_Gold(tt *testing.T) {
	t := assemblerHarness{tt}

	tcs := []goldTestCase{
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
			parser := NewParser(t.logger())
			parser.Parse(tc.input)

			if parser.Err() != nil {
				t.Error(parser.Err())
			}

			syntax := parser.Syntax()
			symbols := parser.Symbols()

			generator := NewGenerator(symbols, syntax)

			var (
				out   bytes.Buffer
				count int64
				err   error
			)

			if tc.expectedHex == nil {
				count, err = generator.writeTo(&out)
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
		})
	}
}
