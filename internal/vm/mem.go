package vm

// mem.go contains the machine's memory controller.

import (
	"fmt"
)

// Memory is where we keep our most precious things: programs and data.
//
// The LC-3 has nearly unlimited memory: 128 kilobytes in a 16-bit address space of 2-byte words.
// The addressable memory space is divided into separate address spaces.
//
//   - system space for operating system code and data
//
//   - user space for unprivileged programs' code and data
//
//   - an I/O page for memory-mapped device-registers
//
// The memory controller, or MMU, mediates access to the address spaces from the CPU.
//
// ## Data Flow ##
//
// The MMU is responsible for translating logical addresses in the memory space to physical memory
// residing in RAM, CPU registers, or memory on external devices.
//
// To read or write to memory, the CPU puts the address into the address register (MAR) and the data
// into the data register (MDR) and either calls Fetch or Store; the controller will read from the
// address into its data register or write to memory from MDR, respectively.
//
// Admittedly, this is a strange design, at least from a software design perspective. We could
// simply use function arguments and return values to pass values instead. However, we use registers
// here in order to reflect the design of the LC-3 reference micro-architecture. For learning
// purposes, it helps to make the data flow explicit and model a separate MMU. (Indeed, the Go
// compiler will inline and optimize much of the code herein so data is kept in registers and on the
// stack.)
//
// ## Access Control ##
//
// The controller also enforces access control to each address space. The system space contains the
// code and data used for operating the machine and must only be accessed by privileged programs.
// When the address register contains an address in the system space (or, is for a privileged
// device) and the processor is running with user privileges, then memory access will raise an
// access control violation (ACV) exception and a fault handler is called.
//
// ## Data and Stacks ##
//
// The user and system spaces are further divided into regions. Primarily, each space contains a
// data region that includes global program data as well as the machine code for programs
// themselves.
//
// Temporary program data is stored on a stack: one for the system, the other for the user. The top
// of the current stack is pointed to by a stack pointer (SP, i.e. R6). The other stack is saved in
// a special-purpose register while it is inactive. That is, the system stack value is saved in SSP
// when running with user privileges; likewise, the user's in USP while with system privileges.
//
// Both stacks grow down; that is, when a word is pushed onto the stack, the address decreases and
// will point at the new data on the top of the stack.
//
// ## Interrupt Vector Tables ##
//
// In addition to system data and code, the system space includes small but important tables
// containing the addresses of service routines for I/O interrupts, traps, and exceptions. The
// system loads these tables with addresses of handlers and jumps to these handlers.
//
// ## Figure ##
//
// Since ASCII art is worth a thousand words:
//
// +========+========+=================+    +-----------------+   +-----------------+
// |        | 0xffff |  Memory-mapped  |    |                 |   |                 |
// |        |   ...  |     I/O page    |--->|                 |-->|DSR              |
// |        |   ...  |                 |--->|       MMIO      |-->|DDR    Device    |
// |        |   ...  |                 |--->|    registers    |-->|KBSR  registers  |
// |        | 0xfe00 |                 |--->|                 |-->|KBDR             |
// +========+========+=================+    |                 |   |                 |
// |        | 0xfdff |                 |    +-----------------+   +-----------------+
// |        |        |                 |             |   |
// |        |  ...   |   User stack    |             V   V
// |        |        |                 |    +-----------------+
// |        | 0x4568 |                 |<---|USP    MCR PSR   |
// |        +--------------------------+    |                 |
// |        | 0x4567 |                 |<---|R7 (RET)       R3|
// |  User  |  ...   |   User data     |    |                 |
// |  space | 0x3000 |                 |<---|R6 (SP)        R2|
// +========+========+=================+    |      CPU ⚙️     |
// |        | 0x2fff |                 |    |R5             R1|
// |        |        |  	       |    |                 |
// |        |   ...  |  System stack   |    |R4             R0|
// |        |        |                 |    |                 |
// |        | 0x2dad |                 |<---|SSP              |
// |        +--------+-----------------+    +-----------------+
// | System | 0x1234 |                 |
// | space  |  ...   |  System data    |
// |        | 0x0200 |                 |
// |        +--------------------------+
// |        | 0x01ff |    Interrupt    |
// |        |  ...   |  vector table   |
// |        | 0x0100 |                 |
// +        +--------+-----------------+
// |        | 0x00ff |      Trap       |
// |        |   ...  |   vector table  |
// |        | 0x0010 |                 |
// +        +--------+-----------------+
// |        | 0x000f |    Exception    |
// |        |   ...  |      table      |
// |        | 0x0000 |                 |
// +========+========+=================+

// Memory represents a memory controller that translates logical addresses to registers in the
// machine.
type Memory struct {
	MAR Register // Memory address register.
	MDR Register // Memory data register.

	// Physical memory in a virtual machine for an imaginary CPU.
	cell PhysicalMemory

	// Memory-mapped Devices registers.
	Devices MMIO

	log logger
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
			log:  defaultLogger(),
		},

		log: defaultLogger(),
	}

	return mem
}

// Fetch loads the data register from the address in the address register.
func (mem *Memory) Fetch() error {
	psr := mem.Devices.PSR()
	if psr&StatusPrivilege == StatusUser && mem.privileged() {
		mem.MDR = Register(psr)
		return &acv{interrupt{}}
	}

	err := mem.load(Word(mem.MAR), &mem.MDR)
	if err != nil {
		return fmt.Errorf("mem: fetch: %w", err)
	}

	return nil
}

// Store writes the word in the data register to the word in the address
// register.
func (mem *Memory) Store() error {
	psr := mem.Devices.PSR()

	if psr.Privilege() == PrivilegeUser && mem.privileged() {
		mem.MDR = Register(psr)

		return &acv{
			interrupt{},
		}
	}

	err := mem.store(Word(mem.MAR), Word(mem.MDR))
	if err != nil {
		return fmt.Errorf("mem: store: %w", err)
	}

	return nil
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

// acv is an memory access control violation exception.
type acv struct {
	interrupt
}

func (acv *acv) Error() string {
	return fmt.Sprintf("EXC: ACV (%s:%s)", acv.table, acv.vec)
}
