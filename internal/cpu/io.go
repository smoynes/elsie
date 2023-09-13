package cpu

// io.go includes code for memory mapped I/O.
//
// The memory controller redirects accesses of addresses in the I/O page to the
// MMIO controller. During boot, addresses are mapped to registers from elsewhere in
// the CPU or external mdevices.
//
// Different kinds of devices have different types of registers, i.e., Register,
// StatusRegister, KeyboardRegister, etc. However, in Go, pointer types are
// unconvertible: we cannot convert from *StatusRegister to *Register, even
// though they have the same underlying type. So, we keep any pointers and type
// cast to the register types that the MMIO supports.

import (
	"errors"
	"fmt"
	"log"
)

// MMIO is the memory-mapped I/O controller. It holds a pointers to registers in
// a table indexed by logical address.
type MMIO map[Word]any

// Map attaches device registers to an address in the I/O page.
func (mmio *MMIO) Map(devices MMIO) error {
	for addr, regp := range devices {
		switch reg := regp.(type) {
		case *ProcessorStatus, *Register:
			(*mmio)[addr] = reg
		default:
			return fmt.Errorf("unsupported register type: %T", regp)
		}
	}

	return nil
}

var errMMIO = errors.New("mmio")

// Store writes a word to a memory-mapped I/O address.
func (mmio MMIO) Store(addr Word, mdr Register) error {
	var (
		devp any
		ok   bool
	)

	if devp, ok = mmio[addr]; !ok {
		return fmt.Errorf("%w: %s: no mmio device", errMMIO, addr)
	}

	switch dev := devp.(type) {
	case *ProcessorStatus:
		*dev = ProcessorStatus(mdr)
	case *Register:
		*dev = mdr
	default:
		panic(fmt.Sprintf("unexpected register type: %T", dev))
	}

	log.Printf("MMIO write addr: %s, word: %s\n", addr, devp.(fmt.Stringer))

	return nil
}

// Load fetches a word from a memory-mapped I/O address.
func (mmio MMIO) Load(addr Word, reg *Register) error {
	devp, ok := mmio[addr]
	if !ok {
		return fmt.Errorf("%w: %s: no mmio device", errMMIO, addr)
	}

	switch dev := devp.(type) {
	case *ProcessorStatus:
		*reg = Register(*dev)
	case *Register:
		*reg = Register(*dev)
	default:
		panic(fmt.Sprintf("unexpected register type: %T", dev))
	}

	log.Printf("MMIO fetch addr: %s, word: %s\n", addr, devp.(fmt.Stringer))

	return nil
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
