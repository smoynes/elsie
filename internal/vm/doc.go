// Package vm implements a basic VM for executing LC3 machine code.
//
// The design of the virtual machine is meant to mimic or reflect the micro-architecture described
// in the text rather than being an efficient or exemplary VM.
//
// # CPU
//
// The machine's CPU is extraordinarily simple. It has just:
//
//   - a few registers: program counter, instuction register, processor status and control register.
//   - user- and system-stack pointer registers
//   - a file of eight general-purpose registers
//   - an device interrupt controller
//   - a memory controller
//
// # Memory
//
// Memory is where we keep our most precious things -- programs and data -- and the LC-3 has nearly
// unlimited memory: 128 kilobytes in a 16-bit address space of 2-byte words. The addressable memory
// space is divided into separate address spaces.
//
//   - system space for operating system code and data
//   - user space for unprivileged programs' code and data
//   - an I/O page for memory-mapped device-registers
//
// The memory controller (or MMU) mediates access to the address spaces from the CPU.
//
// # Data Flow
//
// The MMU translates logical addresses in the memory space into physical memory which might located
// in RAM, CPU registers, or on external devices in the system. The indirection allows fewer simpler
// instructions that work for both memory accesses and device I/O.
//
// To read or write to memory, the CPU puts the address into the address register (MAR) and any data
// into the data register (MDR). By calling either [Memory.Fetch] or [Memory.Store], the controller
// reads from the address into MDR or writes data in MDR to address, respectively, regardless if the
// memory is RAM or device memory.
//
// Admittedly, this is a strange design, at least from a software perspective. We could simply use
// function arguments and return values to pass the address and data values instead. However, we use
// registers to reflect the design of the reference micro-architecture and to model the clock cycles
// used for memory access.
//
// # Access Control
//
// The controller also enforces access control to the machine's address space. The system space
// contains the code and data used for operating the machine and must only be accessed by privileged
// programs. When the address register contains an address in the system space (or, is for a
// privileged device) and the processor is running with user privileges, then memory access will
// raise an access control violation (ACV) exception and a fault handler is called.
//
// # Data and Stacks
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
// # Interrupt Vector Tables
//
// In addition to system data and code, the system space includes small but important tables
// containing the addresses of service routines for I/O interrupts, traps, and exceptions. The
// system loads these tables with addresses of handlers and jumps to these handlers.
//
// # Diagram
//
// Since ASCII art is worth a thousand words:
//
//      +========+========+=================+    +-----------------+
//      |        | 0xffff |  Memory-mapped  |+   |                 |   +-------------------+
//      |        |        |     I/O page    ||   |                 |   |                   |
//      |        |   ...  |                 ||   |      MMU        |-->|	               |
//      |        |        |                 ||   |                 |-->|	   MMIO        |
//      |        | 0xfe00 |                 |+---|                 |-->|	               |
//      +========+========+=================+|   |                 |-->|	               |
//      |        | 0xfdff |                 ||   +--------+---+----+   |                   |
//      |        |        |                 ||            |   |        +--+-----+---+---+--+
//      |        |  ...   |   User stack    ||            |   |           |     |   |   |
//      |        |        |                 ||   +--------V---V----+   +--V-----V-+-V---V--+
//      |        | 0x4568 |                 |<---|USP    MCR PSR   |   | KBD  KBSR|DDR DSR |
//      |        +--------------------------+|   |                 |   |          |        |
//      |        | 0x4567 |                 |<---|R7(RET)        R3|   |          |        |
//      |  User  |  ...   |   User data     ||   |                 |   | Keyboard |Display |
//      |  space | 0x3000 |                 |<---|R6(SP)         R2|   |          |        |
//      +========+========+=================+|   |       CPU ️      |   |          |        |
//      |        | 0x2fff |                 ||   |R5             R1|   +----------+--------+
//      |        |        |                 ||   |                 |
//      |        |   ...  |  System stack   ||   |R4             R0|
//      |        |        |                 ||   |                 |
//      |        | 0x2dad |                 |<---|SSP              |
//      |        +--------+-----------------+|   +-----------------+
//      | System | 0x1234 |                 ||
//      | space  |  ...   |  System data    ||
//      |        | 0x0200 |                 ||
//      |        +--------------------------+|
//      |        | 0x01ff |    Interrupt    ||
//      |        |  ...   |  vector table   ||
//      |        | 0x0100 |                 ||
//      |        +--------+-----------------+|
//      |        | 0x00ff |      Trap       ||
//      |        |   ...  |   vector table  ||
//      |        | 0x0010 |                 ||
//      |        +--------+-----------------+|
//      |        | 0x000f |    Exception    ||
//      |        |   ...  |      table      ||
//      |        | 0x0000 |                 |+
//      +========+========+=================+
package vm
