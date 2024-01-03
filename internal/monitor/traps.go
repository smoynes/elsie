package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

var defaultImageTraps = []Routine{
	TrapHalt,
	TrapOut,
	TrapPuts,
	TrapGetc,
}

// TrapGetc is the system call to prompt the user and wait for a character of input.
//
//   - Table:   0x0000
//   - Vector:  0x20
//   - Handler: 0x04a0
//
// Adapted from Fig. 9.15, 3/e. TODO: This does not disable interrupts.
var TrapGetc = Routine{
	Name:   "GETC",
	Vector: vm.TrapTable + vm.Word(vm.TrapGETC),
	Orig:   0x04a0,
	Symbols: asm.SymbolTable{
		"START":      0x04a0,
		"LOOP":       0x04a2,
		"INPUT":      0x04a6,
		"NEWLINE":    0x04ad,
		"PROMPT":     0x04ae,
		"WRITECHAR":  0x04c3,
		"READCHAR":   0x04c7,
		"SAVEREG":    0x04ca, //
		"RESTOREREG": 0x04d2,

		"SAVER1": 0x04d9,
		"SAVER2": 0x04da,
		"SAVER3": 0x04db,
		"SAVER4": 0x04dc,
		"SAVER5": 0x04dd,
		"SAVER6": 0x04de,

		"DSR":  0x04df,
		"DDR":  0x04e0,
		"KBSR": 0x04e1,
		"KBDR": 0x04e2,
	},
	Code: []asm.Operation{
		&asm.JSR{SYMBOL: "SAVEREG"},
		&asm.LEA{DR: "R1", SYMBOL: "PROMPT"},

		/*LOOP:0x04a2*/
		&asm.LDR{DR: "R2", SR: "R1", OFFSET: 0},   // Get next prompt character.
		&asm.JSR{SYMBOL: "WRITECHAR"},             // Echo prompt character.
		&asm.ADD{DR: "R1", SR1: "R1", LITERAL: 1}, // Increment prompt pointer.
		&asm.BR{NZP: asm.CondNZP, SYMBOL: "LOOP"}, // Iterate to LOOP.

		/*INPUT:0x04a6*/
		&asm.JSR{SYMBOL: "READCHAR"},              // Get character input.
		&asm.ADD{DR: "R2", SR1: "R0", LITERAL: 0}, // Move char for echo.
		&asm.JSR{SYMBOL: "WRITECHAR"},             // Echo to monitor.

		&asm.LD{DR: "R2", SYMBOL: "NEWLINE"},
		&asm.JSR{SYMBOL: "WRITECHAR"},  // Echo newline.
		&asm.JSR{SYMBOL: "RESTOREREG"}, // Restore registers.
		&asm.RTI{},                     // Terminate trap routine.

		/*NEWLINE:0x04ad*/
		&asm.FILL{LITERAL: 0x000a},

		/*PROMPT:0x04ae*/
		&asm.STRINGZ{LITERAL: "\nInput a character> "},

		/*WRITECHAR:0x04c3*/
		&asm.LDI{DR: "R3", SYMBOL: "DSR"},
		&asm.BR{NZP: asm.CondZP, SYMBOL: "WRITECHAR"},
		&asm.STI{SR: "R2", SYMBOL: "DDR"},
		&asm.RET{},

		/*READCHAR:0x04c7*/
		&asm.LDI{DR: "R3", SYMBOL: "KBSR"},
		&asm.BR{NZP: asm.CondZP, SYMBOL: "READCHAR"},
		&asm.LDI{DR: "R0", SYMBOL: "KBDR"},
		&asm.RET{},

		/*SAVEREG:0x04cb*/
		&asm.ST{SR: "R1", SYMBOL: "SAVER1"},
		&asm.ST{SR: "R2", SYMBOL: "SAVER2"},
		&asm.ST{SR: "R3", SYMBOL: "SAVER3"},
		&asm.ST{SR: "R4", SYMBOL: "SAVER4"},
		&asm.ST{SR: "R5", SYMBOL: "SAVER5"},
		&asm.ST{SR: "R6", SYMBOL: "SAVER6"},
		&asm.RET{},

		/*RESTOREREG:0x04d2*/
		&asm.ST{SR: "R1", SYMBOL: "SAVER1"},
		&asm.ST{SR: "R2", SYMBOL: "SAVER2"},
		&asm.ST{SR: "R3", SYMBOL: "SAVER3"},
		&asm.ST{SR: "R4", SYMBOL: "SAVER4"},
		&asm.ST{SR: "R5", SYMBOL: "SAVER5"},
		&asm.ST{SR: "R6", SYMBOL: "SAVER6"},
		&asm.RET{},

		// Stored register allocations.
		/*SAVER1:0x04d9*/
		&asm.BLKW{ALLOC: 0x0001},
		/*SAVER2:0x04da*/
		&asm.BLKW{ALLOC: 0x0001},
		/*SAVER3:0x04db*/
		&asm.BLKW{ALLOC: 0x0001},
		/*SAVER4:0x04dc*/
		&asm.BLKW{ALLOC: 0x0001},
		/*SAVER5:0x04dd*/
		&asm.BLKW{ALLOC: 0x0001},
		/*SAVER6:0x04de*/
		&asm.BLKW{ALLOC: 0x0001},

		// Address constants.
		/*DSR:0x04df*/
		&asm.FILL{LITERAL: 0xfe02},
		/*DDR:0x04e0*/
		&asm.FILL{LITERAL: 0xfe04},
		/*KBSR:0x04e1*/
		&asm.FILL{LITERAL: 0xfe00},
		/*DDR:0x04e2*/
		&asm.FILL{LITERAL: 0xfe02},
	},
}

// TrapHalt is the system call to stop the machine.
//   - Table:   0x0000
//   - Vector:  0x25
//   - Handler: 0x0520
//
// Adapted from https://github.com/chiragsakhuja/lc3tools/tree/master/src/lc3os.cpp
var TrapHalt = Routine{
	Name:   "HALT",
	Vector: vm.TrapTable + vm.Word(vm.TrapHALT),
	Orig:   0x0520,
	Symbols: asm.SymbolTable{
		"RETRY":       0x0521,
		"MCR":         0x0526,
		"MASK":        0x0527,
		"HALTMESSAGE": 0x0528,
	},
	Code: []asm.Operation{
		// Print a message. Alert the media.
		/* 0x0520 */
		&asm.LEA{DR: "R0", SYMBOL: "HALTMESSAGE"},
		&asm.TRAP{LITERAL: 0x22}, // Call Trap PUTS

		// Clear RUN flag in Machine Control Register.
		/* 0x0522 */
		&asm.LDI{DR: "R0", SYMBOL: "MCR"},        // R0 <- [MCR] ; Load MCR.
		&asm.LD{DR: "R1", SYMBOL: "MASK"},        // R1 <- MASK ; Load bitmask.
		&asm.AND{DR: "R0", SR1: "R0", SR2: "R1"}, // R1 <- R0 & R1 ; Clear top bit using bit mask
		&asm.STI{SR: "R0", SYMBOL: "MCR"},        // [MCR]<- R0 ; Replace value in MCR.

		// Halt again, if we reach here, forever.
		/* 0x0526 */
		&asm.BR{
			NZP:    asm.CondNZP,
			SYMBOL: "RETRY",
		},

		// Routine data.
		/* 0x0527 */ &asm.FILL{LITERAL: uint16(vm.MCRAddr)}, // I/O address of MCR.
		/* 0x0528 */ &asm.FILL{LITERAL: 0x7fff}, // MASK to clear top bit.
		/* 0x0529 */ &asm.STRINGZ{LITERAL: "\n\nMACHINE HALTED!\n\n"},
	},
}

// TrapOut is the system call to write a single character to the display.
//
//   - Table:   0x0000
//   - Vector:  0x21
//   - Handler: 0x0420
//   - Input:   R0, character to display.
//
// Adapted from Fig. 9.22, 3/e.
var TrapOut = Routine{
	Name:   "OUT",
	Vector: vm.TrapTable + vm.Word(vm.TrapOUT),
	Orig:   0x0420,
	Symbols: asm.SymbolTable{
		"POLL":    0x0429,
		"INTMASK": 0x0435,
		"PSR":     0x0436,
		"DSR":     0x0437,
		"DDR":     0x0438,
	},
	Code: []asm.Operation{
		// Push R1,R2,R3 onto the stack.
		/*0x0420 */
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 0xffff},
		&asm.STR{SR1: "R1", SR2: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 0xffff},
		&asm.STR{SR1: "R2", SR2: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 0xffff},
		&asm.STR{SR1: "R3", SR2: "R6"},

		// R1 <- [PSR] ; Fetch initial or previous value.
		/*0x0426 */
		&asm.LDI{DR: "R1", SYMBOL: "PSR"},

		// R2 <- [PSR] & ^IE ; Keep PSR with interrupts disabled.
		/*0x0427 */
		&asm.LD{DR: "R2", SYMBOL: "INTMASK"},
		&asm.AND{DR: "R2", SR1: "R1", SR2: "R2"},

		// POLL
		/*0x0429 */
		&asm.STI{SR: "R1", SYMBOL: "PSR"}, // Store R1 -> [PSR] ; Enable interrupts, if prev enabled.
		&asm.STI{SR: "R2", SYMBOL: "PSR"}, // Store R2 -> [PSR] ; Disable interrupts.

		/*0x042b */
		&asm.LDI{DR: "R3", SYMBOL: "DSR"}, // Fetch R3 <- [DSR] ; Check status.
		&asm.BR{ // Branch if top bit is 0, i.e. display not-ready.
			NZP:    uint8(vm.ConditionZero | vm.ConditionPositive),
			SYMBOL: "POLL",
		},

		// R0 -> [DDR] ; Store trap argument from R0 into DDR.
		/*0x042d */
		&asm.STI{SR: "R0", SYMBOL: "DDR"},

		// Restore PSR.
		/*0x042e */
		&asm.STI{SR: "R1", SYMBOL: "PSR"},

		// Restore R3,R2,R1 from stack.
		/*0x042f */
		&asm.LDR{DR: "R3", SR: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},
		&asm.LDR{DR: "R2", SR: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},
		&asm.LDR{DR: "R1", SR: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},

		// Return from trap.
		/*0x0435 */
		&asm.RTI{},

		// Trap-scoped variables.
		/*0x0436 */ &asm.FILL{LITERAL: 0xbfff}, // MASK to disable interrupts.
		/*0x0437 */ &asm.FILL{LITERAL: uint16(vm.PSRAddr)}, // I/O addresses: processor status-,
		/*0x0438 */ &asm.FILL{LITERAL: uint16(vm.DSRAddr)}, // display status-, and
		/*0x0439 */ &asm.FILL{LITERAL: uint16(vm.DDRAddr)}, // display data-registers
	},
}

// TrapPuts is the system call to write a zero-terminated string to the display.
//
//	Table:   0x0000
//	Vector:  0x22
//	Handler: 0x0460
//	Input:   R0, address of zero terminated string.
//
// Adapted from github.com/chriagsakhuja/lc3tools, under the terms of the Apache Software Licence.
// You are free to use this code under those terms rather than CC-BY-SA, if you wish, however
// unlikely it might be.
var TrapPuts = Routine{
	Name:   "PUTS",
	Vector: vm.TrapTable + vm.Word(vm.TrapPUTS),
	Orig:   0x0460,
	Symbols: asm.SymbolTable{
		"LOOP":   0x0464,
		"RETURN": 0x046a,
		"DSR":    0x0429,
		"DDR":    0x0435,
	},
	Code: []asm.Operation{
		// Push R0, R1 onto the stack.
		/*0x0460*/
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 0xffff},
		&asm.STR{SR1: "R0", SR2: "R6"},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 0xffff},
		&asm.STR{SR1: "R1", SR2: "R6"},

		// Move input pointer R0 to R1.
		/*0x0464*/
		&asm.ADD{DR: "R1", SR1: "R0"},

		// Loop over in array and write each value to DDR.
		/*LOOP: 0x0465*/
		&asm.LDR{DR: "R0", SR: "R1"},

		// Return if value is zero.
		&asm.BR{NZP: uint8(vm.ConditionZero), SYMBOL: "RETURN"},

		// Call trap OUT. Increment loop index, and branch to LOOP.
		/*0x0467*/
		&asm.TRAP{LITERAL: uint16(vm.TrapOUT)},
		&asm.ADD{DR: "R1", SR1: "R1", LITERAL: 0x0001},
		&asm.BR{NZP: asm.CondNZP, SYMBOL: "LOOP"},

		// Restore stack.
		/*RETURN: 0x046a*/
		&asm.LDR{DR: "R1", SR: "R6", OFFSET: 0},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},
		&asm.LDR{DR: "R0", SR: "R6", OFFSET: 0},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},

		&asm.RTI{},

		// Trap-scoped variables.
		/*0x046f */ &asm.FILL{LITERAL: uint16(vm.DSRAddr)}, // display status-, and
		/*0x0470 */ &asm.FILL{LITERAL: uint16(vm.DDRAddr)}, // data-registers.
	},
}
