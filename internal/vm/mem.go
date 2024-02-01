package vm

// mem.go contains the machine's memory controller.

import (
	"errors"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// Memory represents a controller that translates logical addresses to registers, memory or devices
// in the machine. The microarchitecture's data path uses two control registers: the address
// register (MAR) and a data register. Together, these register mediate most access to memory and
// simplifies access control.
type Memory struct {
	// Memory address register.
	MAR Register

	// Memory data register.
	MDR Register

	// Physical memory in a virtual machine for an imaginary CPU.
	cell PhysicalMemory

	// Memory-mapped device registers.
	Devices MMIO

	log *log.Logger
}

// Regions of address space. Each region begins at the address and grows upwards towards the next.
const (
	ServiceRoutineAddr Word = 0x0000
	SystemSpaceAddr    Word = 0x0200
	UserSpaceAddr      Word = 0x3000
	IOPageAddr         Word = 0xfe00
	AddrSpace          Word = 0xffff // Logical address space; 65_536 addressable words.
)

// PhysicalMemory is (virtualized) physical memory. The top of the logical address space is reserved
// for the I/O page so the backing buffer is slightly smaller than the full logical address space.
type PhysicalMemory [AddrSpace & IOPageAddr]Word

// NewMemory initializes a memory controller.
func NewMemory(psr *ProcessorStatus) Memory {
	mem := Memory{
		MAR: 0xffff,
		MDR: 0x0ff0,

		cell: PhysicalMemory{},
		Devices: MMIO{
			devs: make(map[Word]any),
			log:  log.DefaultLogger(),
		},

		log: log.DefaultLogger(),
	}

	return mem
}

// Fetch loads the data register from the address in the address register.
func (mem *Memory) Fetch() error {
	psr := mem.Devices.PSR()

	memErr := &MemoryError{
		Addr: Word(mem.MAR),
	}

	if psr&StatusPrivilege == StatusUser && mem.privileged() {
		mem.MDR = Register(psr)

		return fmt.Errorf("%w: fetch: %w", memErr, ErrAccessControl)
	}

	err := mem.load(Word(mem.MAR), &mem.MDR)
	if err != nil {
		return fmt.Errorf("%w: fetch: %w", memErr, err)
	}

	return nil
}

// Store writes the word in the data register to the word in the address
// register.
func (mem *Memory) Store() error {
	psr := mem.Devices.PSR()

	if psr.Privilege() == PrivilegeUser && mem.privileged() {
		mem.MDR = Register(psr)
		return fmt.Errorf("%w: store: %w", ErrMemory, ErrAccessControl)
	}

	err := mem.store(Word(mem.MAR), Word(mem.MDR))
	if err != nil {
		return fmt.Errorf("%w: store: %w", ErrMemory, err)
	}

	return nil
}

// View returns a copy of the memory cells. It is intended as a debugging and
// development tool and is quite expensive computationally.
func (mem *Memory) View() PhysicalMemory {
	var view PhysicalMemory

	copy(view[:], mem.cell[:])

	return view
}

// Loads a word directly, without using the address and data registers.
func (mem *Memory) load(addr Word, reg *Register) error {
	if addr >= IOPageAddr {
		r, err := mem.Devices.Load(addr)
		*reg = r

		return err
	}

	*reg = Register(mem.cell[addr])

	return nil
}

// Stores a word into memory directly without using the address and data
// registers.
func (mem *Memory) store(addr Word, cell Word) error {
	if addr >= IOPageAddr {
		return mem.Devices.Store(addr, Register(cell))
	}

	mem.cell[addr] = cell

	return nil
}

// Privileged returns true if the address in MAR requires system privileges to access.
func (mem *Memory) privileged() bool {
	return (Word(mem.MAR) < UserSpaceAddr ||
		Word(mem.MAR) == MCRAddr ||
		Word(mem.MDR) == PSRAddr)
}

// MemeoryErrors are returned to provide the address if a wrapped ErrMemory.
type MemoryError struct {
	Addr Word
}

func (me *MemoryError) Error() string {
	return fmt.Sprintf("%s: %s", ErrMemory, me.Addr)
}

func (me *MemoryError) Is(err error) bool {
	if err == ErrMemory {
		return true
	} else if _, ok := err.(*MemoryError); ok {
		return true
	} else {
		return false
	}
}

var (
	ErrMemory        = errors.New("memory error")
	ErrAccessControl = errors.New("access control")
)
