package monitor

import (
	"testing"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
)

type testHarness struct{ *testing.T }

// Tests the default system image:
//
//   - Traps, Exceptions, ISRs are present.
//   - Routines are relocated when loaded.
func TestWithSystemImage(tt *testing.T) {
	t := testHarness{tt}

	routine := Routine{
		Name:   "TestRoutine",
		Vector: 0x8000,
		Orig:   0x0400,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BLKW{ALLOC: 0xf},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0401,
		},
	}

	obj, err := Generate(routine)

	if err != nil {
		t.Error(err)
	}

	if obj.Orig != 0x0400 {
		t.Errorf("obj.Orig: want: %0#4x got: %v", 0x4000, obj.Orig)
	}

	t.Errorf("%+v", obj)

	image := NewSystemImage(log.DefaultLogger())
	image.Traps = []Routine{routine}

	t.Errorf("%+v", image)
}
