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
			/* 0x0400 */
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			/* 0x0401 */
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			/* 0x0402 */
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			/* 0x0403 */
			&asm.BLKW{ALLOC: 0x2},
			/* 0x0405 */
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0402,
		},
	}

	obj, err := GenerateRoutine(routine)

	if err != nil {
		t.Error(err)
	}

	for i, exp := range []vm.Word{0x0e01, 0x0e00, 0x0fff, 0x2361, 0x2361, 0x0ffc} {
		if got := obj.Code[i]; got != exp {
			t.Errorf("obj.Orig[%d]: want: %v got: %v", i, exp, got)
		}
	}

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
			&asm.BR{NZP: asm.CondNP, SYMBOL: "LABEL"},
			&asm.BR{NZP: asm.CondNP, SYMBOL: "LABEL2"},
		},
		Symbols: asm.SymbolTable{
			"LABEL":  0x0400,
			"LABEL2": 0x0501,
		},
	}
	image.Exceptions = []Routine{routine}

	routine = Routine{
		Name:   "TestInterrupt",
		Vector: 0x0102,
		Orig:   0x0600,
		Code: []asm.Operation{
			&asm.BR{NZP: asm.CondNZP, SYMBOL: "LABEL"},
			&asm.LD{DR: "R7", SYMBOL: "LABEL2"},
			&asm.BR{NZP: 0, SYMBOL: "LABEL3"},
		},
		Symbols: asm.SymbolTable{
			"LABEL": 0x0700,
			"LABEL2": 0x0700,
			"LABEL3": 0x0502,
		},
	}
	image.ISRs = []Routine{routine}

	machine := vm.New()
	loader := vm.NewLoader(machine)

	err = loadImage(loader, image)

	if err != nil {
		t.Errorf("load failed: %v", err)
		return
	}

	view := machine.Mem.View()

	for _, tc := range []struct {
		addr, want vm.Word
	}{
		{0x0400, 0x0e00},
		{0x0401, 0x0fff},
		{0x0402, 0x0ffe},
		{0x0405, 0x0ffb},

		{0x0500, 0x0aff},
		{0x0501, 0x0bff},

		{0x0600, 0x0eff},
		{0x0601, 0x2efe}, // 0x2eff + 0x0601 - 1
		{0x0602, 0x00ff}, // 0x0602 + 0x00ff - 1
	} {
		want := vm.Instruction(tc.want)
		got := view[tc.addr]

		if vm.Word(want) != got {
			t.Errorf("view[%v]: want: %v != got: %v", tc.addr, want, got)
		}
	}

	t.Logf("%+v", view[0x0600:0x060f])
}
