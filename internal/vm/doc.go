/*
Package vm implements a basic VM for executing LC3 machine code.

With the reason for the project to learn more about computer engineering, the design of the
virtual machine is meant to mimic or reflect the micro-architecture described in the text. For
example, as you might see elsewhere, executing an instruction uses several function executions
to mimic the microarchitecture.

# CPU #

The machine's CPU is extraordinarily simple. It has just:

  - a few registers: program counter, instuction, processor status, and a control regisers.
  - user- and system-stack pointer registers
  - a file of eight general-purpose registers
  - an device interrupt controller
  - a memory controller

# Memory #

Memory is where we keep our most precious things: programs and data. Luckily, the LC-3 has nearly
unlimited memory: 128 kilobytes in a 16-bit address space of 2-byte words. The addressable memory
space is divided into separate address spaces.

  - system space for operating system code and data
  - user space for unprivileged programs' code and data
  - an I/O page for memory-mapped device-registers

The memory controller (or MMU) mediates access to the address spaces from the CPU.

## Data Flow ##

The MMU is translated logical addresses in the memory space to physical memory might be in RAM, CPU
registers, on external devices throughout the system. The indirection of translating logical
addresses to physical memory affords simpler instructions that work for both device I/O and RAM.,

To read or write to memory, the CPU puts the address into the address register (MAR) and the data
into the data register (MDR) and either calls Fetch or Store; the controller will read from the
address into its data register or write to memory from MDR, respectively.

Admittedly, this is a strange design, at least from a software design perspective. We could simply
use function arguments and return values to pass values instead. However, we use registers here in
order to reflect the design of the LC-3 reference micro-architecture. For learning purposes, it
helps to make the data flow explicit and try to model the clock cycles that keeps us all in time.

## Access Control ##

The controller also enforces access control to each address space. The system space contains the
code and data used for operating the machine and must only be accessed by privileged programs.
When the address register contains an address in the system space (or, is for a privileged
device) and the processor is running with user privileges, then memory access will raise an
access control violation (ACV) exception and a fault handler is called.

## Data and Stacks ##

The user and system spaces are further divided into regions. Primarily, each space contains a
data region that includes global program data as well as the machine code for programs
themselves.

Temporary program data is stored on a stack: one for the system, the other for the user. The top
of the current stack is pointed to by a stack pointer (SP, i.e. R6). The other stack is saved in
a special-purpose register while it is inactive. That is, the system stack value is saved in SSP
when running with user privileges; likewise, the user's in USP while with system privileges.

Both stacks grow down; that is, when a word is pushed onto the stack, the address decreases and
will point at the new data on the top of the stack.

# Interrupt Vector Tables #

In addition to system data and code, the system space includes small but important tables
containing the addresses of service routines for I/O interrupts, traps, and exceptions. The
system loads these tables with addresses of handlers and jumps to these handlers.

## Figure ##

Since ASCII art is worth a thousand words:

		+========+========+=================+    +-----------------+
		|        | 0xffff |  Memory-mapped  |+   |                 |   +-------------------+
		|        |        |     I/O page    ||   |                 |   |                   |
		|        |   ...  |                 ||	 |      MMU        |-->|	               |
		|        |        |                 ||	 |                 |-->|	   MMIO        |
		|        | 0xfe00 |                 |+---|                 |-->|	               |
		+========+========+=================+|   |                 |-->|	               |
		|        | 0xfdff |                 ||   +--------+---+----+   | 				   |
		|        |        |                 ||            |   |		   +--+-----+---+---+--+
		|        |  ...   |   User stack    ||            |   |           |     |   |   |
		|        |        |                 ||   +--------V---V----+   +--v-----V-+-V---V-+
		|        | 0x4568 |                 |<---|USP    MCR PSR   |   | KBD  KBSR|DDR DSR|
		|        +--------------------------+|   |                 |   |          |       |
		|        | 0x4567 |                 |<---|R7(RET)        R3|   |          |       |
		|  User  |  ...   |   User data     ||   |                 |   |          |       |
		|  space | 0x3000 |                 |<---|R6(SP)         R2|   |          |       |
		+========+========+=================+|   |      CPU ⚙️      |   |          |       |
		|        | 0x2fff |                 ||   |R5             R1|   +------------------+
		|        |        |                 ||   |                 |
		|        |   ...  |  System stack   ||   |R4             R0|
		|        |        |                 ||   |                 |
		|        | 0x2dad |                 |<---|SSP              |
		|        +--------+-----------------+|   +-----------------+
		| System | 0x1234 |                 ||
		| space  |  ...   |  System data    ||
		|        | 0x0200 |                 ||
		|        +--------------------------+|
		|        | 0x01ff |    Interrupt    ||
		|        |  ...   |  vector table   ||
		|        | 0x0100 |                 ||
	    |        +--------+-----------------+|
	    |        | 0x00ff |      Trap       ||
	    |        |   ...  |   vector table  ||
	    |        | 0x0010 |                 ||
	    |        +--------+-----------------+|
	    |        | 0x000f |    Exception    ||
	    |        |   ...  |      table      ||
	    |        | 0x0000 |                 |+
	    +========+========+=================+
*/
package vm
