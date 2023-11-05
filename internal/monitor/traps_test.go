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

	if testing.Verbose() {
		log.LogLevel.Set(log.Debug)
	}

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

func TestTrap_Out(tt *testing.T) {
	t := trapHarness{tt}

	if testing.Verbose() {
		log.LogLevel.Set(log.Debug)
	}

	obj, err := Generate(TrapOut)

	if err != nil {
		t.Error(err)
	}

	if len(obj.Code) < 15 {
		// Code must be AT LEAST 15 words: 12 instructions and a few bytes of data.
		t.Error("code too short", len(obj.Code))
	} else if len(obj.Code) >= 50 {
		t.Error("code too long", len(obj.Code))
	}

	// We want to test the trap in isolation, without any other traps loaded.
	image := SystemImage{
		log:     t.Logger(),
		Symbols: nil,
		Traps: []Routine{
			TrapOut,
		},
	}

	displayed := make(chan uint16, 10)
	machine := vm.New(
		WithSystemImage(&image),
		vm.WithDisplayListener(func(out uint16) {
			select {
			case displayed <- out:
			}
		}),
	)
	loader := vm.NewLoader(machine)

	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(
				vm.TRAP, uint16(vm.TrapOUT)).Encode(),
			),
		},
	}

	loader.Load(code)

	machine.REG[vm.R0] = 0x2365

	for i := 0; i < 100; i++ {
		err = machine.Step()

		t.Logf("Stepped\n%s\n%s\nerr %v", machine, machine.REG, err)

		if err != nil {
			t.Errorf("Step error %s", err)
			break
		} else if machine.PC > 0x3000 {
			break
		} else if !machine.MCR.Running() {
			break
		}
	}

	close(displayed)

	var vals []uint16
	for out := range displayed {
		vals = append(vals, out)
	}

	if len(vals) != 1 || vals[0] != 0x2365 {
		t.Errorf("displayed %+v", vals)
	}
}

func TestTrap_Puts(tt *testing.T) {
	t := trapHarness{tt}

	if testing.Verbose() {
		log.LogLevel.Set(log.Debug)
	}

	obj, err := Generate(TrapOut)

	if err != nil {
		t.Error(err)
	}

	if len(obj.Code) < 15 {
		// Code must be AT LEAST 15 words: 12 instructions and a few bytes of data.
		t.Error("code too short", len(obj.Code))
	} else if len(obj.Code) >= 50 {
		t.Error("code too long", len(obj.Code))
	}

	// We want to test this trap with another.
	image := SystemImage{
		log:     t.Logger(),
		Symbols: nil,
		Traps: []Routine{
			TrapPuts,
			TrapOut,
		},
	}

	displayed := make(chan uint16, 10)
	machine := vm.New(
		WithSystemImage(&image),
		vm.WithDisplayListener(func(out uint16) {
			select {
			case displayed <- out:
			}
		}),
	)
	loader := vm.NewLoader(machine)

	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(
				vm.TRAP, uint16(vm.TrapOUT)).Encode(),
			),
		},
	}

	loader.Load(code)

	code = vm.ObjectCode{
		Orig: 0x3100,
		Code: []vm.Word{0x2364, 0x2363, 0x2365, 0},
	}

	loader.Load(code)

	machine.REG[vm.R0] = 0x3100

	for i := 0; i < 20; i++ {
		err = machine.Step()

		t.Logf("Stepped\n%s\n%s\nerr %v", machine, machine.REG, err)

		if err != nil {
			t.Errorf("Step error %s", err)
			break
		} else if machine.PC > 0x3000 {
			break
		} else if !machine.MCR.Running() {
			break
		}
	}

	close(displayed)

	var vals []uint16
	for out := range displayed {
		vals = append(vals, out)
	}

	if len(vals) != 1 || vals[0] != 0x2365 {
		t.Errorf("displayed %+v", vals)
	}
}
