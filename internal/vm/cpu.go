// Package vm provides an emulated CPU.
package vm

import (
	"fmt"
)

// LC3 is a computer simulated in software.
type LC3 struct {
	PC  ProgramCounter  // Instruction Pointer
	IR  Instruction     // Instruction Register
	PSR ProcessorStatus // Processor Status Register
	Reg RegisterFile    // General-purpose register file
	USP Register        // User stack pointer
	SSP Register        // System stack pointer
	MCR Register        // Master control register
	Mem Memory          // All the memory you'll ever need.
	log logger          // A log of where we've been.
}

// An OptionFn is modifies the machine during late initialization. That is, the
// function is called after all resources are initialized but before any are used.
type OptionFn func(*LC3)

func WithLogger(log logger) OptionFn {
	return func(vm *LC3) {
		vm.withLogger(log)
	}
}

// New initializes a virtual machine state.
func New(opts ...OptionFn) *LC3 {

	// Initialize processor status...
	var status ProcessorStatus

	// Start with system privileges so we can access privileged memory and
	// configure devices. Privileges are dropped after late initialization.
	status |= (StatusPrivilege & StatusSystem)

	// Don't rush things, priority normal.
	status |= (StatusPriority & StatusNormal)

	// All condition codes are set.
	status |= StatusCondition

	// Set CPU registers to known values.
	cpu := LC3{
		PC:  0x0300,
		IR:  0x0000,
		PSR: status,
		USP: Register(IOPageAddr),    // User stack grows down from the top of user space.
		SSP: Register(UserSpaceAddr), // Similarly, system stack starts where user space ends.
		MCR: Register(0x8000),        // Set the RUN flag. 🤾

		log: defaultLogger(),
	}

	// Initialize general purpose registers to a pleasing pattern.
	copy(cpu.Reg[:], []Register{
		0xffff, 0x0000,
		0xfff0, 0xf000,
		0xff00, 0x0f00,
		cpu.USP, 0x00f0,
	})

	cpu.Reg[SP] = cpu.USP // Initial stack is user.

	// Configure MMU.
	cpu.Mem = NewMemory(&cpu.PSR)

	kbd := newDevice(&cpu, &Keyboard{}, []Word{KBSRAddr, KBDRAddr})
	disp := newDevice(&cpu, nil, []Word{DSRAddr, KBDRAddr})
	devs := map[Word]any{
		MCRAddr:  &cpu.MCR,
		PSRAddr:  &cpu.PSR,
		KBSRAddr: &kbd,
		KBDRAddr: &kbd,
		DSRAddr:  &disp,
		DDRAddr:  &disp,
	}

	// Run early init.
	for _, fn := range opts {
		fn(&cpu)
	}

	err := cpu.Mem.device.Map(devs)

	if err != nil {
		cpu.log.Panic(err)
	}

	// Drop privileges and execute as user.
	cpu.PSR &^= (StatusPrivilege & StatusUser)

	// Run late init..
	for _, fn := range opts {
		fn(&cpu)
	}

	return &cpu
}

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC:  %s IR: %s \nPSR: %s\nUSP: %s SSP: %s MCR: %s\n"+
		"MAR: %s MDR: %s\n",
		cpu.PC, cpu.IR, cpu.PSR, cpu.USP, cpu.SSP, cpu.MCR,
		cpu.Mem.MAR, cpu.Mem.MDR)
}

// PushStack pushes a word onto the current stack.
func (cpu *LC3) PushStack(w Word) error {
	cpu.Reg[SP]--
	cpu.Mem.MAR = cpu.Reg[SP]
	cpu.Mem.MDR = Register(w)
	return cpu.Mem.Store()
}

// PopStack pops a word from the current stack into MDR.
func (cpu *LC3) PopStack() error {
	cpu.Reg[SP]++
	cpu.Mem.MAR = cpu.Reg[SP] - 1
	return cpu.Mem.Fetch()
}
