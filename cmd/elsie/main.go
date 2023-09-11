// elsie is a LC-3 hardware emulator.
package main

import (
	"log"

	"github.com/smoynes/elsie/internal/cpu"
)

func main() {
	machine := cpu.New()

	// TRAP HALT
	instruction := cpu.Register(cpu.Word(cpu.OpcodeTRAP)<<12 | cpu.TrapHALT)
	machine.Mem.MAR = cpu.Register(machine.PC)
	machine.Mem.MDR = instruction
	if err := machine.Mem.Store(); err != nil {
		log.Fatal(err)
	}
	machine.Reg[cpu.R0] = 0xffff

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
