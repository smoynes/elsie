package vm

// devices.go has devices and their drivers.

import (
	"fmt"
)

// Device represents an external device with which a program can read or write data.
//
// If a device has a simple I/O model and supports reading or writing a single word of data, it can
// implement the [RegisterDevice] interface. Otherwise, for more complicated device state or I/O
// models, a device should implement the [Driver] interface, instead. See [DisplayDriver] for an
// example.
type Device interface {

	// Init initializes the device during system startup. A device should configure interrupts,
	// initialize device-state, and allocate resources, as needed.
	Init(machine *LC3, addrs []Word)

	// Stringer for debugs.
	fmt.Stringer
}

// RegisterDevice represents a device that has a single, lonely register for I/O. In contrast to
// more complicated devices, a RegisterDevice does not have other device state and can act as its
// own driver. This type of device exposes three operations:
//
//   - Init, to configure the device,
//   - Get, to read a word from the device, and
//   - Put, to write a word to the device.
//
// Devices may add validation or custom behaviour therein. A minimal implementation would look like:
//
//	type SomeDevice Register
//	func (dev *SomeDevice) Init(_ *LC3, _ []Word)      { *dev = 0x1234 }
//	func (dev SomeDevice)  Get() Register              { return Register(dev) }
//	func (dev *SomeDevice) Put(val Register)           { *dev = val }
//
// Abstractly, this models a single register, but in theory could be just about any kind of device.
type RegisterDevice interface {
	Device

	Get() Register
	Put(Register)
}

// A Driver is the controller for a device. Device drivers may request interrupts, if registered
// with the interrupt controller.
type Driver interface {
	Device

	// InterruptRequested returns true if the device has requested I/O service and interrupts
	// are enabled for the device.
	InterruptRequested() bool
}

// ReadDriver is a driver that provides input to the machine from a device.
type ReadDriver interface {
	Device

	Read(addr Word) (Word, error)
}

// WriteDriver is a driver that writes to a device.
type WriteDriver interface {
	Device

	Write(addr Word, val Register) error
}

// NewDeviceHandle creates a new driver for the given device. The driver is initialized with a
// reference to the device.
func NewDeviceHandle[DP interface {
	fmt.Stringer
	~*D
}, D Device](device D) *DeviceHandle[DP, D] {
	handle := new(DeviceHandle[DP, D])
	handle.device = &device

	return handle
}

// DeviceHandle is holds a reference to an external device. It is unnecessarily abstract and generic
// -- the typeset ranges over both the device and the device pointer type parameters.
type DeviceHandle[DP ~*D, D Device] struct {
	device DP
}

// Init initializes a handle's device.
func (handle *DeviceHandle[DP, D]) Init(vm *LC3, addrs []Word) {
	device := *handle.device
	device.Init(vm, addrs) // This is weird...
}

func (d *DeviceHandle[DP, D]) String() string {
	return (*d.device).String()
}
