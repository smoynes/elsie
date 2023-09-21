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
	vm := LC3{
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
	copy(vm.REG[:], []Register{
		0xffff, 0x0000,
		0xfff0, 0xf000,
		0xff00, 0x0f00,
		vm.USP, 0x00f0, // ... except the user stack.
	})

	// Configure memory.
	vm.Mem = NewMemory(&vm.PSR)

	// Create devices.
	var (
		// The keyboard device is hardwired and does not have a separate
		// driver.
		kbd *Keyboard = NewKeyboard()

		// The display is more complicated: a driver configures the
		// device with the addresses for the display registers.
		display       = Display{DDR: '!'}
		handle        = NewDeviceHandle[*Display, Display](display)
		displayDriver = DisplayDriver{handle: *handle}

		// Device configuration for the I/O.
		devices = map[Word]any{
			MCRAddr:  &vm.MCR,
			PSRAddr:  &vm.PSR,
			KBSRAddr: kbd,
			KBDRAddr: kbd,
			DSRAddr:  &displayDriver,
			DDRAddr:  &displayDriver,
		}
	)

	// Run early init.
	for _, fn := range opts {
		fn(&vm)
	}

	err := vm.Mem.Devices.Map(devices)

	if err != nil {
		vm.log.Panic(err)
	}

	vm.log.Print("Configuring devices and drivers")

	kbd.Init(&vm, nil)                                 // Keyboard needs no configuration.
	displayDriver.Init(&vm, []Word{DSRAddr, KBDRAddr}) // Configure the display's address range.

	// Drop privileges and execute as user.
	vm.PSR &^= (StatusPrivilege & StatusUser)

	// Run late init...
	for _, fn := range opts {
		fn(&vm)
	}

	return &vm
}

func (vm *LC3) String() string {
	return fmt.Sprintf("PC:  %s IR:  %s \nPSR: %s\nUSP: %s SSP: %s MCR: %s\n"+
		"MAR: %s MDR: %s",
		vm.PC.String(), vm.IR.String(), vm.PSR, vm.USP, vm.SSP, vm.MCR,
		vm.Mem.MAR, vm.Mem.MDR)
}

// PushStack pushes a word onto the current stack.
func (vm *LC3) PushStack(w Word) error {
	vm.REG[SP]--
	vm.Mem.MAR = vm.REG[SP]
	vm.Mem.MDR = Register(w)

	return vm.Mem.Store()
}

// PopStack pops a word from the current stack into MDR.
func (vm *LC3) PopStack() error {
	vm.REG[SP]++
	vm.Mem.MAR = vm.REG[SP] - 1

	return vm.Mem.Fetch()
}

// An OptionFn is modifies the machine during late initialization. That is, the
// function is called after all resources are initialized but before any are used.
type OptionFn func(*LC3)

func WithSystemPrivileges() OptionFn {
	return func(vm *LC3) {
		vm.PSR &^= (StatusPrivilege & StatusUser)
	}
}
