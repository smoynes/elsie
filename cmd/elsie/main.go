// elsie is a LC-3 hardware emulator.
package main

import (
	"github.com/smoynes/elsie/internal/cpu"
)

func main() {
	machine := cpu.New()
	machine.Mem[machine.PC] = cpu.Word(cpu.OpcodeAND)<<12 | 0x0020 | 0x0011
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
