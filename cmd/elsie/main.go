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

	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")

	if err := machine.Cycle(); err != nil {
		log.Fatal(err)
	}
	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")

}
