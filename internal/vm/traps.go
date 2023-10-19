package vm

// traps.go defines trap handlers or system calls.

// initializeTrapHandlers loads default trap handlers.
func (vm *LC3) initializeTrapHandlers() {
	var err error

	loader := NewLoader(vm)

	vm.log.Debug("Loading trap handlers", "traps", []Word{TrapHALT})

	count, err := loader.Load(trapHaltHandler)
	if err != nil {
		panic(err)
	}

	c, err := loader.Load(trapHaltVector)
	if err != nil {
		panic(err)
	}

	count += c

	vm.log.Debug("Loaded trap handlers", "size", count)
}

var trapHaltVector = ObjectCode{
	Orig: TrapTable + TrapHALT,
	Code: []Word{0x1000},
}

// TODO: remove for system image
var trapHaltHandler = ObjectCode{
	Orig: 0x1000,
	Code: []Word{
		/* 0x1000 */ Word(NewInstruction(AND, 0x0020)), // AND R0,R0,0  ; Clear R0.
		/* 0x1001 */ Word(NewInstruction(LEA, 0x0201)), // LEA R1,[MCR] ; Load MCR addr into R1.
		/* 0x1002 */ Word(NewInstruction(STR, 0x0040)), // STR R0,R1,#0 ; Write R0 to MCR addr.
		/* 0x1003 */ 0xfffe, // ; MCR addr.
	},
}
