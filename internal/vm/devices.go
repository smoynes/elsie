package vm

// devices.go has devices and their drivers.

import (
	"fmt"
	"sync"
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

const (
	KeyboardReady  = Register(1 << 15)
	KeyboardEnable = Register(1 << 14)
)

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	lock *sync.RWMutex
	intr *sync.Cond

	KBSR, KBDR Register
}

func NewKeyboard() *Keyboard {
	lock := new(sync.RWMutex)

	return &Keyboard{
		lock: lock,
		intr: sync.NewCond(lock),

		KBSR: 0x0000,
		KBDR: '?',
	}
}

func (k Keyboard) String() string {
	k.lock.RLock()
	defer k.lock.RUnlock()

	return fmt.Sprintf("Keyboard(status:%s,data:%s)", k.KBSR, k.KBDR)
}

// Wait blocks the caller until interrupts are enabled.
func (k Keyboard) Wait() {
	k.lock.Lock()
	k.intr.Wait()
}

func (k Keyboard) Device() string { return "Keyboard(ModelM)" } // Simply the best.

func (k *Keyboard) Init(vm *LC3, _ []Word) {
	k.KBSR = ^KeyboardReady | KeyboardEnable // Enable interrupts, clear ready flag.
	k.KBDR = 0x0000

	isr := ISR{vector: 0xff, driver: k}
	vm.INT.Register(PriorityNormal, isr)
}

// InterruptRequested returns true if the keyboard has requested interrupt and interrupts are
// enabled. That is, both the R and IE bits are set in the status register.
func (k *Keyboard) InterruptRequested() bool {
	k.lock.RLock()
	defer k.lock.RUnlock()

	return k.KBSR == (KeyboardEnable | KeyboardReady)
}

// Read returns the value of a keyboard's register. If the data register is read then the ready flag
// is cleared.
func (k *Keyboard) Read(addr Word) (Word, error) {
	if addr == KBSRAddr {
		k.lock.RLock()
		defer k.lock.RUnlock()

		return Word(k.KBSR), nil
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	wasDisabled := k.KBSR&KeyboardEnable == KeyboardEnable
	val := Word(k.KBDR)
	k.KBSR = (KeyboardEnable & ^KeyboardReady)
	k.KBDR = 0x0000 // ??

	if wasDisabled {
		k.intr.Broadcast()
	}

	return val, nil
}

// Write updates the status keyboard status register.
func (k *Keyboard) Write(addr Word, val Register) error {
	if addr != KBSRAddr {
		return fmt.Errorf("kbd: %w: %s", ErrNoDevice, addr)
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	enabled := (k.KBSR & ^KeyboardEnable != 0) && (val&KeyboardEnable != 0)
	k.KBSR = val

	if enabled {
		k.intr.Broadcast()
	}

	return nil
}

// Update blocks until the keyboard has interrupts enabled and atomically sets the data and ready
// flag.
func (k *Keyboard) Update(key uint16) {
	k.intr.L.Lock()

	for !(k.KBSR&(KeyboardEnable & ^KeyboardReady) != 0) {
		k.intr.Wait()
	}

	k.KBDR = Register(key)
	k.KBSR |= KeyboardReady // Data is ready.
	k.intr.Broadcast()
	k.intr.L.Unlock()
}
