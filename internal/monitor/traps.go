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
	Code: []asm.Operation{
		&asm.ST{SR: "R1", OFFSET: 11}, // TODO R1 will hold MCR address.
		&asm.ST{SR: "R0", OFFSET: 9},  // TODO R0 is a temporary value.

		// Print a message. Alert the media.
		&asm.LD{DR: "R0", SYMBOL: "ASCIINEWLINE"}, // Static symbol.
		&asm.TRAP{LITERAL: 0x21},                  // OUT
		&asm.LEA{DR: "R0", SYMBOL: "HALTMESSAGE"}, // Static symbol.
		&asm.TRAP{LITERAL: 0x22},                  // PUTS
		&asm.LD{DR: "R0", SYMBOL: "ASCIINEWLINE"}, // Static symbol.
		&asm.TRAP{LITERAL: 0x21},                  // OUT

		// Clear RUN flag in Machine Control Register.
		&asm.LDI{DR: "R1", OFFSET: 8},            // Load R1 <- MCR address
		&asm.LD{DR: "R0", OFFSET: 8},             // Load R0 <- MASK
		&asm.AND{DR: "R0", SR1: "R1", SR2: "R0"}, // Clear top bit.
		&asm.STI{SR: "R0", OFFSET: 6},            // Store value to MCR addr.

		// Exit from HALT(!): restore registers and return from trap.
		&asm.LD{DR: "R1", OFFSET: 0x0003},
		&asm.LD{DR: "R0", OFFSET: 0x0001},
		&asm.RTI{},

		// Routine data.
		&asm.BLKW{ALLOC: 1},                    // Stored R0.
		&asm.BLKW{ALLOC: 1},                    // Stored R1.
		&asm.FILL{LITERAL: uint16(vm.MCRAddr)}, // What it says.
		&asm.FILL{LITERAL: 0x7fff},             // MASK to clear top bit.
	},
}

// TrapOut is the system call to write a character to the display register.
//
// Adapted from Fig. 9.22, 3/e.
var TrapOut = Routine{
	Vector: vm.TrapTable + vm.TrapOUT,
	Orig:   0x0420,
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

	Symbols: asm.SymbolTable{
		"POLL":    0x0429 - 1,
		"INTMASK": 0x0436 - 1,
		"PSR":     0x0437 - 1,
		"DSR":     0x0438 - 1,
		"DDR":     0x0439 - 1,
	},
}
