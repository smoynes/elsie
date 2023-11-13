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
	name, input, inputFilename string

	expectCode vm.ObjectCode
	expectErr  error
}

func TestHexEncoder_UnmarshalText(t *testing.T) {
	tcs := []unmarshalTestCase{
		{
			name:          "empty",
			input:         "",
			inputFilename: "",
			expectErr:     emptyErr,
		},
		{
			name:      "eof record",
			input:     ":000001FF",
			expectErr: emptyErr,
		},
		{
			name:      "eof record with newlines",
			input:     "\n\n:000001FF\n\n",
			expectErr: emptyErr,
		},
		{
			name:      "invalid bytes",
			input:     ":invalid",
			expectErr: invalidEncodingErr,
		},
		{
			name:      "nonsense",
			input:     "u wot mate",
			expectErr: invalidEncodingErr,
		},
		{
			name:      "data record",
			input:     ":000102cafe",
			expectErr: nil,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, err := unmarshal(tc)

			t.Logf("got: %v, err: %v", code, err)

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
				if code.Orig != tc.expectCode.Orig {
					t.FailNow()
				}
			}

		})
	}
}

func unmarshal(tc unmarshalTestCase) (vm.ObjectCode, error) {
	decoder := HexEncoding{}
	err := decoder.UnmarshalText([]byte(tc.input))
	return decoder.Code(), err
}
