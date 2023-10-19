package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

// TrapHalt is the system call to halt the machine.
//
//   - Handler: 0x1000
//   - Table: 0x00
//   - Vector: 0x25
var TrapHalt = Routine{
	Vector: vm.TrapTable + vm.TrapHALT,
	Orig:   0x1000,
	Code: []asm.Operation{
		&asm.AND{DR: "R0", SR1: "R0", LITERAL: 0},
		&asm.LEA{DR: "R1", OFFSET: 0x01},
		&asm.STR{SR1: "R0", SR2: "R1", OFFSET: 0},
		&asm.FILL{LITERAL: uint16(vm.MCRAddr)},
	},
}
