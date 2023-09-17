package vm

// devices.go has devices and their drivers.

import (
	"fmt"
)

func NewDeviceDriver[T DrivableDevice](dev T) *DeviceDriver[T] {
	driver := new(DeviceDriver[T]) // TODO: no generics
	driver.device = &dev

	return driver
}

// IODevice represents a device has a single, lonely register for I/O.
type IODevice interface {
	Get() Register
	Put(Register)

	fmt.Stringer
}

type Driver interface {
	Configure(machine *LC3, dev DrivableDevice, addrs []Word)

	fmt.Stringer
}

type DeviceReader interface {
	DrivableDevice
	Read(addr Word) (Word, error)
}

type DeviceWriter interface {
	DrivableDevice
	Write(addr Word, val Register) error
}

type DrivableDevice interface {
	Device() string
	fmt.Stringer
}

type DeviceDriver[T DrivableDevice] struct {
	device *T
}

func (driver *DeviceDriver[T]) String() string {
	return fmt.Sprintf("DeviceDriver[%T](%s)", driver.device, driver.device)
}

type DisplayDriver struct {
	device DeviceDriver[Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word
}

func (driver *DisplayDriver) Configure(vm *LC3, drv DrivableDevice, addrs []Word) {
	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	var display *Display = drv.(*Display)
	display.DDR = 'X'
	display.DSR = 0x8000

	driver.device.device = display
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

func (device Display) Device() string {
	return device.String()
}

func (display Display) String() string {
	return fmt.Sprintf("Display(status:%s,data:%s)", display.DDR, display.DSR)
}

// Display is a device for outputting characters.
type Display struct {
	DSR, DDR Register
}

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	KBSR, KBDR Register
}

func (k Keyboard) String() string {
	return fmt.Sprintf("Keyboard(status:%s,data:%s)", k.KBDR, k.KBSR)
}

func (k Keyboard) Device() string { return "Keyboard(ModelM)" }

func (k *Keyboard) Configure(machine *LC3, dev DrivableDevice, addrs []Word) {
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
