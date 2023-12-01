package vm

import (
	"fmt"
	"sync"
)

// Display is a logical device for outputting characters. It has a status register (DSR) and a data
// register (DDR). A device driver controls the display and exposes output operations to the rest of
// the machine.
//
// When the machine writes to the display, the device automatically clears its interrupt-ready
// status-flag to indicate the display buffer is full. The driver must set the flag after the
// character is completely displayed. (Imagine there is a vsync or something the driver waits for.)
// In the meantime, if the machine writes to the display before the device sets the ready flag, data
// is overwritten and precious data could be lost. Don't do it! Instead, programs should poll until
// the device is ready.
type Display struct {
	// Display Status Register. The top two bits encode the interrupt-ready and enable flags. When
	// IR is set, the display can receive another character for output.
	//
	// 	| IR | IE |       |
	// 	+----+----+-------+
	// 	| 15 | 14 |13    0|
	dsr Register

	// Display Data Register
	ddr Register
}

// NewDisplay creates a display.
func NewDisplay() *Display {
	return &Display{}
}

// Display status-register bit-fields for ready and interrupt-enabled flags.
const (
	DisplayReady   = Register(1 << 15) // Ready
	DisplayEnabled = Register(1 << 14) // IE
)

func (disp Display) device() string { return "CRT(PHOSPHOR)" }

// Init initializes the device.
func (disp *Display) Init(_ *LC3, _ []Word) {
	disp.dsr = DisplayReady // Born ready.
	disp.ddr = 0x2368       // ⍨
}

// Write updates the display data register with the given data. It clears the ready flag and the
// driver should set it after the data is successfully displayed.
func (disp *Display) Write(data Register) {
	// Clear ready flag.
	disp.dsr &^= DisplayReady
	disp.ddr = data
}

// Read returns the value of the display data register.
func (disp *Display) Read() Register {
	return disp.ddr
}

// DSR returns the value of the display status register.
func (disp *Display) DSR() Register {
	return disp.dsr
}

// Status updates the value of the display status register and returns the previous value.
func (disp *Display) SetDSR(val Register) Register {
	prev := disp.dsr
	disp.dsr = val

	return prev
}

func (disp *Display) String() string {
	return fmt.Sprintf("Display(status:%s,data:%s)", disp.dsr, disp.ddr)
}

// DisplayDriver is a driver for an extremely simple terminal display. It ensures mutually exclusive
// access to the display hardware and maintains a list of listeners that are notified whenever a
// character is output to the device. A listener should display the data from the virtual device to
// the physical user.
type DisplayDriver struct {
	handle DeviceHandle[*Display, Display]

	// Addresses to which the registers are mapped.
	statusAddr Word
	dataAddr   Word

	// Mut provides mutually exclusive R/W access to the device registers.
	mut *sync.Mutex

	// Listeners. Each listener function is called every time the data register is written. Listener
	// functions must not block, fail, or panic. The value should be written to a buffered channel
	// or be otherwise asynchronously handled.
	list []func(uint16)
}

// NewDisplayDriver creates a new driver for the display and allocates resources. The driver has
// ownership of the device.
func NewDisplayDriver(display *Display) *DisplayDriver {
	return &DisplayDriver{
		handle:     NewDeviceHandle(display),
		statusAddr: DSRAddr,
		dataAddr:   DDRAddr,
		mut:        new(sync.Mutex),
		list:       nil,
	}
}

// Init initializes the display and the driver.
func (driver *DisplayDriver) Init(vm *LC3, addrs []Word) {
	if driver.mut == nil {
		panic("uninitialized lock")
	}

	driver.mut.Lock()
	defer driver.mut.Unlock()

	driver.statusAddr = addrs[0]
	driver.dataAddr = addrs[1]

	driver.handle.Init(vm, addrs)
}

// Read gets the status of the display device. Reading any other address returns an error.
func (driver *DisplayDriver) Read(addr Word) (Word, error) {
	if addr == driver.statusAddr {
		driver.mut.Lock()
		defer driver.mut.Unlock()

		return Word(driver.handle.device.DSR()), nil
	} else if addr == driver.dataAddr {
		driver.mut.Lock()
		defer driver.mut.Unlock()

		return Word(driver.handle.device.Read()), nil
	} else {
		return Word(0xdea1), fmt.Errorf("read: %w: %s:%s", ErrNoDevice, addr, driver)
	}
}

// InterruptRequested returns true when the display raises an interrupt request. For our purposes,
// the display never interrupts the CPU.
func (driver *DisplayDriver) InterruptRequested() bool {
	driver.mut.Lock()
	defer driver.mut.Unlock()

	return driver.handle.device != nil &&
		driver.handle.device.DSR() == (DisplayReady|DisplayEnabled)
}

// Write sets the data or status registers of the display device. When the data register is written,
// listeners are asynchronously notified.
func (driver *DisplayDriver) Write(addr Word, value Register) error {
	driver.mut.Lock()
	defer driver.mut.Unlock()

	if addr == driver.dataAddr {
		return driver.write(value)
	} else if addr == driver.statusAddr {
		driver.handle.device.SetDSR(value)
		return nil
	} else {
		return fmt.Errorf("write: %w: %s:%s", ErrNoDevice, addr, driver)
	}
}

// Listen adds a display listener. Each time a character is displayed, all listeners are called
// sequentially.
func (driver *DisplayDriver) Listen(listener func(uint16)) {
	driver.list = append(driver.list, listener)
}

// write writes the value to the display device and asynchronously notifies the listeners of the
// good news: there is data to be seen!
func (driver *DisplayDriver) write(value Register) error {
	device := driver.handle.device
	device.Write(value)

	// Asynchronously notify listeners of the write.
	go func() {
		for _, fn := range driver.list {
			fn(uint16(value))
		}

		// After all notifications, we can set the ready flag.
		driver.mut.Lock()
		dsr := device.DSR()
		device.SetDSR(dsr | DisplayReady)
		driver.mut.Unlock()
	}()

	return nil
}

func (driver *DisplayDriver) String() string {
	driver.mut.Lock()
	defer driver.mut.Unlock()

	if driver.handle.device == nil {
		return "DisplayDriver(display:nil)"
	}

	return fmt.Sprintf("DisplayDriver(display:%s)", driver.handle.device)
}

func (driver *DisplayDriver) device() string {
	driver.mut.Lock()
	defer driver.mut.Unlock()

	if driver.handle.device != nil {
		return driver.handle.device.device()
	}

	return "DISP(DRIVER)"
}
