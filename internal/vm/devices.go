package vm

// devices.go has devices and their drivers.

import (
	"fmt"
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

func (d Device) Log(msg string) {
	if d.driver != nil {
		d.driver.Log(msg)
	}
}

func (d Device) Logf(format string, args ...any) {
	if d.driver != nil {
		d.driver.Log(fmt.Sprintf(format, args...))
	}
}

// newDevice creates a device using a driver to control access.
func newDevice(vm *LC3, driver Driver, addrs []Word) Device {
	dev := Device{
		status: DeviceRegister(0x0000),
		data:   DeviceRegister(0x003f), // ?
	}

	if driver != nil {
		dev.driver = driver
		driver.Configure(vm, &dev, addrs)
	}

	return dev
}

// Read delegates reads to a device's driver if driver is readable. Otherwise,
// zero is returned.
func (d *Device) Read(addr Word) (DeviceRegister, error) {
	d.driver.Log("%s")
	d.Logf("dev: read: %s", d)

	if driver, ok := d.driver.(DeviceReader); ok {
		return driver.Read(d, addr)
	}

	d.Logf("dev: read unsupported: %T", d.driver) // TODO

	return DeviceRegister(0x0000), nil
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
	Log(string)
}

// DeviceReader is a driver that supports reading from a device register.
type DeviceReader interface {
	Driver
	Read(dev *Device, addr Word) (DeviceRegister, error)
}

// DeviceWriter is a driver that supports writing a value to a device
// register.
type DeviceWriter interface {
	Driver
	Write(dev *Device, addr Word, val DeviceRegister) error
}

func (d *Device) Write(addr Word, word DeviceRegister) error {
	d.Logf("dev: read: %T, %s", d.driver, addr)

	if driver, ok := d.driver.(DeviceWriter); ok {
		return driver.Write(d, addr, word)
	} else {
		return fmt.Errorf("dev: write: %T", d.driver)
	}
}

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	Device

	log logger
}

var (
	_ DeviceReader = &Keyboard{}
)

func (k *Keyboard) String() string {
	return fmt.Sprintf("kbd: %s %v", k.Device, k.log)
}

const kbdStatusReady = DeviceRegister(1 << 15) // READY

// Configure sets up the driver.
func (k *Keyboard) Configure(machine *LC3, dev *Device, addrs []Word) {
	k.log = machine.log
}

func (k *Keyboard) WithLogger(log logger) {
	k.log = log
}

// Read gets the value of the last pressed key and clears the device's ready status.
func (k *Keyboard) Read(dev *Device, addr Word) (DeviceRegister, error) {
	k.log.Printf("kbd: read: addr: %s, status: %s, data: %s", addr, dev.status, dev.data)
	k.log.Printf(fmt.Sprintf("kbd: read: addr: %s, status: %s, data: %s", addr, dev.status, dev.data))

	switch addr {
	case KBSRAddr:
		k.Logf("kbd: read: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.status), nil
	case KBDRAddr:
		k.status ^= kbdStatusReady
		k.Logf("kbd: read: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.data), nil
	default:
		return 0, fmt.Errorf("kbd: read: bad addr: %s", addr) // TODO
	}
}

// Write sets the value of the device status register.
func (k *Keyboard) Write(dev *Device, addr Word, word DeviceRegister) error {
	k.Logf("kbd: write: addr: %s, status: %s, data: %s", addr, dev.status, word)

	switch addr {
	case KBSRAddr:
		dev.status = word
		return nil
	case KBDRAddr:
		k.Logf("kbd: write: addr: %s", addr)
		return nil
	default:
		return fmt.Errorf("kbd: write: bad addr: %s", addr)
	}
}
