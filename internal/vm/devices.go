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
	device() string
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
//	func (dev SomeDevice)  device()          string   { return "CoolDevice()" }
//	func (dev SomeDevice)  Get()             Register { return Register(dev)  }
//	func (dev *SomeDevice) Put(val Register)          { *dev = val            }
//
// Abstractly, this models a single register, but in theory could be just about any kind of device.
type RegisterDevice interface {
	Device

	Get() Register
	Put(Register)

	// Stringer for debugs.
	fmt.Stringer
}

// Drivers are controllers for a devices. Device drivers may request interrupts, if registered with
// the interrupt controller.
type Driver interface {
	Device

	// InterruptRequested returns true if the device has requested I/O service and interrupts
	// are enabled for the device.
	InterruptRequested() bool

	// Init initializes the device during system startup. The driver can configure interrupts,
	// initialize device-state, and allocate resources, as needed.
	Init(machine *LC3, addrs []Word)

	// Stringer for debugs.
	fmt.Stringer
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

// DeviceHandle is holds a reference to an external device. It is unnecessarily abstract and generic
// -- the typeset ranges over both the device- and the reference- type parameters.
type DeviceHandle[DP DeviceP[D], D Device] struct {
	device DP
}

// DeviceP is a type constraint for references to devices.
type DeviceP[D Device] interface {
	~*D

	// Init initializes the device during system startup. Importantly, this method should allocate
	// locks. (?)
	Init(machine *LC3, addrs []Word)

	// For debugs.
	fmt.Stringer
}

// NewDeviceHandle creates a new reference to the given device.
func NewDeviceHandle[DP DeviceP[D], D Device](dev DP) DeviceHandle[DP, D] { /* 🫰 */
	handle := new(DeviceHandle[DP, D])
	handle.device = dev

	return *handle
}

// Init initializes a handle's device.
func (handle DeviceHandle[DP, D]) Init(vm *LC3, addrs []Word) {
	handle.device.Init(vm, addrs)
}

func (d DeviceHandle[DP, D]) String() string {
	return d.device.String()
}
