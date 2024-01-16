package monitor

import (
	"testing"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

type testHarness struct{ *testing.T }

// Tests the default system image:
//
//   - Traps, Exceptions, ISRs are present.
//   - Routines are relocated when loaded.
func TestWithSystemImage(tt *testing.T) {
	t := testHarness{tt}

	if testing.Verbose() {
		log.LogLevel.Set(log.Debug)
	} else {
		log.LogLevel.Set(log.Warn)
	}

	image := NewSystemImage(log.DefaultLogger())

	routine := Routine{
		Name:   "TestRoutine",
		Vector: 0x8000,
		Orig:   0x0400,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BLKW{ALLOC: 0x2},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0401,
		},
	}

	obj, err := GenerateRoutine(routine)

	if err != nil {
		t.Error(err)
	}

	if obj.Orig != 0x0400 {
		t.Errorf("obj.Orig: want: %0#4x got: %v", 0x4000, obj.Orig)
	}

	t.Errorf("%+v", obj)

	routine = Routine{
		Name:   "TestTrap",
		Vector: 0x0100,
		Orig:   0x0400,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.BLKW{ALLOC: 0x2},
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0401,
		},
	}
	image.Traps = []Routine{routine}

	routine = Routine{
		Name:   "TestException",
		Vector: 0x0101,
		Orig:   0x0500,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0400,
		},
	}
	image.Exceptions = []Routine{routine}

	routine = Routine{
		Name:   "TestInterrupt",
		Vector: 0x0102,
		Orig:   0x0600,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x1000,
		},
	}
	image.ISRs = []Routine{routine}

	t.Errorf("%+v", image)

	machine := vm.New()
	loader := vm.NewLoader(machine)
	err = loadImage(loader, image)
	view := machine.Mem.View()

	t.Logf("%+v", view[0x0600:0x0601])

	for _, tc := range []struct {
		addr, want vm.Word
	}{
		{0x0400, 0x0e01},
		{0x0401, 0x0e00},
		{0x0402, 0x0fff},
		{0x0405, 0x0ffc},
		{0x0500, 0x0f00},
		{0x0600, 0x0fff}, // overflow
	} {
		want := vm.NewInstruction(vm.BR, uint16(tc.want))
		got := view[tc.addr]

		if vm.Word(want) != got {
			t.Errorf("view[%v]: want: %v != got: %v", tc.addr, tc.want, got)
		}
	}
}
