package vm

// devices.go has devices and their drivers.

import (
	"fmt"
	"log"
)

// Device is an external device. Every device has a [Driver] that
// exposes input/output operations to the system.
type Device struct {
	driver Driver
	status DeviceRegister
	data   DeviceRegister
}

func (d Device) String() string {
	return fmt.Sprintf("dev: status: %s, data: %s", d.status, d.data)
}

// newDevice creates a device using a driver to control access.
func newDevice(vm *LC3, drv Driver, addrs []Word) Device {
	dev := Device{
		status: DeviceRegister(0x0000),
		data:   DeviceRegister('?'),
		driver: drv,
	}

	if drv != nil {
		drv.Configure(vm, &dev, addrs)
	}

	return dev
}

// Read delegates reads to a device's driver if driver is readable. Otherwise,
// zero is returned.
func (d *Device) Read(addr Word) DeviceRegister {
	log.Printf("dev: read: %T, %s", d.driver, addr)

	if driver, ok := d.driver.(DeviceReader); ok {
		return driver.Read(d, addr)
	}

	log.Printf("dev: read unsupported: %T", d.driver)
	return DeviceRegister(0x0000)
}

// DeviceRegister represents a register on a (virtual) device and is how
// the (virtual) machine exchanges data with the world.
type DeviceRegister Register

func (d DeviceRegister) String() string {
	return Register(d).String()
}

// A Driver controls a device.
type Driver interface {
	// Configure initializes the device and machine for I/O.
	Configure(machine *LC3, dev *Device, addrs []Word)
}

// DeviceReader is a driver that supports reading from a device register.
type DeviceReader interface {
	Driver
	Read(dev *Device, addr Word) DeviceRegister
}

// DeviceWriter is a driver that supports writing a value to a device
// register.
type DeviceWriter interface {
	Driver
	Write(dev *Device, addr Word, val DeviceRegister) // TODO: error
}

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard Device

func (d *Device) Write(addr Word, word DeviceRegister) {
	log.Printf("dev: read: %T, %s", d.driver, addr)

	if driver, ok := d.driver.(DeviceWriter); ok {
		driver.Write(d, addr, word)
	} else {
		panic(fmt.Sprintf("dev: write: %T", d.driver))
	}
}

const kbdStatusReady = DeviceRegister(1 << 15) // READY

// Configure sets up the keyboard.
func (k *Keyboard) Configure(machine *LC3, dev *Device, addrs []Word) {}

// Read gets the value of the last pressed key and clears the device's ready status.
func (k *Keyboard) Read(dev *Device, addr Word) DeviceRegister {
	log.Printf("kbd: read: addr: %s, status: %s, data: %s", addr, dev.status, dev.data)

	switch addr {
	case KBSRAddr:
		log.Printf("kbd: read: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.status)
	case KBDRAddr:
		dev.status ^= kbdStatusReady
		log.Printf("kbd: read: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.data)
	default:
		panic("kbd: read: bad addr: " + addr.String())
	}
}

// Write sets the value of the device status register.
func (k *Keyboard) Write(dev *Device, addr Word, word DeviceRegister) {
	log.Printf("kbd: write: addr: %s, status: %s, data: %s", addr, dev.status, word)

	switch addr {
	case KBSRAddr:
		dev.status = word
	case KBDRAddr:
		log.Printf("kbd: write: addr: %s", addr)
	default:
		panic("kbd: write: bad addr: " + addr.String())
	}
}
