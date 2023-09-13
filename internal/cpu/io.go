package cpu

// io.go includes code for memory mapped I/O.

import (
	"errors"
	"fmt"
	"log"
)

// The memory controller redirects accesses of addresses in the I/O page to the
// MMIO controller. During boot, addresses are mapped to registers in the CPU or
// external devices.
//
// Different kinds of devices have different types of registers, i.e., Register,
// StatusRegister, KeyboardRegister, etc. However, in Go, pointer types are
// unconvertible: we cannot convert from *StatusRegister to *Register, even
// though they have the same underlying type. So, we keep any pointers and type
// cast to the register types that the MMIO supports.

// MMIO is the memory-mapped I/O controller. It holds a table indexed by logical address and points
// to either to registers or a device driver.
type MMIO map[Word]any

// Addresses of memory-mapped device registers.
const (
	KBSRAddr Word = 0xfe00 // Keyboard status and data registers.
	KBDRAddr Word = 0xfe02
	DSRAddr  Word = 0xfe04 // Display status and data registers.
	DDRAddr  Word = 0xfe06
	PSRAddr  Word = 0xfffc // Processor status register. Privileged.
	MCRAddr  Word = 0xfffe // Machine control register. Privileged.
)

// errMMIO is a wrapped error.
var errMMIO = errors.New("mmio")

// Store writes a word to a memory-mapped I/O address.
func (mmio MMIO) Store(addr Word, mdr Register) error {
	var (
		dev any
		ok  bool
	)

	if dev, ok = mmio[addr]; !ok {
		return fmt.Errorf("%w: store: %s: no mmio device", errMMIO, addr)
	}

	switch dev := dev.(type) {
	case *Device:
		dev.Write(addr, DeviceRegister(mdr))
	case *ProcessorStatus:
		*dev = ProcessorStatus(mdr)
	case *Register:
		*dev = mdr
	default:
		panic(fmt.Sprintf("unexpected device: %s, type: %T", dev, dev))
	}

	log.Printf("MMIO write addr: %s, word: %s\n", addr, dev.(fmt.Stringer))

	return nil
}

// Load fetches a word from a memory-mapped I/O address.
func (mmio MMIO) Load(addr Word, reg *Register) error {
	log.Printf("%s: load: %+v", errMMIO, mmio)

	dev, ok := mmio[addr]
	if !ok {
		return fmt.Errorf("%w: load: %s: no mmio device", errMMIO, addr)
	}

	switch dev := dev.(type) {
	case *Device:
		var d DeviceRegister = dev.Read(addr)
		*reg = Register(d)
	case *ProcessorStatus:
		*reg = Register(*dev)
	case *Register:
		*reg = Register(*dev)
	default:
		panic(fmt.Sprintf("unexpected register type: %T", dev))
	}

	log.Printf("MMIO fetch addr: %s, word: %s\n", addr, reg)

	return nil
}

// Map attaches device registers to an address in the I/O page.
func (mmio MMIO) Map(devices MMIO) error {
	log.Printf("mmio: map: devices: %s", devices)

	for addr, dev := range devices {
		log.Printf("mmio: map: device: %T %s", dev, dev)
		switch dev.(type) {
		case *ProcessorStatus, *Register:
			log.Printf("mmio: map: device: %s", dev)
			mmio[addr] = dev
		case *Device:
			mmio[addr] = dev
		default:
			return fmt.Errorf("%w: map: unsupported device: %T", errMMIO, dev)
		}
	}

	return nil
}
