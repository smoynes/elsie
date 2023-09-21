package vm

import (
	"testing"
)

type TestDisplayAdapter *Display

func TestInterrupt(tt *testing.T) {
	var (
		t      = NewTestHarness(tt)
		intr   = Interrupt{}
		kbd    = &Keyboard{}
		isrKbd = ISR{vector: 0xad, driver: kbd}
		disp   = &DisplayDriver{
			handle: DeviceHandle[*Display, Display]{
				device: &Display{
					DSR: DisplayEnabled | DisplayReady,
				},
			},
		}
		isrDisp = ISR{vector: 0xdd, driver: disp}
	)

	intr.Register(PriorityHigh, isrKbd)
	intr.Register(PL6, isrDisp)

	idt := intr.idt[len(intr.idt)-1]
	if idt.vector != 0xad {
		t.Errorf("idt vector incorrect: want: %0#2x, got: %0#2x", 0xad, idt.vector)
	}

	if idt.driver != kbd {
		t.Errorf("idt vector incorrect: want: %s, got: %s", kbd, idt.driver)
	}

	if vec, ok := intr.Requested(PL0); !ok {
		t.Errorf("expected interrupt raised")
	} else if vec != 0xdd {
		t.Errorf("expected display interrupt vector: want: %0#2x, got: %0#2x", 0xdd, vec)
	}
}