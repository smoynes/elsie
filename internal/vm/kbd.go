package vm

import (
	"fmt"
	"math/rand"
	"sync"
)

// Keyboard is a hardwired input device for typos. It is its own driver.
type Keyboard struct {
	// mut provides mutual exclusion for the device. It might be interesting to contrast the lock
	// used here with the use of channels in the Display device.
	mut sync.Mutex

	// intr signals waiters when keyboard interrupts are enabled.
	intr *sync.Cond

	// Keyboard Status Register.
	KBSR Register

	// Keyboard Data Register.
	KBDR Register
}

// Bit fields for keyboard status flags.
const (
	KeyboardReady  = Register(1 << 15) // IR
	KeyboardEnable = Register(1 << 14) // IE
)

// NewKeyboard creates a new keyboard device and allocates resources
func NewKeyboard() *Keyboard {
	k := &Keyboard{
		mut:  sync.Mutex{},
		KBSR: 0x0000,
		KBDR: Register(a[rand.Intn(len(a))]),
	}
	k.intr = sync.NewCond(&k.mut)

	return k
}

// Init configures the keyboard device for use. It registers the device with the interrupt
// controller and enables interrupts.
func (k *Keyboard) Init(vm *LC3, _ []Word) {
	isr := ISR{vector: 0xff, driver: k}
	vm.INT.Register(PriorityNormal, isr)

	k.mut.Lock()
	k.KBSR = ^KeyboardReady | KeyboardEnable // Enable interrupts, clear ready flag.
	k.KBDR = Register(a[rand.Intn(len(a))])
	k.mut.Unlock()

	k.intr.Broadcast()
}

// InterruptRequested returns true if the keyboard has requested interrupt and interrupts are
// enabled. That is, both the R and IE bits are set in the status register.
func (k *Keyboard) InterruptRequested() bool {
	k.mut.Lock()
	defer k.mut.Unlock()

	return k.KBSR&(KeyboardEnable|KeyboardReady) == KeyboardEnable|KeyboardReady
}

// Read returns the value of a keyboard's register. If the data register is read then the ready flag
// is cleared.
func (k *Keyboard) Read(addr Word) (Word, error) {
	k.mut.Lock()
	defer k.mut.Unlock()

	if addr == KBSRAddr {
		return Word(k.KBSR), nil
	}

	val := Word(k.KBDR)
	k.KBDR = 0x0000
	k.KBSR = KeyboardReady | KeyboardEnable // ??

	return val, nil
}

// Write updates the status keyboard status register.
func (k *Keyboard) Write(addr Word, val Register) error {
	if addr != KBSRAddr {
		return fmt.Errorf("kbd: %w: %s", ErrNoDevice, addr)
	}

	k.mut.Lock()
	defer k.mut.Unlock()

	enabled := (k.KBSR&KeyboardEnable == 0) && (val&KeyboardEnable != 0)
	k.KBSR = val

	if enabled && (k.KBSR&KeyboardReady != 0) {
		k.intr.Signal()
	}

	return nil
}

// Update blocks until the keyboard interrupt is enabled and atomically sets the data and ready
// flag.
func (k *Keyboard) Update(key uint16) {
	k.mut.Lock()
	defer k.mut.Unlock()

	for !(k.KBSR&(KeyboardEnable & ^KeyboardReady) != 0) {
		k.intr.Wait()
	}

	k.KBDR = Register(key)
	k.KBSR |= KeyboardReady // Data is ready.
	k.intr.Signal()
}

func (k *Keyboard) String() string {
	k.mut.Lock()
	defer k.mut.Unlock()

	return fmt.Sprintf("Keyboard(status:%s,data:%s)", k.KBSR, k.KBDR)
}

func (*Keyboard) device() string { return "Keyboard(ModelM)" } // Simply the best.

var a = []rune{
	0x2361, 0x2362, 0x2363, 0x2364, 0x2365, 0x2368, 0x2369,
}
