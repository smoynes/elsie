package monitor

import (
	"errors"
	"testing"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

type trapHarness struct{ *testing.T }

func (*trapHarness) Logger() *log.Logger {
	log.LogLevel.Set(log.Debug)
	return log.DefaultLogger()
}

func TestTrap_Halt(tt *testing.T) {
	t := trapHarness{tt}

	obj, err := Generate(TrapHalt)

	if err != nil {
		t.Error(err)
	}

	if len(obj.Code) < 9 {
		// Code must be AT LEAST 10 words: 7 instructions and a few bytes of data.
		t.Error("code too short", len(obj.Code))
	} else if len(obj.Code) >= 50 {
		// It really should not be TOO long.
		t.Error("code too long", len(obj.Code))
	}

	// We wish to test this trap without depending upon others so we stub the OUT trap.
	putsRoutine := Routine{
		Name:   "Stub OUT",
		Orig:   TrapOut.Orig,
		Vector: TrapOut.Vector,
		Code: []asm.Operation{
			&asm.RTI{},
		},
		Symbols: map[string]uint16{},
	}

	image := SystemImage{
		log:     t.Logger(),
		Symbols: nil,
		Traps: []Routine{
			TrapHalt,
			putsRoutine},
	}

	machine := vm.New(
		WithSystemImage(&image),
	)

	// Now, we create code to execute the trap under test.
	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(vm.TRAP, 0x25).Encode()),
		},
	}

	loader := vm.NewLoader(machine)
	loader.Load(code)

	machine.MCR = 0xffff

	for i := 0; i < 100; i++ {
		err = machine.Step()
		if errors.Is(err, vm.ErrHalted) {
			break
		} else if err != nil {
			t.Error(err)
		}
	}

	if machine.MCR.Running() {
		t.Error("pc", machine)
	}
}
