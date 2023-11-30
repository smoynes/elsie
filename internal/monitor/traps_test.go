package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

type trapHarness struct{ *testing.T }

func NewHarness(t *testing.T) *trapHarness {
	t.Helper()

	tt := trapHarness{t}

	if testing.Verbose() {
		log.LogLevel.Set(log.Debug)
	} else {
		log.LogLevel.Set(log.Warn)
	}

	return &tt
}

func (*trapHarness) Logger() *log.Logger {
	return log.DefaultLogger()
}

func TestTrap_Halt(tt *testing.T) {
	t := NewHarness(tt)

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
		Name:   "Stub PUTS",
		Orig:   TrapPuts.Orig,
		Vector: TrapPuts.Vector,
		Code: []asm.Operation{
			&asm.RTI{},
		},
		Symbols: map[string]uint16{},
	}

	image := SystemImage{logger: t.Logger(), Symbols: nil, Traps: []Routine{TrapHalt, putsRoutine}}

	machine := vm.New(
		WithSystemImage(&image),
	)

	// Now, we create code to execute the trap under test.
	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.NewInstruction(vm.TRAP, 0x25).Encode(),
		},
	}

	loader := vm.NewLoader(machine)
	unsafeLoad(loader, code)

	machine.MCR = 0xffff

	for i := 0; i < 300; i++ {
		err = machine.Step()

		if testing.Verbose() {
			t.Logf("Stepped\n%s\n%s\nerr %v", machine, machine.REG, err)
		}

		if errors.Is(err, vm.ErrHalted) {
			break
		} else if !machine.MCR.Running() {
			break
		} else if err != nil {
			t.Error(err)
			break
		}
	}

	if machine.MCR.Running() {
		t.Errorf("MCR not stopped.\n%s\n", machine)
	}
}

func TestTrap_Out(tt *testing.T) {
	t := NewHarness(tt)

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
		logger:  t.Logger(),
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
			vm.NewInstruction(vm.TRAP, uint16(vm.TrapOUT)).Encode(),
		},
	}

	unsafeLoad(loader, code)

	machine.REG[vm.R0] = 0x2365

	for i := 0; i < 100; i++ {
		err = machine.Step()

		if testing.Verbose() {
			t.Logf("Stepped\n%s\n%s\nerr %v", machine, machine.REG, err)
		}

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

	vals := make([]uint16, 0, len(displayed))
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
		logger:  t.Logger(),
		Symbols: nil,
		Traps: []Routine{
			TrapPuts,
			TrapOut,
		},
	}

	displayed := make(chan uint16, 10)
	machine := vm.New(
		WithSystemImage(&image),

		// TODO: the names out and displayed are inverted by meaning here.
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
			vm.NewInstruction(
				vm.TRAP, uint16(vm.TrapPUTS)).Encode(),
		},
	}

	unsafeLoad(loader, code)

	machine.REG[vm.R0] = 0x3100
	code = vm.ObjectCode{
		Orig: 0x3100,
		Code: []vm.Word{vm.Word('!'), vm.Word('!' + 1), vm.Word('!' + 2), 0},
	}

	unsafeLoad(loader, code)

	// We expect that only a few dozen instructions are executed to output a few bytes.
	for i := 0; i < 100; i++ { // TODO
		err = machine.Step()

		if testing.Verbose() {
			t.Logf("Stepped\n%s\n%s\nerr %v", machine.String(), machine.REG, err)
		}

		if err != nil {
			t.Errorf("Step error %s", err)
			break
		} else if machine.PC > 0x3000 {
			break
		} else if !machine.MCR.Running() {
			break
		}
	}

	time.Sleep(100 * time.Millisecond)
	close(displayed)

	vals := make([]uint16, 0, len(displayed))
	for out := range displayed {
		vals = append(vals, out)
	}

	t.Log("displayed", len(vals), "values:")

	if len(vals) != 3 {
		t.Errorf("expected 3 displayed values, got %d", len(vals))
	}

	for i := range vals {
		if vals[i] != 0x0021+uint16(i) {
			t.Errorf("vals[%d] != 0x0022, got: %04x", i, vals[i])
		}
	}
}

func unsafeLoad(loader *vm.Loader, code vm.ObjectCode) {
	_, err := loader.Load(code)
	if err != nil {
		panic(err.Error())
	}
}
