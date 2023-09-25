package vm

import (
	"context"
	"fmt"
	"sync"
)

// Display is a logical device for outputting characters. It has a status register (DSR) and a data
// register (DDR). A device driver controls the device and exposes output operations to the rest of
// the machine.
//
// When the machine writes to the display, the device automatically clears its interrupt-ready
// status-flag to indicate the display buffer is full. Once the device completely outputs the
// character, it sets the ready flag again. In the meantime, if the machine writes to the display
// before the device clears the buffer, data is overwritten and precious data is lost. Don't do it!
// Instead, programs should poll until the device is ready.
type Display struct {
	// mut provides mutually exclusive R/W access to the device registers.
	mut *sync.Mutex

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

// NewDisplay creates a display and allocates its resources.
func NewDisplay() *Display {
	return &Display{
		mut: new(sync.Mutex),
	}
}

// Display status register bit fields for ready and interrupt enabled.
const (
	DisplayReady   = Register(1 << 15) // Ready
	DisplayEnabled = Register(1 << 14) // IE
)

func (d Display) device() string { return "CRT(PHOSPHOR)" }

func (d *Display) Init(_ *LC3, _ []Word) {
	if d.mut == nil {
		panic("lock uninitialized")
	}

	d.mut.Lock()
	d.dsr = DisplayReady // Born ready.
	d.ddr = 0x2368       // ⍨
	d.mut.Unlock()

	d.notify()
}

// Write updates the display data register with the given data. It (briefly) clears the ready
// status-flag until it notifies all listeners with the data. Then the ready flag is set again.
func (disp *Display) Write(data Register) {
	disp.mut.Lock()
	defer disp.mut.Unlock()

	disp.ddr = data
	disp.dsr &^= DisplayReady

	disp.notify()
}

// DSR returns the value of the display status register.
func (disp *Display) DSR() Register {
	// Locking here is dubious.
	disp.mut.Lock()
	defer disp.mut.Unlock()

	return disp.dsr
}

// Listen adds a display listener. Every listener function is called every time the display data is
// written.
func (disp *Display) Listen(listener func(uint16)) {
	disp.mut.Lock()
	defer disp.mut.Unlock()

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
	disp.mut.Lock()
	defer disp.mut.Unlock()

	return fmt.Sprintf("Display(status:%s,data:%s)", disp.dsr, disp.ddr)
}

// DisplayDriver is a driver for an extremely simple terminal display.
type DisplayDriver struct {
	handle DeviceHandle[*Display, Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word
}

// WithDisplayDriver creates a new display and its driver. It returns a cancellation function that
// will release display resources.
func WithDisplayDriver(parent context.Context) (context.Context, *DisplayDriver, context.CancelFunc) {
	var (
		display     = NewDisplay()
		driver      = NewDisplayDriver(display)
		ctx, cancel = context.WithCancel(parent)
	)

	return ctx, driver, cancel
}

// NewDisplayDriver creates a new driver for the display. The driver has ownership of the device.
func NewDisplayDriver(display *Display) *DisplayDriver {
	return &DisplayDriver{
		handle:     NewDeviceHandle(display),
		statusAddr: DSRAddr,
		dataAddr:   DDRAddr,
	}
}

// Init initializes the display and the driver.
func (driver *DisplayDriver) Init(vm *LC3, addrs []Word) {
	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	driver.handle.Init(vm, addrs)
}

// Read gets the status of the display device. Reading any other address returns an error.
func (driver *DisplayDriver) Read(addr Word) (Word, error) {
	if addr == driver.statusAddr {
		return Word(driver.handle.device.DSR()), nil
	}

	return Word(0xdea1), fmt.Errorf("read: %w: %s:%s", ErrNoDevice, addr, driver)
}

// InterruptRequested returns true when the display raises an interrupt request. For our purposes,
// the display never interrupts the CPU.
func (driver *DisplayDriver) InterruptRequested() bool {
	return driver.handle.device != nil &&
		driver.handle.device.DSR() == (DisplayReady|DisplayEnabled)
}

// Write sets the data register of the display device. Writing any other address returns an error.
func (driver *DisplayDriver) Write(addr Word, value Register) error {
	if addr == driver.dataAddr {
		driver.handle.device.Write(value)
		return nil
	}

	return fmt.Errorf("write: %w: %s:%s", ErrNoDevice, addr, driver)
}

func (driver *DisplayDriver) String() string {
	if driver.handle.device != nil {
		return fmt.Sprintf("DisplayDriver(display:%s)", driver.handle.device)
	} else {
		return "DisplayDriver(display:nil)"
	}
}

func (driver *DisplayDriver) device() string {
	if driver.handle.device != nil {
		return driver.handle.device.device()
	}

	return "DISP(DRIVER)"
}
