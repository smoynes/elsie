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
	_ Device      = d
	_ WriteDriver = d
	_ ReadDriver  = d

	// Keyboard is its own driver.
	k             = &Keyboard{}
	_ Device      = k
	_ WriteDriver = k
	_ ReadDriver  = k
)

var uninitialized = Register(0x0101)

func TestKeyboardDriver(tt *testing.T) {
	t := NewTestHarness(tt)
	vm := t.Make()

	var (
		kbd    *Keyboard   = NewKeyboard()
		driver Device      = kbd
		reader ReadDriver  = kbd
		writer WriteDriver = kbd
	)

	kbd.KBDR = uninitialized
	kbd.KBSR = uninitialized

	t.Logf("cool 🕶️ %s", kbd)

	driver.Init(vm, nil)
	addr := Word(KBSRAddr)

	if err := writer.Write(addr, Register(0xffff)); err != nil {
		t.Error(err)
	} else if got, err := reader.Read(addr); err != nil {
		t.Errorf("read status: %s", err)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized status register: %s", addr)
	} else if got != Word(0xffff) {
		t.Errorf("status register unwritten: %s", addr)
	}

	addr = Word(KBDRAddr)
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("expected read error: %s", addr)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized data register: %s:%s", addr, got)
	}

	addr = Word(KBSRAddr)
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("read error: %s: %s", addr, err)
	} else if got != Word(KeyboardEnable) {
		t.Errorf("expected status ready: want: %s, got: %s", KeyboardEnable, got)
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

	displayDriver.handle.device.DSR = uninitialized
	displayDriver.handle.device.DDR = uninitialized

	displayDriver.Init(vm, []Word{0xface, 0xf001})

	addr := Word(0xface)

	if got, err := displayDriver.Read(addr); err != nil {
		t.Error(err)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized status register: %s, want: %s, got: %s",
			addr, Word(0x8000), got)
	} else if got != Word(0x8000) {
		t.Errorf("unexpected status register: %s, want: %s, got: %s",
			addr, Word(0x8000), got)
	}

	addr = Word(0xf001)

	if got, err := displayDriver.Read(addr); err == nil {
		t.Errorf("expected read error: %s", addr)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized display register: %s:%s", addr, got)
	}

	val := Register('?')
	if err := displayDriver.Write(addr, val); err != nil {
		t.Errorf("write error: %s: %s", addr, err)
	}
}
