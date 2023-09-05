package cpu

import (
	"math"
)

// Number of memory addresses: 2^16 values, i.e. {0x0000, ..., 0xFFFF}
const AddressSpace = math.MaxUint16

// Addressable Memory: 16-bit words with an address space of 16 bits.
type Memory [AddressSpace]Word

func (mem *Memory) Load(addr Word) Word {
	return mem[addr]
}
