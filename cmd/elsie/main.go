// elsie is a LC-3 hardware emulator.
package main

import (
	"github.com/smoynes/elsie/internal/cpu"
)

func main() {
	machine := cpu.New()
	machine.Mem.MAR = cpu.Register(machine.PC)
	machine.Mem.MDR = cpu.Register(cpu.OpcodeAND)<<12 | 0x0040 | 0x0011
	machine.Mem.Store()
	machine.Reg[cpu.R0] = 0xffff
	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")

	machine.Cycle()
	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")

	machine.Cycle()
	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")

	machine.Cycle()
	print(machine.String(), "\n")
	print(machine.Reg.String(), "\n")
}
