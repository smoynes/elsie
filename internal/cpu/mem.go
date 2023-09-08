package cpu

import (
	"math"
)

// Memory is where we keep our most precious things: programs and data.
//
// The LC-3 has nearly unlimited memory: 128 kilobytes in a 16-bit address space
// of 2-byte words. Within addressable memory we have:
//
//   - system space for operating system code and data
//   - user space for unprivileged programs
//   - I/O page for memory-mapped device-registers
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
// Importantly, the memory control logic in [Memory] does not enforce access
// control -- each CPU instruction checks the privilege level in the status
// register (PSR) before loading or storing words.
//
// TODO: interrupt vector table
// TODO: memory-mapped I/O
//
// Since ASCII art is worth a thousand words:
// +========+========+=================+
// |        | 0x0000 |    Interrupt    |
// |        |   ...  |  vector table   |--+
// |        | 0x00ff |                 |  |
// +        +--------+-----------------+  |
// | System | 0x1000 |     System      |<-+ ISR
// | space  |        |      data       |
// +        +--------+-----------------+   +-----------------+
// |        | 0x2ff0 |                 |<--|SSP              |
// |        |   ...  |  System stack   |   |                 |
// |        | 0x2fff |                 |   |                 |
// +========+========+=================+   |      CPU ⚙️     |
// |        | 0x3000 |                 |   |                 |
// |        |        |                 |<--|RET(R7)          |
// |        |  ...   |   User data     |   |                 |
// |        |        |                 |   |PSR              |
// |        |        |                 |<=>|MDR              |
// |  User  | 0xfdef |                 |<==|MDR              |
// + space  +--------+-----------------+   |                 |
// |        | 0xfdf0 |                 |<--|SP(R6)           |
// |        |        |                 |   |                 |
// |        |   ...  |   User stack    |<--|USP   MCR        |
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
// .
type Memory struct {
	MAR  Register // Memory address register.
	MDR  Register // Memory data register.
	cell [AddressSpace]Word
}

// Size of addressable memory: 2 ^^ 16
const AddressSpace = math.MaxUint16

// Fetch loads the cell pointed at by the address register and puts the result
// in the data register.
//
// Admittedly, this is a strange design. We use registers to mimic the design of
// the LC-3 reference architecture and make clock cycles visible in the code structure.
func (mem *Memory) Fetch() {
	addr := Word(mem.MAR)
	cell := mem.load(addr)
	mem.MDR = Register(cell)
}

// Load loads a word from a memory directly, without simulated clock cycles.
func (mem *Memory) load(addr Word) Word {
	return mem.cell[addr]
}

func (mem *Memory) Store() {
	mem.cell[mem.MAR] = Word(mem.MDR)
}

// Store stores a word into memory directly, without simulated clock cycles.
func (mem *Memory) store(addr Word, cell Word) {
	mem.cell[addr] = cell
}
