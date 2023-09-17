package vm

import (
	"testing"
)

// Type assertions for expected devices.
var (
	// CPU registers are simple I/O devices.
	_ IODevice = (*ProcessorStatus)(nil)
	_ IODevice = (*ControlRegister)(nil)

	// Display has a driver.
	d                = &DisplayDriver{}
	_ Driver         = d
	_ DeviceWriter   = d
	_ DeviceReader   = d
	_ DrivableDevice = d

	// Keyboard is its own driver.
	k                = &Keyboard{}
	_ Driver         = k
	_ DrivableDevice = k
	_ DeviceWriter   = d
	_ DeviceReader   = d
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
		driver Driver       = &kbd
		reader DeviceReader = &kbd
		writer DeviceWriter = &kbd
	)

	driver.Configure(vm, &kbd, nil)

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
		driver        = NewDeviceDriver(Display{})
		deviceDriver  = driver
		displayDriver = &DisplayDriver{*deviceDriver, Word(0xface), Word(0xf001)}
	)

	displayDriver.device.device.DSR = uninitializedRegister
	displayDriver.device.device.DDR = uninitializedRegister

	displayDriver.Configure(vm, &Display{}, []Word{0xface, 0xf001})

	addr := Word(0xface)

	if got, err := displayDriver.Read(addr); err != nil {
		t.Error(err)
	} else if got == Word(uninitializedRegister) {
		t.Errorf("uninitialized status register: %s", addr)
	} else if got != Word(0x8000) {
		t.Errorf("uninitialized status register: %s", addr)
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

	driver.device.Device()
}
