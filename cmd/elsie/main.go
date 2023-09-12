// elsie is a LC-3 hardware emulator.
package main

import (
	"log"

	"github.com/smoynes/elsie/internal/cpu"
)

func main() {
	machine := cpu.New()
	copy(machine.Reg[:], []cpu.Register{0xffff, 0xface, 0xadad, 0xf001})

	// TRAP HALT
	program := cpu.Register(cpu.Word(cpu.OpcodeTRAP)<<12 | cpu.TrapHALT)
	machine.Mem.MAR = cpu.Register(machine.PC)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// TRAP HALT handler
	program = cpu.Register(0x1000)
	machine.Mem.MAR = cpu.Register(0x0025)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// AND R0,R0,0
	program = cpu.Register(cpu.Word(cpu.OpcodeAND)<<12 | 0x0020)
	machine.Mem.MAR = cpu.Register(0x1000)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	// STI MCR,R0
	program = cpu.Register(cpu.Word(cpu.OpcodeSTI)<<12 | 0x0010)
	machine.Mem.MAR = cpu.Register(0x1001)
	machine.Mem.MDR = program
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}

	println(machine.String())
	println(machine.Reg.String())

	if err := machine.Cycle(); err != nil {
		log.Fatal(err)
	}

	println()
	println("Post cycle state:")
	println(machine.String())
	println(machine.Reg.String())

	if err := machine.Cycle(); err != nil {
		log.Fatal(err)
	}

	println()
	println("Post cycle state:")
	println(machine.String())
	println(machine.Reg.String())

	if err := machine.Cycle(); err != nil {
		log.Fatal(err)
	}

	println()
	println("Post cycle state:")
	println(machine.String())
	println(machine.Reg.String())
	if err := machine.Cycle(); err != nil {
		log.Fatal(err)
	}

	println()
	println("Post cycle state:")
	println(machine.String())
	println(machine.Reg.String())

}
