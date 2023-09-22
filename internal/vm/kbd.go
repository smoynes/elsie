package vm

import (
	"fmt"
	"sync"
)

// Bit fields for keyboard status flags.
const (
	KeyboardReady  = Register(1 << 15) // IR
	KeyboardEnable = Register(1 << 14) // IE
)

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	sync.Mutex
	intr *sync.Cond

	KBSR, KBDR Register
}

func NewKeyboard() *Keyboard {
	k := &Keyboard{
		Mutex: sync.Mutex{},
		KBSR:  0x0000,
		KBDR:  '?',
	}
	k.intr = sync.NewCond(&k.Mutex)

	return k
}

func (k *Keyboard) String() string {
	k.Lock()
	defer k.Unlock()

	return fmt.Sprintf("Keyboard(status:%s,data:%s)", k.KBSR, k.KBDR)
}

// Wait blocks the caller until interrupts are enabled.
func (k *Keyboard) Wait() {
	k.Lock()
	defer k.Unlock()

	for !(k.KBSR&(KeyboardEnable & ^KeyboardReady) != 0) {
		k.intr.Wait()
	}
}

func (k *Keyboard) Device() string { return "Keyboard(ModelM)" } // Simply the best.

// Init configures the keyboard device for use. It registers the device with the interrupt
// controller and enables interrupts.
func (k *Keyboard) Init(vm *LC3, _ []Word) {
	isr := ISR{vector: 0xff, driver: k}
	vm.INT.Register(PriorityNormal, isr)

	k.Lock()
	k.KBSR = ^KeyboardReady | KeyboardEnable // Enable interrupts, clear ready flag.
	k.KBDR = 0x0000
	k.Unlock()

	k.intr.Broadcast()
}

// InterruptRequested returns true if the keyboard has requested interrupt and interrupts are
// enabled. That is, both the R and IE bits are set in the status register.
func (k *Keyboard) InterruptRequested() bool {
	k.Lock()
	defer k.Unlock()

	return k.KBSR == (KeyboardEnable | KeyboardReady)
}

// Read returns the value of a keyboard's register. If the data register is read then the ready flag
// is cleared.
func (k *Keyboard) Read(addr Word) (Word, error) {
	if addr == KBSRAddr {
		k.Lock()
		defer k.Unlock()

		return Word(k.KBSR), nil
	}

	k.Lock()
	defer k.Unlock()

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

	k.Lock()
	defer k.Unlock()

	enabled := (k.KBSR & ^KeyboardEnable != 0) && (val&KeyboardEnable != 0)
	k.KBSR = val

	if enabled {
		k.intr.Broadcast()
	}

	return nil
}

// Update blocks until the keyboard interrupt is enabled and atomically sets the data and ready
// flag.
func (k *Keyboard) Update(key uint16) {
	k.Lock()
	defer k.Unlock()

	for !(k.KBSR&(KeyboardEnable & ^KeyboardReady) != 0) {
		k.intr.Wait()
	}

	k.KBDR = Register(key)
	k.KBSR |= KeyboardReady // Data is ready.
	k.intr.Broadcast()
}
