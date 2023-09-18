package vm

// devices.go has devices and their drivers.

import (
	"fmt"
)

// NewDeviceDriver creates a new driver for the given device. The driver is initialized with a
// reference to the device.
func NewDeviceDriver[T Device](device T) *DeviceDriver[T] {
	driver := new(DeviceDriver[T]) // TODO: no generics
	driver.device = &device

	return driver
}

// Device is the target of I/O.
type Device interface {
	// Init is called at startup to initialize a device or its driver.
	Init(machine *LC3, addrs []Word)

	fmt.Stringer
}

// RegisterDevice represents a device that has a single, lonely register for I/O. In contrast to more
// complicated devices, a RegisterDevice does not have a device driver or other state.
type RegisterDevice interface {
	Device
	Get() Register
	Put(Register)
}

type DeviceReader interface {
	Device
	Read(addr Word) (Word, error)
}

type DeviceWriter interface {
	Device
	Write(addr Word, val Register) error
}

type DeviceDriver[T Device] struct {
	device *T
}

func (d *DeviceDriver[T]) String() string {
	return fmt.Sprintf("DeviceDriver[%T](%s)", d.device, d.device)
}

// DisplayDriver is a device for an extremely simple terminal display.
type DisplayDriver struct {
	device DeviceDriver[Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word
}

func (driver *DisplayDriver) Init(vm *LC3, addrs []Word) {
	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	device := driver.device.device
	device.DSR = 0x8000
}

func (driver *DisplayDriver) Read(addr Word) (Word, error) {
	if addr == driver.statusAddr {
		return Word(driver.device.device.DSR), nil
	}

	return Word(0xdea1), fmt.Errorf("read: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver *DisplayDriver) Write(addr Word, value Register) error {
	if addr == driver.dataAddr {
		driver.device.device.DDR = value
		println(value) // TODO: decode and output value

		return nil
	}

	println(value) // TODO: decode and output value

	return fmt.Errorf("write: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver DisplayDriver) Device() string {
	return driver.device.String()
}

func (driver DisplayDriver) String() string {
	return fmt.Sprintf("DisplayDriver(display:%s)", driver.device.device)
}

// Display is a device for outputting characters.
type Display struct {
	DSR, DDR Register
}

func (d Display) Init(_ *LC3, _ []Word) {}

func (display Display) String() string {
	return fmt.Sprintf("Display(status:%s,data:%s)", display.DDR, display.DSR)
}

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	KBSR, KBDR Register
}

func (k Keyboard) String() string {
	return fmt.Sprintf("Keyboard(status:%s,data:%s)", k.KBDR, k.KBSR)
}

func (k Keyboard) Device() string { return "Keyboard(ModelM)" }

func (k *Keyboard) Init(machine *LC3, _ []Word) {
	k.KBSR = 0x0000
	k.KBDR = 0x0000
}

func (k *Keyboard) Read(addr Word) (Word, error) {
	if addr == KBSRAddr {
		return Word(k.KBSR), nil
	}

	k.KBSR = 0x0000

	return Word(k.KBDR), nil
}

func (k *Keyboard) Write(addr Word, val Register) error {
	if addr != KBSRAddr {
		return fmt.Errorf("kbd: %w: %s", ErrNoDevice, addr)
	}

	k.KBSR = val

	return nil
}
