package vm

// io.go includes code for memory mapped I/O.

import (
	"errors"
	"fmt"
)

// The memory controller redirects accesses of addresses in the I/O page to the MMIO controller.
// During boot, addresses are mapped to registers in the CPU or external devices.
//
// Different kinds of devices have different types of registers, i.e., Register, StatusRegister,
// KeyboardRegister, etc. However, in Go, pointer types are unconvertible: we cannot convert from
// *StatusRegister to *Register, even though they have the same underlying type. So, we keep any
// pointers and type cast to the register types that the MMIO supports.

// MMIO is the memory-mapped I/O controller. It holds a table indexed by logical address and points
// to either a register or a device driver.
type MMIO struct {
	devs map[Word]any
	log  logger
}

// NewMMIO creates a memory-mapped I/O controller with default configuration.
func NewMMIO() *MMIO {
	m := MMIO{
		devs: make(map[Word]any),
		log:  defaultLogger(),
	}

	return &m
}

// Addresses of memory-mapped device registers.
const (
	KBSRAddr Word = 0xfe00 // Keyboard status and data registers.
	KBDRAddr Word = 0xfe02
	DSRAddr  Word = 0xfe04 // Display status and data registers.
	DDRAddr  Word = 0xfe06
	PSRAddr  Word = 0xfffc // Processor status register. Privileged.
	MCRAddr  Word = 0xfffe // Machine control register. Privileged.
)

var (
	errMMIO = errors.New("mmio")

	// ErrNoDevice is returned when reading or writing to an unmapped address.
	ErrNoDevice = fmt.Errorf("%w: no device", errMMIO)
)

// Store writes a word to a memory-mapped I/O address.
func (mmio MMIO) Store(addr Word, mdr Register) error {
	dev := mmio.devs[addr]

	switch dev := dev.(type) {
	case *Device:
		if err := dev.Write(addr, DeviceRegister(mdr)); err != nil {
			return fmt.Errorf("mmio: %w", err)
		}
	case *ProcessorStatus:
		*dev = ProcessorStatus(mdr)
	case *Register:
		*dev = mdr
	case nil:
		return fmt.Errorf("%w: addr: %s", ErrNoDevice, addr)
	default:
		mmio.log.Panicf("%s: addr: %s: %s", ErrNoDevice, addr, dev)
	}

	mmio.log.Printf("mmio: store: %s := %s\n", addr, mdr)

	return nil
}

// Load fetches a word from a memory-mapped I/O address.
func (mmio MMIO) Load(addr Word, reg *Register) error {
	dev := mmio.devs[addr]

	switch dev := dev.(type) {
	case *Device:
		d, err := dev.Read(addr)
		*reg = Register(d)

		if err != nil {
			return fmt.Errorf("io: %w", err)
		}
	case *ProcessorStatus:
		*reg = Register(*dev)
	case *Register:
		*reg = Register(*dev)
	case nil:
		mmio.log.Panicf("%s: addr: %s", ErrNoDevice, addr)
		//return fmt.Errorf("%w: addr: %s", ErrNoDevice, addr)
	default:
		mmio.log.Panicf("%s: addr: %s", ErrNoDevice, addr)
	}

	mmio.log.Printf("%s: load: %s", errMMIO, addr)

	return nil
}

// Map configures the memory mapping for device I/O. Keys in the map are addresses and values are
// devices or registers.
func (mmio MMIO) Map(devices map[Word]any) error {
	for addr, dev := range devices {
		dev := dev

		switch dev := dev.(type) {
		case *Device:
			mmio.log.Printf("mmio: map: addr: %s, device: %s", addr, dev)
			mmio.devs[addr] = dev
		case *ProcessorStatus, *Register:
			mmio.log.Printf("mmio: map: register: %s %#T, ", addr, dev)
			mmio.devs[addr] = dev
		default:
			return fmt.Errorf("%w: map: unsupported device: %T", errMMIO, dev)
		}
	}

	return nil
}

// PSR returns the value of the status register, if it has been mapped.
func (mmio MMIO) PSR() ProcessorStatus {
	var psr ProcessorStatus

	if dev := mmio.devs[PSRAddr]; dev != nil {
		psr = *(dev.(*ProcessorStatus)) // \\((* * \\)) - praise punctuation!
	}

	return psr
}
