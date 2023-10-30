package vm

// io.go includes code for memory mapped I/O.

import (
	"errors"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// MMIO is the memory-mapped I/O controller. It holds a table indexed by logical address and points
// to either a register or a device driver that will perform the actual data exechange.
//
// Different kinds of devices have different types of registers, i.e., Register, StatusRegister,
// KeyboardRegister, etc. However, in Go, pointer types are unconvertible: we cannot convert from
// *StatusRegister to *Register, even though they have the same underlying type. So, we keep any
// pointers and type cast to the register types that the MMIO supports.
type MMIO struct {
	devs map[Word]any
	log  *log.Logger
}

// NewMMIO creates a memory-mapped I/O controller with default configuration.
func NewMMIO() *MMIO {
	m := MMIO{
		devs: make(map[Word]any),
		log:  log.DefaultLogger(),
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

	if dev == nil {
		return fmt.Errorf("%w: write: addr: %s %v", ErrNoDevice, addr, mmio.devs)
	} else if reg, ok := dev.(RegisterDevice); ok && reg != nil {
		reg.Put(mdr)
	} else if driver, ok := dev.(WriteDriver); ok && driver != nil {
		err := driver.Write(addr, mdr)
		if err != nil {
			return fmt.Errorf("mmio: write: %s:%s: %w", addr, dev, err)
		}
	} else {
		mmio.log.Error("%s: addr: %s: %T", ErrNoDevice, addr, dev)
		panic(ErrNoDevice.Error())
	}

	mmio.log.Debug("stored", log.String("ADDR", addr.String()), log.String("DATA", mdr.String()))

	return nil
}

// Load fetches a word from a memory-mapped I/O address.
func (mmio MMIO) Load(addr Word) (Register, error) {
	var value Word

	dev := mmio.devs[addr]

	if dev == nil {
		return Register(0xffff), fmt.Errorf("%w: write: addr: %s", ErrNoDevice, addr)
	} else if reg, ok := dev.(RegisterDevice); ok {
		value = Word(reg.Get())
	} else if driver, ok := dev.(ReadDriver); ok {
		var err error
		value, err = driver.Read(addr)

		if err != nil {
			return Register(0xffff), fmt.Errorf("mmio: write: %s:%s: %w", addr, dev, err)
		}
	} else {
		mmio.log.Error("%s: addr: %s: %T", ErrNoDevice, addr, dev)
		panic(ErrNoDevice)
	}

	mmio.log.Debug("loaded", log.String("ADDR", addr.String()), log.String("DATA", value.String()))

	return Register(value), nil
}

// Map configures the memory mapping for device I/O. Keys in the map are addresses and values are
// device drivers or registers.
func (mmio *MMIO) Map(devices map[Word]any) error {
	for addr, dev := range devices {
		if dev == nil {
			return fmt.Errorf("%w: map: bad device: %s, %T", errMMIO, addr, dev)
		} else if dd, ok := dev.(Device); ok && dd != nil {
			mmio.log.Debug("mapped device",
				log.String("ADDR", addr.String()),
				log.String("DEVICE", dd.device()),
			)
		} else {
			mmio.log.Error("mmio: map: unsupported device: %s %T %#v", dev, dev, dev)
			return fmt.Errorf("%w: map: unsupported device: %s %T", errMMIO, addr, dev)
		}
	}

	// Add new mappings only if all devices are valid.
	for addr, dev := range devices {
		mmio.devs[addr] = dev
	}

	return nil
}

func (mmio MMIO) Get(addr Word) any {
	return mmio.devs[addr]
}

// PSR returns the value of the status register, if it has been mapped.
func (mmio MMIO) PSR() ProcessorStatus {
	var psr ProcessorStatus

	if dev := mmio.devs[PSRAddr]; dev != nil {
		psr = *(dev.(*ProcessorStatus)) // \\((* * \\)) - praise punctuation!
	}

	return psr
}

// DDR returns the value of the display data register, if it has been mapped.
func (mmio MMIO) DDR() rune {
	if dev := mmio.devs[DDRAddr]; dev != nil {
		val := dev.(*DisplayDriver)
		ddr := val.handle.device.ddr

		return rune(ddr)
	}

	return '⍝'
}
