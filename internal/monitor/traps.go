package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

// TrapHalt is the system call to halt the machine.
//   - Handler: 0x1000
//   - Table: 0x00
//   - Vector: 0x25
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
		&asm.TRAP{LITERAL: 0x21},                  // PUTC
		&asm.LEA{DR: "R0", SYMBOL: "HALTMESSAGE"}, // Static symbol.
		&asm.TRAP{LITERAL: 0x22},                  // PUTS
		&asm.LD{DR: "R0", SYMBOL: "ASCIINEWLINE"}, // Static symbol.
		&asm.TRAP{LITERAL: 0x21},                  // PUTC

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
