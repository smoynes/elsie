package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/vm"
)

type generator interface {
	Generate(sym asm.SymbolTable, pc uint16) ([]uint16, error)
}

func encode(oper generator) vm.Word {
	code, _ := oper.Generate(nil, 0)
	return vm.Word(code[0])
}

var haltHandler = vm.ObjectCode{
	Orig: 0x1000,
	Code: []vm.Word{
		encode(asm.AND{DR: "R0", SR1: "R0", LITERAL: 0}),
		encode(&asm.LEA{DR: "R1", OFFSET: 0x01}),
		encode(&asm.STR{SR1: "R0", SR2: "R1", OFFSET: 0}),
		vm.MCRAddr,
	},
}

type TrapHalt struct{}

var haltVector = vm.TrapTable + vm.TrapHALT

func (*TrapHalt) Vector() (vm.Word, vm.ObjectCode) {
	return haltVector, haltHandler
}
