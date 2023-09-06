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
