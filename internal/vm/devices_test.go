package vm

import (
	"testing"
)

// Type assertions for expected devices.
var (
	// CPU registers are simple I/O devices.
	_ RegisterDevice = (*ProcessorStatus)(nil)
	_ RegisterDevice = (*ControlRegister)(nil)

	// Display has a driver.
	d             = &DisplayDriver{}
	_ Driver      = d
	_ WriteDriver = d
	_ ReadDriver  = d

	// Keyboard is its own driver.
	k             = &Keyboard{}
	_ Driver      = k
	_ WriteDriver = k
	_ ReadDriver  = k
)

var uninitializedRegister = Register(0x0101)

func TestKeyboardDriver(tt *testing.T) {
	t := NewTestHarness(tt)
	vm := t.Make()

	var (
		kbd = Keyboard{
			KBSR: uninitializedRegister,
			KBDR: uninitializedRegister,
		}
		driver Driver      = &kbd
		reader ReadDriver  = &kbd
		writer WriteDriver = &kbd
	)

	driver.Init(vm, nil)

	t.Log(kbd.Device())
	t.Logf("cool 🕶️ %s", kbd)

	addr := Word(KBSRAddr)

	if err := writer.Write(addr, Register(0xffff)); err != nil {
		t.Error(err)
	} else if got, err := reader.Read(addr); err != nil {
		t.Error(err)
	} else if got == Word(uninitializedRegister) {
		t.Errorf("uninitialized status register: %s", addr)
	} else if got != Word(0xffff) {
		t.Errorf("uninitialized status register: %s", addr)
	}

	addr = Word(KBDRAddr)
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("expected read error: %s", addr)
	} else if got == Word(uninitializedRegister) {
		t.Errorf("uninitialized data register: %s:%s", addr, got)
	}

	addr = Word(KBSRAddr)
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("read error: %s: %s", addr, err)
	} else if got != Word(0x0000) {
		t.Errorf("unexpected status: want: %s, got: %s", Word(0x0000), got)
	}
}

func TestDisplayDriver(tt *testing.T) {
	var (
		t             = NewTestHarness(tt)
		vm            = t.Make()
		display       = Display{}
		handle        = NewDeviceHandle[*Display](display)
		displayDriver = &DisplayDriver{*handle, Word(0xface), Word(0xf001)}
	)

	displayDriver.handle.device.DSR = uninitializedRegister
	displayDriver.handle.device.DDR = uninitializedRegister

	displayDriver.Init(vm, []Word{0xface, 0xf001})

	addr := Word(0xface)

	if got, err := displayDriver.Read(addr); err != nil {
		t.Error(err)
	} else if got == Word(uninitializedRegister) {
		t.Errorf("uninitialized status register: %s, want: %s, got: %s",
			addr, Word(0x8000), got)
	} else if got != Word(0x8000) {
		t.Errorf("unexpected status register: %s, want: %s, got: %s",
			addr, Word(0x8000), got)
	}

	addr = Word(0xf001)

	if got, err := displayDriver.Read(addr); err == nil {
		t.Errorf("expected read error: %s", addr)
	} else if got == Word(uninitializedRegister) {
		t.Errorf("uninitialized display register: %s:%s", addr, got)
	}

	val := Register('?')
	if err := displayDriver.Write(addr, val); err != nil {
		t.Errorf("write error: %s: %s", addr, err)
	}
}
