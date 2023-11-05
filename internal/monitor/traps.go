package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

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
		"HALTMESSAGE": 0x0527,
		"MCR":         0x0528,
		"MASK":        0x0529,
	},
	Code: []asm.Operation{
		// Print a message. Alert the media.
		/* 0x0520 */
		&asm.LEA{DR: "R0", SYMBOL: "HALTMESSAGE"},
		&asm.TRAP{LITERAL: 0x21}, // TODO: PUTS

		// Clear RUN flag in Machine Control Register.
		/* 0x0522 */
		&asm.LDI{DR: "R0", SYMBOL: "MCR"},        // R0 <- [MCR] ; Load MCR.
		&asm.LD{DR: "R1", SYMBOL: "MASK"},        // R1 <- MASK ; Load bitmask.
		&asm.AND{DR: "R0", SR1: "R0", SR2: "R1"}, // R1 <- R0 & R1 ; Clear top bit using bit mask
		&asm.STI{SR: "R0", SYMBOL: "MCR"},        // [MCR]<- R0 ; Replace value in MCR.

		// Halt again, if we reach here, forever.
		/* 0x0526 */
		&asm.BR{
			NZP:    uint8(vm.ConditionNegative | vm.ConditionPositive | vm.ConditionZero),
			SYMBOL: "",
			OFFSET: ^(uint16(5)) + 1,
		},

		// Routine data.
		/* 0x0527 */ &asm.STRINGZ{LITERAL: "HALT!"}, // HALTMESSAGE.
		/* 0x0528 */ &asm.FILL{LITERAL: uint16(vm.MCRAddr)}, // I/O address of MCR.
		/* 0x0529 */ &asm.FILL{LITERAL: 0x7fff}, // MASK to clear top bit.
	},
}

// TrapOut is the system call to write a single character to the display.
//   - Table:   0x0000
//   - Vector:  0x21
//   - Handler: 0x0420
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
// Adapted from github.com/chriagsakhuja/lc3tools under the terms of the Apache Software Licence.
var TrapPuts = Routine{
	Name:   "PUTS",
	Vector: vm.TrapTable + vm.Word(vm.TrapPUTS),
	Orig:   0x0460,
	Symbols: asm.SymbolTable{
		"LOOP":   0x0465,
		"RETURN": 0x046a,
		"DSR":    0x0429,
		"DDR":    0x0435,
	},
	Code: []asm.Operation{
		// Push R0,R1 onto the stack.
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
		&asm.BR{NZP: uint8(vm.ConditionNegative), OFFSET: 0x0000}, // Return if value is zero.

		// Call trap OUT.
		&asm.TRAP{LITERAL: uint16(vm.TrapOUT)},

		// Increment loop pointer.
		&asm.ADD{DR: "R1", SR1: "R1", LITERAL: 0x0001},

		&asm.BR{NZP: asm.CondNZP, OFFSET: uint16(^(-4))},

		// Restore stack.
		/*RETURN: 0x046a*/
		&asm.LDR{DR: "R1", SR: "R6", OFFSET: 0},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},
		&asm.LDR{DR: "R1", SR: "R6", OFFSET: 0},
		&asm.ADD{DR: "R6", SR1: "R6", LITERAL: 1},

		&asm.RTI{},

		// Trap-scoped variables.
		/*0x0438 */ &asm.FILL{LITERAL: uint16(vm.DSRAddr)}, // display status-, and
		/*0x0439 */ &asm.FILL{LITERAL: uint16(vm.DDRAddr)}, // data-registers.
	},
}
