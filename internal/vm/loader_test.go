package vm_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/smoynes/elsie/internal/log"
	. "github.com/smoynes/elsie/internal/vm"
)

type loaderHarness struct {
	*testing.T
}

func (*loaderHarness) Logger() *log.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

type loaderCase struct {
	name         string
	origin       Word
	instructions []Word
	expLoaded    uint16
	expErr       error
}

func TestLoader(tt *testing.T) {
	tt.Parallel()

	tcs := []loaderCase{{
		name:   "Ok",
		origin: 0x3100,
		instructions: []Word{
			Word(NewInstruction(LEA, 0o73)),
			Word(NewInstruction(TRAP, 0x25)),
			Word(NewInstruction(STI, 0xdad)),
		},
		expLoaded: 3,
	}, {
		name:         "too short",
		instructions: []Word{},
		expErr:       ErrObjectLoader,
	},
	}

	for _, tc := range tcs {
		tc := tc

		tt.Run(tc.name, func(tt *testing.T) {
			t := loaderHarness{tt}
			t.Parallel()

			loader := NewLoader()
			machine := New(WithLogger(t.Logger()))

			obj := ObjectCode{Orig: tc.origin, Code: tc.instructions}
			loaded, err := loader.Load(machine, obj)

			if loaded != tc.expLoaded {
				t.Errorf("Wrong loaded count: got: %d != want: %d", loaded, tc.expLoaded)
			}

			switch {
			case tc.expErr == nil && err != nil:
				t.Error("unexpected error ", err)
			case tc.expErr != nil && err == nil:
				t.Error("expected error:", "want:", tc.expErr, "got:", err)
			case !errors.Is(err, tc.expErr):
				t.Error("unexpected error:", "want", tc.expErr, "got", err)
			}

			if loaded == 0 && err == nil {
				t.Error("none loaded")
			}
		})
	}
}

type objectCase struct {
	name      string
	bytes     []byte
	expObject ObjectCode
	expRead   int
	expErr    error
}

func TestObjectCode(t *testing.T) {
	t.Parallel()

	tcs := []objectCase{{
		name: "Ok",
		bytes: []byte{
			0x40, 0x00,
			0x12, 0x34,
			0x56, 0x78,
		},
		expRead: 6,
		expObject: ObjectCode{
			Orig: Word(0x4000),
			Code: []Word{
				Word(Instruction(0x1234).Encode()),
				Word(Instruction(0x5678).Encode()),
			},
		},
	}, {
		name:   "too short",
		bytes:  nil,
		expErr: ErrObjectLoader,
	}}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(tt *testing.T) {
			t := loaderHarness{tt}
			t.Parallel()

			obj := ObjectCode{}
			read, err := obj.Read(tc.bytes)

			if read != tc.expRead {
				t.Error("unexpected read count", "want:", tc.expRead, "got:", read)
			}

			if obj.Orig != tc.expObject.Orig {
				t.Error("unexpected origin", "want:", tc.expObject.Orig, "got:", obj.Orig)
			}

			switch {
			case tc.expErr == nil && err != nil:
				t.Error("unexpected error:", "got:", err)
			case tc.expErr != nil && err == nil:
				t.Error("expected error:", "want:", tc.expErr, "got:", err)
			case tc.expErr != nil && err != nil:
				if !errors.Is(err, tc.expErr) {
					t.Error("unexpected error:", "want", tc.expErr, "got", err)
				}
			}

			if len(obj.Code) != len(tc.expObject.Code) {
				t.Error("code length", "want:", len(tc.expObject.Code), "got:", len(obj.Code))
			}

			for i := range obj.Code {
				if obj.Code[i] != tc.expObject.Code[i] {
					t.Errorf("unexpected code: want: %0#4x, got: %0#4x",
						tc.expObject.Code[i], obj.Code[i])
				}
			}
		})
	}
}
