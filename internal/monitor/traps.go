package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

// TrapHalt is the system call to stop the machine.
//   - Table:   0x0000
//   - Vector:  0x25
//   - Handler: 0x0520
//   - Size:    0x13
//
// Adapted from Fig. 9.14, 3/e.
var TrapHalt = Routine{
	Vector: vm.TrapTable + vm.TrapHALT,
	Orig:   0x0520,
	Symbols: asm.SymbolTable{
		"ASCIINEWLINE": 0x052f - 1,
		"SAVER0":       0x0530 - 1,
		"SAVER1":       0x0531 - 1,
		"HALTMESSAGE":  0x0532 - 1,
		"MCR":          0x0533 - 1,
		"MASK":         0x0534 - 1,
	},
	Code: []asm.Operation{
		/* 0x0520 */
		&asm.ST{SR: "R1", SYMBOL: "SAVER1"}, // R1 -> [SAVER1] ; Store R1 to hold MCR address.
		&asm.ST{SR: "R0", SYMBOL: "SAVER0"}, // R0 -> [SAVER0] ; Store R0 to hold temporary value.

		// Print a message. Alert the media.
		/* 0x0522 */
		&asm.LD{DR: "R0", SYMBOL: "ASCIINEWLINE"},
		&asm.TRAP{LITERAL: 0x21}, // OUT
		&asm.LEA{DR: "R0", SYMBOL: "HALTMESSAGE"},
		&asm.TRAP{LITERAL: 0x21}, // TODO: PUTS
		&asm.LD{DR: "R0", SYMBOL: "ASCIINEWLINE"},
		&asm.TRAP{LITERAL: 0x21}, // OUT

		// Clear RUN flag in Machine Control Register.
		/* 0x0528 */
		&asm.LDI{DR: "R1", SYMBOL: "MCR"},        // Fetch R1 <- [MCR] ; Load MCR.
		&asm.LD{DR: "R0", SYMBOL: "MASK"},        // Fetch R0 <- MASK ; Load bitmask.
		&asm.AND{DR: "R0", SR1: "R1", SR2: "R0"}, // Clear top bit.
		&asm.STI{SR: "R0", SYMBOL: "MCR"},        // Store R0 -> [MCR] ; Replace value in MCR.

		// Exit from HALT(!?): restore registers and return from trap.
		/* 0x052c */
		&asm.LD{DR: "R1", OFFSET: 0x0003},
		&asm.LD{DR: "R0", OFFSET: 0x0001},
		&asm.RTI{},

		// Routine data.
		/* 0x052f */ &asm.FILL{LITERAL: uint16('\n')}, // ASCIINEWLINE.
		/* 0x0530 */ &asm.BLKW{ALLOC: 1}, // Stored R0.
		/* 0x0531 */ &asm.BLKW{ALLOC: 1}, // Stored R1.
		/* 0x0532 */ &asm.STRINGZ{LITERAL: "!"}, // HALTMESSAGE.
		/* 0x0533 */ &asm.FILL{LITERAL: uint16(vm.MCRAddr)}, // What it says.
		/* 0x0534 */ &asm.FILL{LITERAL: 0x7fff}, // MASK to clear top bit.
	},
}

// TrapOut is the system call to write a character to the display.
//   - Table:   0x0000
//   - Vector:  0x21
//   - Handler: 0x0420
//   - Size:    ??
//
// Adapted from Fig. 9.14, 3/e.
//
// Adapted from Fig. 9.22, 3/e.
var TrapOut = Routine{
	Vector: vm.TrapTable + vm.TrapOUT,
	Orig:   0x0420,
	Symbols: asm.SymbolTable{
		"POLL":    0x0429 - 1,
		"INTMASK": 0x0436 - 1,
		"PSR":     0x0437 - 1,
		"DSR":     0x0438 - 1,
		"DDR":     0x0439 - 1,
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

		// R1 <- [PSR] & ^IE ; Keep PSR with interrupts disabled.
		/*0x0427 */
		&asm.LD{DR: "R2", SYMBOL: "INTMASK"},
		&asm.AND{DR: "R2", SR1: "R1", SR2: "R2"},

		// POLL
		/*0x0429 */
		&asm.STI{SR: "R1", SYMBOL: "PSR"}, // Store R1 -> [PSR] ; Enable interrupts (possibly).
		&asm.STI{SR: "R2", SYMBOL: "PSR"}, // Store R2 -> [PSR] ; Disable interrupts.

		/*0x042b */
		&asm.LDI{DR: "R3", SYMBOL: "DSR", OFFSET: 0}, // Fetch R3 <- [DSR] ; Check status.
		&asm.BR{ // Branch if not-ready.
			NZP:    uint8(vm.ConditionZero | vm.ConditionPositive),
			SYMBOL: "POLL",
		},

		// R0 -> [DDR] ; Store trap argument R0 in DDR.
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
		/*0x0437 */ &asm.FILL{LITERAL: uint16(vm.PSRAddr)}, // Register I/O addresses.
		/*0x0438 */ &asm.FILL{LITERAL: uint16(vm.DSRAddr)},
		/*0x0439 */ &asm.FILL{LITERAL: uint16(vm.DDRAddr)},
	},
}
