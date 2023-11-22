package encoding

import (
	"encoding"
	"errors"
	"testing"

	"github.com/smoynes/elsie/internal/vm"
)

// Assert interface implemented.
var (
	_ encoding.TextMarshaler   = (*HexEncoding)(nil)
	_ encoding.TextUnmarshaler = (*HexEncoding)(nil)
)

type unmarshalTestCase struct {
	name, input string

	expectCodes int
	expectErr   error
}

func TestHexEncoder_UnmarshalText(t *testing.T) {
	t.Parallel()

	tcs := []unmarshalTestCase{
		{
			name:      "empty",
			input:     "",
			expectErr: errEmpty,
		},
		{
			name:      "eof record",
			input:     ":00000001FF",
			expectErr: errEmpty,
		},
		{
			name:      "eof record with newlines",
			input:     "\n\n:000001FF\n\n",
			expectErr: errEmpty,
		},
		{
			name:      "invalid bytes",
			input:     ":invalid",
			expectErr: errInvalidHex,
		},
		{
			name:      "nonsense",
			input:     "u wot mate",
			expectErr: errInvalidHex,
		},
		{
			name:        "data record",
			input:       ":10246200464C5549442050524F46494C4500464C33\n",
			expectCodes: 1,
		},
		{
			name:        "another data record",
			input:       ":10001300AC12AD13AE10AF1112002F8E0E8F0F2244",
			expectCodes: 1,
		},
		{
			name:        "data records",
			input:       ":10246200464C5549442050524F46494C4500464C33\n:10246200464C5549442050524F46494C4500464C33\n",
			expectCodes: 2,
		},
		{
			// Our ISA is 16 bit
			name:      "odd length",
			input:     ":03020301FACE00",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":0",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":00",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF0",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF0",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF0000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF00000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF000000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF0000000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF00000000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF000000000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF0000000000",
			expectErr: errInvalidHex,
		},
		{
			name:      "too short",
			input:     ":FF00000000000",
			expectErr: errInvalidHex,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, err := unmarshal(tc)

			t.Logf("have: %q, got: %+v, err: %v", tc.input, code, err)

			switch {
			case tc.expectErr != nil && err != nil:
				if !errors.Is(err, tc.expectErr) {
					t.Errorf("Unexpected error: got: %s, want: %s",
						err.Error(), tc.expectErr.Error())
				}
			case tc.expectErr != nil && err == nil:
				t.Errorf("Expected error: %s", tc.expectErr.Error())
			case tc.expectErr == nil && err != nil:
				t.Errorf("Unexpected error: got: %v", err)
			case len(code) != tc.expectCodes:
				t.Errorf("Unexpected code: want: %d, got: %d", tc.expectCodes, len(code))
			default:
				for i := range code {
					if code[i].Orig == 0x0000 {
						t.Error("Origin not set: code:,", i)
					}
				}
			}
		})
	}
}

type marshalTestCase struct {
	name  string
	input []vm.ObjectCode

	expectOutput string
	expectErr    error
}

func TestHexEncoder_MarshalText(t *testing.T) {
	t.Parallel()

	tcs := []marshalTestCase{
		{
			name:         "nil",
			input:        nil,
			expectOutput: ":00000001ff\n",
		},
		{
			name: "fixed string",
			input: []vm.ObjectCode{
				{
					Orig: vm.Word(0x2462),
					Code: []vm.Word{
						0x464c, 0x5549, 0x4420, 0x5052, 0x4f46, 0x494c, 0x4500, 0x464c,
					},
				},
			},
			expectOutput: ":10246200464c5549442050524f46494c4500464c33\n:00000001ff\n",
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			output, err := marshal(tc)

			t.Logf("have: %+v, got: %q, err: %v", tc.input, output, err)

			switch {
			case tc.expectErr != nil && err != nil:
				if !errors.Is(err, tc.expectErr) {
					t.Errorf("Unexpected error: got: %s, want: %s",
						err.Error(), tc.expectErr.Error())
				}
			case tc.expectErr != nil && err == nil:
				t.Errorf("Expected error: %s", tc.expectErr.Error())
			case tc.expectErr == nil && err != nil:
				t.Errorf("Unexpected error: got: %v", err)
			default:
				if tc.expectOutput != output {
					t.Errorf("got: %q, want: %q", output, tc.expectOutput)
				}
			}

		})
	}
}

func marshal(tc marshalTestCase) (string, error) {
	encoder := HexEncoding{
		Code: tc.input,
	}
	out, err := encoder.MarshalText()

	return string(out), err
}

func unmarshal(tc unmarshalTestCase) ([]vm.ObjectCode, error) {
	decoder := HexEncoding{}
	err := decoder.UnmarshalText([]byte(tc.input))

	return decoder.Code, err
}
