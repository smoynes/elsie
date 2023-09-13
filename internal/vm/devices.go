package vm

// devices.go has devices and their drivers.

import (
	"fmt"
	"log"
)

type (
	Device struct {
		driver Driver
		status DeviceRegister
		data   DeviceRegister
	}
	DeviceRegister Register
	Driver         interface{ driver() }
	DeviceReader   interface {
		Driver
		Read(dev Device, addr Word) DeviceRegister
	}
	DeviceWriter interface {
		Driver
		Write(dev Device, addr Word, data DeviceRegister)
	}

	// Keyboard is a hardwired input device for typos.
	Keyboard Device
)

func newDevice(drv Driver) Device {
	return Device{
		status: DeviceRegister(0x0000),
		data:   DeviceRegister('!'),
		driver: drv,
	}
}

func (d Device) String() string {
	return fmt.Sprintf("dev: status: %s, data: %s", d.status, d.data)
}
func (d DeviceRegister) String() string {
	return Register(d).String()
}

func (d Device) Read(addr Word) DeviceRegister {
	log.Printf("dev: read: %T, %s", d.driver, addr)

	if driver, ok := d.driver.(DeviceReader); ok {
		return driver.Read(d, addr)
	}
	panic(fmt.Sprintf("dev: read: %T", d.driver))
}

func (d Device) Write(addr Word, word DeviceRegister) {
	log.Printf("dev: read: %T, %s", d.driver, addr)

	if driver, ok := d.driver.(DeviceWriter); ok {
		driver.Write(d, addr, word)
	} else {
		panic(fmt.Sprintf("dev: write: %T", d.driver))
	}
}

type KeyboardDriver struct{}

func (KeyboardDriver) driver() {}

type KeyboardStatus DeviceRegister
type KeyboardData DeviceRegister

const KeyboardReady = KeyboardStatus(1 << 15) // READY

func (k KeyboardDriver) Read(dev Device, addr Word) DeviceRegister {
	log.Printf("kbd: read: addr: %s, status: %s, data: %s", addr, dev.status, dev.data)

	switch addr {
	case KBSRAddr:
		log.Printf("kbd: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.status)
	case KBDRAddr:
		dev.status = 0x0000
		log.Printf("kbd: addr: %s, status: %s, data: %s\n", addr, dev.status, dev.data)
		return DeviceRegister(dev.data)
	default:
		panic("kbd: read: bad addr: " + addr.String())
	}
}

func (k KeyboardDriver) Write(dev Device, addr Word, word DeviceRegister) {
	log.Printf("kbd: write: addr: %s, status: %s, data: %s", addr, dev.status, word)

	switch addr {
	case KBSRAddr:
		dev.status = word
	default:
		panic("kbd: write: bad addr: " + addr.String())
	}
}
