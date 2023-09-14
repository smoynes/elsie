// elsie is a LC-3 hardware emulator.
package main

import (
	"log"

	"github.com/smoynes/elsie/internal/vm"
)

func main() {
	log.SetFlags(log.Lmsgprefix | log.Lmicroseconds | log.Lshortfile)
	log.Println("Initializing machine")
	machine := vm.New()

	// TRAP HALT handler
	log.Println("Loading trap handlers")
	program := vm.Register(0x1000)
	machine.Mem.MAR = vm.Register(0x0025)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// AND R0,R0,0 ; clear R0
	program = vm.Register(vm.Word(vm.OpcodeAND)<<12 | 0x0020)
	machine.Mem.MAR = vm.Register(0x1000)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// LEA R1,[MCR] ; load MCR addr into R1
	program = vm.Register(vm.Word(vm.OpcodeLEA)<<12 | 0x0201)
	machine.Mem.MAR = vm.Register(0x1001)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// STR R0,R1,0
	program = vm.Register(vm.Word(vm.OpcodeSTR)<<12 | 0x0040)
	machine.Mem.MAR = vm.Register(0x1002)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// Store MCR addr
	machine.Mem.MAR = vm.Register(0x1003)
	machine.Mem.MDR = vm.Register(0xfffe)
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// TRAP HALT
	program = vm.Register(vm.Word(vm.OpcodeTRAP)<<12 | vm.TrapHALT)
	machine.Mem.MAR = vm.Register(machine.PC)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	if err := machine.Run(); err != nil {
		log.Fatal(err)
	}
}
