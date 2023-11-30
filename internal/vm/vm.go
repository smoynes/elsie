package vm

// vm.go defines the virtual machine and assembles it from smaller parts.

import (
	"fmt"
	"strings"

	"github.com/smoynes/elsie/internal/log"
)

// LC3 is a computer simulated in software.
type LC3 struct {
	PC  ProgramCounter  // Instruction Pointer.
	IR  Instruction     // Instruction Register
	PSR ProcessorStatus // Processor Status Register.
	MCR ControlRegister // Master Control Register.
	USP Register        // User Stack Pointer.
	SSP Register        // System Stack Pointer.
	REG RegisterFile    // General-purpose Register File
	INT Interrupt       // Interrupt Line.
	Mem Memory          // All the memory you'll ever need!

	log *log.Logger // A record of where we've been.
}

// New creates and initializes a virtual machine. The initial state may be affected passing a
// sequence of OptionFn. Each function is called in sequence **twice**:
//
//   - early, before drivers are initialized and devices are mapped; and
//   - late, after device configuration.
//
// Notably, early init executes with system privileges and stack; late init runs after privileges
// are dropped.
//
// # Bugs
//
// This is a weird design.
func New(opts ...OptionFn) *LC3 {
	vm := LC3{}
	vm.initializeRegisters()

	// Configure memory.
	vm.Mem = NewMemory(&vm.PSR)

	// Create devices.
	var (
		// The keyboard device is hardwired and does not have a separate driver.
		kbd = NewKeyboard()

		// The display is more complicated: a driver configures the device with the addresses for
		// the display registers.
		display       = NewDisplay()
		displayDriver = NewDisplayDriver(display)

		// Device configuration for memory-mapped I/O.
		devices = map[Word]any{
			MCRAddr:  &vm.MCR,
			PSRAddr:  &vm.PSR,
			KBSRAddr: kbd,
			KBDRAddr: kbd,
			DSRAddr:  displayDriver,
			DDRAddr:  displayDriver,
		}
	)

	vm.updateLogger(log.DefaultLogger())

	err := vm.Mem.Devices.Map(devices)
	if err != nil {
		vm.log.Error(err.Error())
		panic(err)
	}

	// Run early-init after mapping devices but before initializing them. This allows options to
	// override or replace drivers before initialization and afterwards during late-init.
	for _, fn := range opts {
		fn(&vm, false)
	}

	vm.log.Debug("Configuring devices and drivers")

	kbd.Init(&vm, nil)                                // Keyboard needs no configuration.
	displayDriver.Init(&vm, []Word{DSRAddr, DDRAddr}) // Configure the display's address range.

	// Drop privileges and switch to user execution context.
	vm.PSR &^= (StatusPrivilege & StatusUser)
	vm.PSR |= (StatusPriority & StatusNormal) // Debatable.
	vm.REG[SP] = vm.USP

	// Run late init...
	for _, fn := range opts {
		fn(&vm, true)
	}

	return &vm
}

func (vm *LC3) String() string {
	return fmt.Sprintf("PC:  %s IR:  %s \nPSR: %s\nUSP: %s SSP: %s MCR: %s\n"+
		"MAR: %s MDR: %s",
		vm.PC.String(), vm.IR.String(), vm.PSR.String(), vm.USP.String(), vm.SSP.String(),
		vm.MCR.String(), vm.Mem.MAR.String(), vm.Mem.MDR.String())
}

// initializeRegisters sets the initial values of the virtual machine.
func (vm *LC3) initializeRegisters() {
	// Start with system privileges so we can access privileged memory and configure devices; in
	// particular, access is permitted to the system memory space. Privileges are dropped after late
	// initialization.
	vm.PSR = (StatusPrivilege & StatusSystem)

	// Don't rush things, low priority.
	vm.PSR |= (StatusPriority & StatusLow)

	// No condition codes are set, initially, though this is undefined.
	vm.PSR |= StatusCondition & ^(StatusNegative | StatusZero | StatusPositive)

	vm.PC = ProgramCounter(UserSpaceAddr) // First instruction is at the bottom of user space.
	vm.USP = Register(IOPageAddr)         // User stack grows down from the top of user space.
	vm.SSP = Register(UserSpaceAddr)      // System stack starts where user space begins, grows down.
	vm.MCR = ControlRegister(0x8000)      // Set the RUN flag. 🤾

	// Initialize general purpose registers to a pleasing pattern... except for the stack pointer.
	// Here, REG[SP] is set to SSP, but as for the privilege level, the stack is reset to the user
	// context.
	copy(vm.REG[:], []Register{
		0xffff, 0x0000,
		0xfff0, 0xf000,
		0xff00, 0x0f00,
		vm.SSP, 0x00f0,
	})
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

// ProgramCounter is a special-purpose register that points to the next instruction in memory.
type ProgramCounter Register

func (p ProgramCounter) String() string {
	return Word(p).String()
}

// ProcessStatus is a special-purpose register that records important CPU flags:
// privilege, priority level, and condition flags.
//
// | PR | 000 0 | PL | 0000 0 | COND |
// +----+-------+----+--------+------+
// | 15 |14   12|11 9|8      3|2    0|
type ProcessorStatus Register

// Init configures the device at startup.
func (ps *ProcessorStatus) Init(_ *LC3, _ []Word) {
	*ps = ProcessorStatus(0x8080)
}

// Get reads the register for I/O.
func (ps ProcessorStatus) Get() Register {
	return Register(ps)
}

// Put sets the register value for I/O.
func (ps *ProcessorStatus) Put(val Register) {
	*ps = ProcessorStatus(val)
}

// Status flags in PSR vector.
const (
	StatusPositive  ProcessorStatus = 0x0001
	StatusZero      ProcessorStatus = 0x0002
	StatusNegative  ProcessorStatus = 0x0004
	StatusCondition ProcessorStatus = StatusNegative | StatusZero | StatusPositive

	StatusPriority ProcessorStatus = 0x0700
	StatusHigh     ProcessorStatus = 0x0700
	StatusNormal   ProcessorStatus = 0x0300
	StatusLow      ProcessorStatus = 0x0000

	StatusPrivilege ProcessorStatus = 0x8000
	StatusUser      ProcessorStatus = 0x8000
	StatusSystem    ProcessorStatus = 0x0000
)

func (ps ProcessorStatus) String() string {
	return fmt.Sprintf(
		"%s (N:%t Z:%t P:%t PR:%s PL:%d)",
		Word(ps), ps.Negative(), ps.Zero(), ps.Positive(),
		ps.Privilege(), ps.Priority(),
	)
}

// Cond returns the condition codes from the status register.
func (ps ProcessorStatus) Cond() Condition {
	return Condition(ps & StatusCondition)
}

// Any returns true if any of the flags in the condition are set in the status
// register.
func (ps ProcessorStatus) Any(cond Condition) bool {
	return ps.Cond()&cond != 0
}

// Set sets the condition flags based on the zero, negative, and
// positive attributes of the register value.
func (ps *ProcessorStatus) Set(reg Register) {
	// Clear condition flags.
	*ps &= ^StatusCondition

	// Set condition flag from register sign.
	switch {
	case reg == 0:
		*ps |= StatusZero
	case int16(reg) > 0:
		*ps |= StatusPositive
	default:
		*ps |= StatusNegative
	}
}

// Positive returns true if the P flag is set.
func (ps ProcessorStatus) Positive() bool {
	return ps&StatusPositive != 0
}

// Negative returns true if the N flag is set.
func (ps ProcessorStatus) Negative() bool {
	return ps&StatusNegative != 0
}

// Zero returns true if the Z flag is set.
func (ps ProcessorStatus) Zero() bool {
	return ps&StatusZero != 0
}

// Priority returns the priority level of the current task.
func (ps ProcessorStatus) Priority() Priority {
	return Priority(ps & StatusPriority >> 8)
}

// Privilege returns the privilege of the current task.
func (ps ProcessorStatus) Privilege() Privilege {
	return Privilege(ps & StatusPrivilege >> 15)
}

func (ps *ProcessorStatus) device() string {
	return Register(*ps).String()
}

// RegisterFile is the set of general purpose registers.
type RegisterFile [NumGPR]Register

func (rf RegisterFile) String() string {
	b := strings.Builder{}
	for i := 0; i < len(rf)/2; i++ {
		fmt.Fprintf(&b, "R%d:  %s R%d: %s\n",
			i, rf[i], i+len(rf)/2, rf[i+len(rf)/2])
	}

	return b.String()
}

func (rf RegisterFile) LogValue() log.Value {
	return log.GroupValue(
		log.String("R0", rf[R0].String()),
		log.String("R1", rf[R1].String()),
		log.String("R2", rf[R2].String()),
		log.String("R3", rf[R3].String()),
		log.String("R4", rf[R4].String()),
		log.String("R5", rf[R5].String()),
		log.String("R6", rf[R6].String()),
		log.String("R7", rf[R7].String()),
	)
}

// An OptionFn is modifies the machine during initialization. The function is called twice:
type OptionFn func(maching *LC3, late bool)

// WithSystemContext initializes the machine to use system context, i.e. with system privileges and
// stack.
func WithSystemContext() OptionFn {
	return func(vm *LC3, late bool) {
		vm.PSR &^= (StatusPrivilege & StatusUser)
		vm.REG[SP] = vm.SSP
	}
}

// WithDisplay is an option function that configures a callback that is called for displayed words.
// It uses late initialization under the assumption startup output is not listened for.
func WithDisplayListener(listener func(uint16)) OptionFn {
	return func(vm *LC3, late bool) {
		if late {
			driver := vm.Mem.Devices.Get(DDRAddr).(*DisplayDriver)
			driver.Listen(listener)
		}
	}
}
