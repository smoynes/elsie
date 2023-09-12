package cpu

// devices.go includes code for device I/O. That includes only memory-mapped
// I/O.
//
// # Memory-mapped I/O #
//
// The memory controller redirects accesses of addresses in the I/O page to the
// MMIO controller. During boot, addresses are mapped to registers elsewhere in
// the CPU or provided by devices.
//
// Different kinds of devices have different types of registers, i.e., Register,
// StatusRegister, KeyboardRegister, etc. However, in Go, pointer types are
// unconvertible: we cannot convert from *StatusRegister to *Register, even
// though they have the same underlying type. So, we keep any pointers and type
// cast to the register types that the MMIO supports.

import (
	"fmt"
)

// MMIO is the memory-mapped I/O controller.
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

// Store writes a word to a memory-mapped I/O address.
func (mmio MMIO) Store(addr Word, mdr Register) error {
	var (
		devp any
		ok   bool
	)

	if devp, ok = mmio[addr]; !ok {
		panic("mmio no device")
	}

	switch dev := devp.(type) {
	case *ProcessorStatus:
		*dev = ProcessorStatus(mdr)
	case *Register:
		*dev = mdr
	default:
		panic(fmt.Sprintf("unexpected register type: %T", dev))
	}

	fmt.Printf("MMIO write addr: %s, word: %s\n", addr, devp.(fmt.Stringer))

	return nil
}

// Load fetches a word from a memory-mapped I/O address.
func (mmio MMIO) Load(addr Word, reg *Register) error {
	devp, ok := mmio[addr]
	if !ok {
		panic("mmio no device")
	}

	switch dev := devp.(type) {
	case *ProcessorStatus:
		*reg = Register(*dev)
	case *Register:
		*reg = Register(*dev)
	default:
		panic(fmt.Sprintf("unexpected register type: %T", dev))
	}

	fmt.Printf("MMIO fetch addr: %s, word: %s\n", addr, devp.(fmt.Stringer))

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
