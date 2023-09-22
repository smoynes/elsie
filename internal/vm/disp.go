package vm

import (
	"fmt"
)

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

func (driver *DisplayDriver) InterruptRequested() bool {
	// For our purposes, the display is always ready.
	return driver.handle.device != nil &&
		(driver.handle.device.DSR|DisplayReady)&DisplayEnabled != 0
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
	// Display status register. The top two bits encode the interrupt-ready and enable flags.
	//
	//  | IR | IE |       |
	//  +----+----+-------+
	//  | 15 | 14 |13    0|
	//.
	DSR Register

	DDR Register
}

func (d Display) Init(_ *LC3, _ []Word) {
	println("Hello! 🍏")
}

const (
	DisplayReady   = Register(1 << 15)
	DisplayEnabled = Register(1 << 14)
)

func (display Display) String() string {
	return fmt.Sprintf("Display(status:%s,data:%s)", display.DSR, display.DDR)
}
