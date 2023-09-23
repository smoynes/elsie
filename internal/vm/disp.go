package vm

import (
	"context"
	"fmt"
	"sync"
)

// DisplayDriver is a driver for an extremely simple terminal display.
type DisplayDriver struct {
	handle DeviceHandle[*Display, Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word
}

func WithDisplayDriver(parent context.Context) (context.Context, *DisplayDriver, context.CancelFunc) {
	display := &Display{
		Mutex: &sync.Mutex{},
	}
	driver := NewDisplayDriver(display)

	ctx, cancel := context.WithCancel(parent)

	return ctx, driver, cancel
}

func NewDisplayDriver(display *Display) *DisplayDriver {
	handle := NewDeviceHandle(display)
	driver := DisplayDriver{
		handle:     handle,
		statusAddr: DSRAddr,
		dataAddr:   DDRAddr,
	}

	return &driver
}

// Init initializes the display and the driver.
func (driver *DisplayDriver) Init(vm *LC3, addrs []Word) {
	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	driver.handle.Init(vm, addrs)
}

// Read gets the status of the display device.
func (driver *DisplayDriver) Read(addr Word) (Word, error) {
	if addr == driver.statusAddr {
		return Word(driver.handle.device.DSR()), nil
	}

	return Word(0xdea1), fmt.Errorf("read: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver *DisplayDriver) InterruptRequested() bool {
	// For our purposes, the display never interrupts the CPU.

	return driver.handle.device != nil &&
		driver.handle.device.DSR() == (DisplayReady|DisplayEnabled)
}

// Write sets the data register of the display device.
func (driver *DisplayDriver) Write(addr Word, value Register) error {
	if addr == driver.dataAddr {
		driver.handle.device.Write(value)
		return nil
	}

	return fmt.Errorf("write: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver *DisplayDriver) String() string {
	if driver.handle.device != nil && driver.handle.device.Mutex != nil {
		return fmt.Sprintf("DisplayDriver(display:%s)", driver.handle.device)
	} else {
		return "DisplayDriver(display:nil)"
	}
}

func (driver *DisplayDriver) device() string {
	return driver.String()
}

// Display is a device for outputting characters. It has a status register (DSR) and a data register
// (DDR).
//
// When the machine writes to DDR, the device clears its interrupt-ready status-flag to indicate the
// display buffer, as it were, is full. When the device outputs the character, it sets the ready
// flag again. If the machine writes again to DDR, before the device clears the buffer, data is
// overwritten and precious data is lost. Don't do it! Instead, programs should poll until the
// device is ready..
type Display struct {
	// sync.Mutex provides atomic R/W to the device registers.
	*sync.Mutex

	// Display Status Register. The top two bits encode the interrupt-ready and enable flags. When
	// IR is set, the display can receive another character for output.
	//
	//  | IR | IE |       |
	//  +----+----+-------+
	//  | 15 | 14 |13    0|
	//.
	dsr Register

	// Display Data Register. Its value is output as a character to every listener.
	ddr Register

	// Listeners. Each listener function is called every time the data register is written. Listener
	// functions must not block, fail, or panic. Really, the value should be written to a buffered
	// channel or otherwise asynchronously handle the event.
	list []func(uint16)
}

// Display status register bit fields for ready and interrupt enabled.
const (
	DisplayReady   = Register(1 << 15) // Ready
	DisplayEnabled = Register(1 << 14) // IE
)

func (d Display) device() string { return "CRT(PHOSPHOR)" }

func (d *Display) Init(_ *LC3, _ []Word) {
	d.Mutex = &sync.Mutex{}
	d.Lock()
	d.dsr = DisplayReady // Born ready.
	d.ddr = 0x2368       // ⍨
	d.Unlock()

	d.notify()
}

func (disp *Display) Write(data Register) {
	disp.Lock()
	defer disp.Unlock()

	disp.ddr = data
	disp.dsr &^= DisplayReady

	disp.notify()
}

func (disp *Display) DSR() Register {
	disp.Lock()
	defer disp.Unlock()

	return disp.dsr
}

func (disp *Display) Listen(listener func(uint16)) {
	disp.Lock()
	defer disp.Unlock()

	disp.list = append(disp.list, listener)
}

// notify wakes all listeners and tells them the good news: there is data to be seen!
func (disp *Display) notify() {
	for _, fn := range disp.list {
		fn(uint16(disp.ddr))
	}

	// After notifying all that care, the display is ready for more data. Despite the recommendation
	// to poll for device readiness, with few listeners and the device lock held, the program isn't
	// likely to see the device as not ready.
	disp.dsr |= DisplayReady
}

func (disp *Display) String() string {
	disp.Lock()
	defer disp.Unlock()

	return fmt.Sprintf("Display(status:%s,data:%s)", disp.dsr, disp.ddr)
}
