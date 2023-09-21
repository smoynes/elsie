// Package vm provides an emulated CPU.
package vm

import (
	"fmt"
)

// LC3 is a computer simulated in software.
type LC3 struct {
	PC  ProgramCounter  // Instruction Pointer.
	IR  Instruction     // Instruction Register
	PSR ProcessorStatus // Processor Status Register.
	REG RegisterFile    // General-purpose Register File
	USP Register        // User Stack Pointer.
	SSP Register        // System Stack Pointer.
	MCR ControlRegister // Master Control Register.
	INT Interrupt       // Interrupt Line.
	Mem Memory          // All the memory you'll ever need!

	log logger // A log of where we've been.
}

// New initializes a virtual machine state.
func New(opts ...OptionFn) *LC3 {
	// Initialize processor status...
	var status ProcessorStatus

	// Start with system privileges so we can access privileged memory and
	// configure devices. Privileges are dropped after late initialization.
	status |= (StatusPrivilege & StatusSystem)

	// Don't rush things, low priority.
	status |= (StatusPriority & StatusLow)

	// All condition codes are set.
	status |= StatusCondition

	// Set CPU registers to known values.
	cpu := LC3{
		PC:  0x0300,
		IR:  0x0000,
		PSR: status,
		USP: Register(IOPageAddr),    // User stack grows down from the top of user space.
		SSP: Register(UserSpaceAddr), // Similarly, system stack starts where user space ends.
		MCR: ControlRegister(0x8000), // Set the RUN flag. 🤾
		INT: Interrupt{
			log: defaultLogger(),
		},

		log: defaultLogger(),
	}

	// Initialize general purpose registers to a pleasing pattern.
	copy(cpu.REG[:], []Register{
		0xffff, 0x0000,
		0xfff0, 0xf000,
		0xff00, 0x0f00,
		cpu.USP, 0x00f0, // ... except the user stack.
	})

	// Configure memory.
	cpu.Mem = NewMemory(&cpu.PSR)

	// Create devices.
	var (
		// The keyboard device is hardwired and does not have a separate
		// driver.
		kbd = Keyboard{KBSR: 0x0000, KBDR: '?'}

		// The display is more complicated: a driver configures the
		// device with the addresses for the display registers.
		display       = Display{DDR: '!'}
		handle        = NewDeviceHandle[*Display, Display](display)
		displayDriver = DisplayDriver{handle: *handle}

		// Device configuration for the I/O.
		devices = map[Word]any{
			MCRAddr:  &cpu.MCR,
			PSRAddr:  &cpu.PSR,
			KBSRAddr: &kbd,
			KBDRAddr: &kbd,
			DSRAddr:  &displayDriver,
			DDRAddr:  &displayDriver,
		}
	)

	// Run early init.
	for _, fn := range opts {
		fn(&cpu)
	}

	err := cpu.Mem.Devices.Map(devices)

	if err != nil {
		cpu.log.Panic(err)
	}

	cpu.log.Print("Configuring devices and drivers")

	kbd.Init(&cpu, nil)                                 // Keyboard needs no configuration.
	displayDriver.Init(&cpu, []Word{DSRAddr, KBDRAddr}) // Configure the display's address range.

	// Drop privileges and execute as user.
	cpu.PSR &^= (StatusPrivilege & StatusUser)

	// Run late init...
	for _, fn := range opts {
		fn(&cpu)
	}

	return &cpu
}

func (cpu *LC3) String() string {
	return fmt.Sprintf("PC:  %s IR:  %s \nPSR: %s\nUSP: %s SSP: %s MCR: %s\n"+
		"MAR: %s MDR: %s",
		cpu.PC.String(), cpu.IR.String(), cpu.PSR, cpu.USP, cpu.SSP, cpu.MCR,
		cpu.Mem.MAR, cpu.Mem.MDR)
}

// PushStack pushes a word onto the current stack.
func (cpu *LC3) PushStack(w Word) error {
	cpu.REG[SP]--
	cpu.Mem.MAR = cpu.REG[SP]
	cpu.Mem.MDR = Register(w)

	return cpu.Mem.Store()
}

// PopStack pops a word from the current stack into MDR.
func (cpu *LC3) PopStack() error {
	cpu.REG[SP]++
	cpu.Mem.MAR = cpu.REG[SP] - 1

	return cpu.Mem.Fetch()
}

// An OptionFn is modifies the machine during late initialization. That is, the
// function is called after all resources are initialized but before any are used.
type OptionFn func(*LC3)

func WithSystemPrivileges() OptionFn {
	return func(vm *LC3) {
		vm.PSR &^= (StatusPrivilege & StatusUser)
	}
}
