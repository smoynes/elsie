package cpu

import (
	"math"
)

// Addressable Memory: 16-bit words with an address space of 16 bits.
// TODO: add memory access control
type Memory [AddressSpace]Word

// Size of addressable memory.
const AddressSpace = math.MaxUint16

func (mem *Memory) Load(addr Word) Word {
	return mem[addr]
}

func (mem *Memory) Store(addr Word, cell Word) {
	mem[addr] = cell
}

// PushStack pushes a word onto the current stack.
func (cpu *LC3) PushStack(w Word) {
	cpu.Mem[cpu.Reg[SP]-1] = w
	cpu.Reg[SP]--
}

// PopStack pops a word from the current stack into a register.
func (cpu *LC3) PopStack() Word {
	cpu.Reg[SP]++
	return cpu.Mem[cpu.Reg[SP]-1]
}
