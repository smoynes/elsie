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
		kbd                = NewKeyboard()
		driver Driver      = kbd
		reader ReadDriver  = kbd
		writer WriteDriver = kbd
	)

	kbd.KBDR = uninitialized
	kbd.KBSR = uninitialized

	t.Logf("cool 🕶️ %s", kbd)

	driver.Init(vm, nil)

	addr := KBSRAddr
	if err := writer.Write(addr, Register(0xffff)); err != nil {
		t.Error(err)
	} else if got, err := reader.Read(addr); err != nil {
		t.Errorf("read status: %s", err)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized status register: %s", addr)
	} else if got != Word(0xffff) {
		t.Errorf("status register unwritten: %s", addr)
	}

	addr = KBDRAddr
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("expected read error: %s", addr)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized data register: %s:%s", addr, got)
	}

	addr = KBSRAddr
	if got, err := reader.Read(addr); err != nil {
		t.Errorf("read error: %s: %s", addr, err)
	} else if got != Word(KeyboardEnable|KeyboardReady) {
		t.Errorf("expected status ready: want: %s, got: %s", KeyboardEnable|KeyboardReady, got)
	}
}

func TestDisplayDriver(tt *testing.T) {
	var (
		t             = NewTestHarness(tt)
		vm            = t.Make()
		display       = NewDisplay()
		displayDriver = NewDisplayDriver(display)
		statusAddr    = Word(0xface)
		dataAddr      = Word(0xf001)
	)

	display.dsr = uninitialized
	display.ddr = uninitialized

	displayDriver.Init(vm, []Word{statusAddr, dataAddr})

	if got, err := displayDriver.Read(statusAddr); err != nil {
		t.Error(err)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized status register: %s, want: %s, got: %s",
			statusAddr, Word(0x8000), got)
	} else if got != Word(0x8000) {
		t.Errorf("unexpected status register: %s, want: %s, got: %s",
			statusAddr, Word(0x8000), got)
	}

	if got, err := displayDriver.Read(dataAddr); err == nil {
		t.Errorf("expected read error: %s", dataAddr)
	} else if got == Word(uninitialized) {
		t.Errorf("uninitialized display register: %s:%s", dataAddr, got)
	}

	val := Register('?')
	if err := displayDriver.Write(dataAddr, val); err != nil {
		t.Errorf("write error: %s: %s", dataAddr, err)
	}

	if got, err := displayDriver.Read(statusAddr); err != nil {
		t.Errorf("write error: %s: %s", dataAddr, err)
	} else if got != Word(DisplayReady) {
		t.Errorf("expected status: %s, got: %s", Word(DisplayReady), got)
	}
}
