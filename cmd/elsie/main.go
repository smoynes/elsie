// elsie is a LC-3 hardware emulator.
package main

import (
	"log"

	"github.com/smoynes/elsie/internal/cpu"
)

func main() {
	log.Println("Initializing machine")
	machine := cpu.New()

	// TRAP HALT handler
	log.Println("Loading trap handlers")
	program := cpu.Register(0x1000)
	machine.Mem.MAR = cpu.Register(0x0025)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// AND R0,R0,0 ; clear R0
	program = cpu.Register(cpu.Word(cpu.OpcodeAND)<<12 | 0x0020)
	machine.Mem.MAR = cpu.Register(0x1000)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// LEA R1,[MCR] ; load MCR addr into R1
	program = cpu.Register(cpu.Word(cpu.OpcodeLEA)<<12 | 0x0201)
	machine.Mem.MAR = cpu.Register(0x1001)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// STR R0,R1,0
	program = cpu.Register(cpu.Word(cpu.OpcodeSTR)<<12 | 0x0040)
	machine.Mem.MAR = cpu.Register(0x1002)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// Store MCR addr
	machine.Mem.MAR = cpu.Register(0x1003)
	machine.Mem.MDR = cpu.Register(0xfffe)
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// TRAP HALT
	program = cpu.Register(cpu.Word(cpu.OpcodeTRAP)<<12 | cpu.TrapHALT)
	machine.Mem.MAR = cpu.Register(machine.PC)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	if err := machine.Run(); err != nil {
		log.Fatal(err)
	}
}
