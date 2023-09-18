package vm

// devices.go has devices and their drivers.

import (
	"fmt"
)

// Driver represents an external device with which a program can read or write data.
//
// If a device has a simple I/O model and supports reading or writing a single word of data, it
// should implement the [RegisterData] interface. Otherwise, for more complicated state or I/O
// models, a device should implement the [Driver] interface, instead. See [DisplayDriver] for an
// example.
type Driver interface {
	Init(machine *LC3, addrs []Word)
	fmt.Stringer
}

// RegisterDevice represents a device that has a single, lonely register for I/O. In contrast to
// more complicated devices, a RegisterDevice does not have a device driver or other state so acts
// as its own driver. This type of driver exposes three operations:
//
//   - Init, to configure the device,
//   - Get, to read a word from the device, and
//   - Put, to write a word to the device.
//
// Devices can add validation or custom behaviour therein. -- a minimal implementation would look
// like:
//
//	type SomeDevice Register
//	func (dev *SomeDevice) Init(_ *LC3, _ []Word)      { *dev = 0x1234 }
//	func (dev SomeDevice)  Get() Register              { return Register(dev) }
//	func (dev *SomeDevice) Put(val Register)           { *dev = val }
//
// Abstractly, this models a single register, but in theory could be just about any kind of device.
type RegisterDevice interface {
	Driver
	Get() Register
	Put(Register)
}

// DeviceHandle is holds a reference to an external device. It is unnecessarily abstract and generic
// -- the typeset ranges over both the device and the device pointer type parameters.
type DeviceHandle[DP ~*D, D Driver] struct {
	device DP
}

// NewDeviceHandle creates a new driver for the given device. The driver is initialized with a
// reference to the device.
func NewDeviceHandle[DP interface {
	fmt.Stringer
	~*D
}, D Driver](device D) *DeviceHandle[DP, D] {
	handle := new(DeviceHandle[DP, D])
	handle.device = &device

	return handle
}

// Init initializes a handle's device.
func (handle *DeviceHandle[TP, T]) Init(vm *LC3, addrs []Word) {
	device := *handle.device
	device.Init(vm, addrs)
}

func (d *DeviceHandle[DP, D]) String() string {
	return (*d.device).String()
}

// ReadDriver is a driver that provides input to the machine from a device.
type ReadDriver interface {
	Driver
	Read(addr Word) (Word, error)
}

// WriteDriver is a driver that writes to a device.
type WriteDriver interface {
	Driver
	Write(addr Word, val Register) error
}

// DisplayDriver is a driver for an extremely simple terminal display.
type DisplayDriver struct {
	handle DeviceHandle[*Display, Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word
}

// Init initializes the display and the driver.
func (driver *DisplayDriver) Init(vm *LC3, addrs []Word) {
	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	driver.handle.Init(vm, addrs)
	driver.handle.device.DSR = 0x8000
}

// Read gets the status of the display device.
func (driver *DisplayDriver) Read(addr Word) (Word, error) {
	if addr == driver.statusAddr {
		return Word(driver.handle.device.DSR), nil
	}

	return Word(0xdea1), fmt.Errorf("read: %w: %s:%s", ErrNoDevice, addr, driver)
}

// Write sets the data register of the display device.
func (driver *DisplayDriver) Write(addr Word, value Register) error {
	if addr == driver.dataAddr {
		driver.handle.device.DDR = value
		println(value) // TODO: decode and output value

		return nil
	}

	println(value) // TODO: decode and output value

	return fmt.Errorf("write: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver DisplayDriver) String() string {
	return fmt.Sprintf("DisplayDriver(display:%s)", driver.handle.device)
}

// Display is a device for outputting characters.
type Display struct {
	DSR, DDR Register
}

func (d Display) Init(_ *LC3, _ []Word) {
	println("Hello!")
}

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
