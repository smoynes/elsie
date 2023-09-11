package cpu

// Memory is where we keep our most precious things: programs and data.
//
// The LC-3 has nearly unlimited memory: 128 kilobytes in a 16-bit address space
// of 2-byte words. Within this addressable memory space there is:
//
//   - system space for operating system code and data
//   - user space for unprivileged programs
//   - I/O page for memory-mapped device-registers
//
// ## Usage ##
//
// To read or write to memory, the CPU puts the address into the address
// register and either calls Fetch which reads the value and puts it in the data
// register, or it puts a value in the data register and calls Store, which
// writes the value to the address.
//
// Admittedly, this is a strange design, at least from a software perspective.
// We could simply use function arguments and return values to pass values
// instead. However, registers are used here to reflect the design of the LC-3
// reference architecture and make the clock cycles visible in the code
// structure.
//
// ## Access Control ##
//
// ACV
//
// ## Stacks ##
//
// Each space is further divided into regions.
//
// The top of the current stack is pointed to by the stack pointer (register SP,
// i.e. R6). The inactive stack is saved in another register until the privilege
// level changes. For example, the system stack value is saved in SSP when
// running with user privileges; conversely, USP with system privileges.
//
// Both stacks grow down; that is, when a word is pushed onto the stack, the
// address decreases and will point at the new data on the top of the stack.
// (Note, however, that the stacks grow typographically upwards in the diagram
// below.)
//
// ## Memory-mapped I/O ##
//
// ## Interrupt and Trap Vectors ##
//
// Since ASCII art is worth a thousand words:
//
// +========+========+=================+
// |        | 0x0100 |   Exception     |
// |        |   ...  |  vector table   |--+
// |        | 0x01ff |                 |  |
// |        +--------------------------+
// |        | 0x0??? |   Trap vector   |
// |        |   ...  |      table      |--+
// |        | 0x0??? |                 |  |
// |        +--------------------------+
// |        | 0x0??? |    Interrupt    |
// |        |   ...  |  vector table   |--+
// |        | 0x0??? |                 |  |
// |        +--------------------------+  |
// | System | 0x1000 |     System      |<-+ ISR
// | space  |        |      data       |
// |        +--------+-----------------+   +-----------------+
// |        | 0x2ff0 |                 |<--|SSP              |
// |        |   ...  |  System stack   |   |                 |
// |        | 0x2fff |                 |   |                 |
// +========+========+=================+   |      CPU ⚙️     |
// |        | 0x3000 |                 |   |                 |
// |        |        |                 |<--|RET(R7)          |
// |        |  ...   |   User data     |   |                 |
// |        |        |                 |   |                 |
// |        |        |                 |<=>|MDR              |
// |  User  | 0xfdef |                 |<==|MDR              |
// + space  +--------+-----------------+   |                 |
// |        | 0xfdf0 |                 |<--|SP(R6)           |
// |        |        |                 |   |                 |
// |        |   ...  |   User stack    |<--|USP   MCR PSR    |
// |        |        |                 |   +-----------------+
// |        |        |                 |           ^
// |        | 0xfdff |                 |  +--------+
// +========+========+=================+  | +-----------------+
// |        | 0xfe00 |                 |--+ |                 |
// | I/O    |        |  Memory-mapped  |--->|DSR   Device     |
// | page   |   ...  |     I/O page    |--->|DDR   registers  |
// |        |        |                 |--->|KBSR             |
// |        | 0xffff |                 |--->|KBDR             |
// +========+========+=================+    +-----------------+

import (
	"fmt"
	"math"
)

// Memory represents a memory controller for logical addresses.
type Memory struct {
	MAR Register // Memory address register.
	MDR Register // Memory data register.

	PSR *ProcessorStatus // CPU status register.

	// Physical memory in a virtual machine for an imaginary CPU.
	cell PhysicalMemory

	// Memory-mapped device registers.
	device MMIO
}

// Logical memory address space.
const (
	MaxAddress  Word = math.MaxUint16
	AdressSpace int  = int(MaxAddress) + 1
)

// PhysicalMemory is . The top of the address space is reserved for memory-mapped
// I/O.
type PhysicalMemory [MaxAddress & IOPageAddr]Word

func NewMemory(psr *ProcessorStatus) Memory {
	mem := Memory{
		MAR: 0xffff,
		MDR: 0x0000,
		PSR: psr,

		device: make(MMIO),
		cell:   PhysicalMemory{},
	}

	return mem
}

// Fetch loads the data register from the address in the address register.
func (mem *Memory) Fetch() error {
	if mem.PSR.Privilege() == PrivilegeUser && mem.privileged() {
		mem.MDR = Register(*mem.PSR)
		return &acv{interrupt{}}
	}
	cell := mem.load(Word(mem.MAR))
	mem.MDR = Register(cell)

	return nil
}

// Privileged returns true if the address in MAR requires privileges to access.
func (mem *Memory) privileged() bool {
	return (Word(mem.MAR) < UserSpaceAddr ||
		Word(mem.MAR) == MCRAddr ||
		Word(mem.MDR) == PSRAddr)
}

// Store writes the word in the data register to the word in the address
// register.
func (mem *Memory) Store() error {
	if mem.PSR.Privilege() == PrivilegeUser && mem.privileged() {
		mem.MDR = Register(*mem.PSR)
		return &acv{
			interrupt{},
		}
	}

	mem.cell[mem.MAR] = Word(mem.MDR)

	return nil
}

// Loads a word from a memory directly without using the address and data
// registers.
func (mem *Memory) load(addr Word) Word {
	if addr >= IOPageAddr {
		panic("bad addr")
	}
	return mem.cell[addr]
}

// Stores a word into memory directly without using the address and data
// registers.
func (mem *Memory) store(addr Word, cell Word) {
	if addr >= IOPageAddr {
		panic("bad addr")
	}
	mem.cell[addr] = cell
}

// Map attaches a device register to an address in the I/O page.
func (mem *Memory) Map(devices MMIO) {
	for addr, reg := range devices {
		mem.device[addr] = reg
	}
}

type MMIO map[Word]*Register

// Address space regions.
const (
	UserSpaceAddr Word = 0x3000 // Start of user address space.
	IOPageAddr    Word = 0xfe00 // I/O page address space.
)

// Addresses of memory-mapped device registers.
const (
	KBSRAddr Word = 0xfe00 // Keyboard status and data registers.
	KBDRAddr Word = 0xfe02
	DSRAddr  Word = 0xfe04 // Display status and data registers.
	DDRAddr  Word = 0xfe06
	PSRAddr  Word = 0xfffc // Processor status register. Privileged.
	MCRAddr  Word = 0xfffe // Machine control register. Privileged.
)

// Exception vector table and defined vectors in the table.
const (
	ExceptionTable = Word(0x0100)
	ExceptionPMV   = Word(0x00)
	ExceptionXOP   = Word(0x01)
	ExceptionACV   = Word(0x02)
)

// Trap handler table and defined vectors in the table.
const (
	TrapTable = Word(0x0000)
	TrapHALT  = Word(0x0025)
)

type acv struct {
	interrupt
}

func (acv *acv) Error() string {
	return fmt.Sprintf("INT: ACV (%s:%s)", acv.table, acv.vec)
}
