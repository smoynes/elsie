// Package monitor implements a system monitor or BIOS for the machine.
package monitor

import (
	"fmt"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

// WithSystemImage initializes the machine with a given image.
func WithSystemImage(image *SystemImage) vm.OptionFn {
	return func(machine *vm.LC3, late bool) {
		if !late {
			return
		}

		loader := vm.NewLoader(machine)
		err := loadImage(loader, image)

		if err != nil {
			panic(err)
		}
	}
}

// WithDefaultSystemImage initializes the machine with the default system image. You should probably
// use this.
func WithDefaultSystemImage() vm.OptionFn {
	return func(machine *vm.LC3, late bool) {
		if late {
			logger := log.DefaultLogger()
			image := NewSystemImage(logger)
			loader := vm.NewLoader(machine)

			if err := loadImage(loader, image); err != nil {
				panic(err)
			}
		}
	}
}

// SystemImage holds the initial state of memory for the machine. After construction, the image is
// loaded into the machine using the poorly named LoadTo function.
type SystemImage struct {
	Symbols    asm.SymbolTable // System or monitor symbol table.
	Data       vm.ObjectCode   // System data, globally shared among all routines.
	Traps      []Routine       // System calls are called from user context to do basic I/O.
	ISRs       []Routine       // Interrupt service routines are called from interrupt context.
	Exceptions []Routine       // Exception handlers are called in response to program faults.

	logger *log.Logger
}

// Routine represents a system-defined system handler. Each routine's code is stored at an origin
// offset. The machine dispatches to the routine using an entry in a vector table.
type Routine struct {
	Name    string          // Debug friend.
	Vector  vm.Word         // Vector table-entry.
	Orig    vm.Word         // Origin-offset address.
	Code    []asm.Operation // Code and data.
	Symbols asm.SymbolTable // Routine symbols.
}

// NewSystemImage creates a default system image including basic I/O system calls and exception
// handlers.
func NewSystemImage(logger *log.Logger) *SystemImage {
	data := vm.ObjectCode{
		Orig: 0x0500,
		Code: []vm.Word{},
	}

	sym := asm.SymbolTable{} // TODO: No global symbols.

	return &SystemImage{
		Symbols: sym,
		Data:    data,
		Traps: []Routine{
			TrapHalt,
			TrapOut,
			TrapPuts,
		},
		ISRs:       []Routine{},
		Exceptions: []Routine{},
		logger:     logger,
	}
}

// GenerateRoutine takes a monitor routine, i.e. a trap, interrupt, or exception handler, and
// generates the code for it.
func GenerateRoutine(routine Routine) (vm.ObjectCode, error) {
	obj := vm.ObjectCode{
		Orig: routine.Orig,
		Code: make([]vm.Word, 0, len(routine.Code)),
	}

	pc := routine.Orig

	for _, oper := range routine.Code {
		if oper == nil {
			return obj, fmt.Errorf("generate: operation is nil")
		}

		encoded, err := oper.Generate(routine.Symbols, pc)

		if err != nil {
			return obj, fmt.Errorf("generate: %s: %w", oper, err)
		}

		obj.Code = append(obj.Code, encoded...)
		pc += vm.Word(len(encoded))
	}


	return obj, nil
}

func loadImage(loader *vm.Loader, image *SystemImage) error {
	for _, trap := range image.Traps {
		image.logger.Debug("loading trap", "TRAP", trap.Name)

		obj, err := GenerateRoutine(trap)
		if err != nil {
			return err
		}

		_, err = loader.LoadVector(trap.Vector, obj)
		if err != nil {
			return err
		}
	}

	// TODO: load data, ISRs, exceptions
	return nil
}
