// Package monitor implements a system monitor or BIOS for the machine.
package monitor

import (
	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

// WithSystemImage initializes the machine with a given image.
func WithSystemImage(image *SystemImage) vm.OptionFn {
	return func(machine *vm.LC3, late bool) {
		if late {
			loader := vm.NewLoader(machine)

			if _, err := image.LoadTo(loader); err != nil {
				panic(err) // TODO: return error
			}
		}
	}
}

// WithDefaultSystemImage initializes the machine with the default system image. You should probably
// use this.
func WithDefaultSystemImage() vm.OptionFn {
	return WithSystemImage(NewSystemImage())
}

// SystemImage holds the initial state of memory for the machine.
type SystemImage struct {
	Symbols    asm.SymbolTable // Static symbol table.
	Data       vm.ObjectCode   // System data, globally shared among all routines.
	Traps      []Routine       // System calls are called from user context to do basic I/O.
	ISRs       []Routine       // Interrupt service routines are called from interrupt context.
	Exceptions []Routine       // Exception handlers are called in response to program faults.

	log *log.Logger
}

// Routine represents a system-defined system handler. Each routine's code is stored at an origin
// offset. The machine dispatches to the routine using an entry in a vector table.
type Routine struct {
	Vector vm.Word         // Vector table-entry.
	Orig   vm.Word         // Origin-offset address.
	Code   []asm.Operation // Code and data.
}

// NewSystemImage creates a default system image including basic I/O system calls and exception
// handlers.
func NewSystemImage() *SystemImage {
	logger := log.DefaultLogger()

	data := vm.ObjectCode{
		Orig: 0x0500,
		Code: []vm.Word{
			vm.Word('\n'),
			vm.Word('b'), vm.Word('y'), vm.Word('e'), 0,
		},
	}
	sym := asm.SymbolTable{}
	sym["ASCIINEWLINE"] = 0x0500
	sym["HALTMESSAGE"] = 0x0501

	return &SystemImage{
		Symbols:    sym,
		Data:       data,
		Traps:      []Routine{TrapHalt},
		ISRs:       []Routine{},
		Exceptions: []Routine{},
		log:        logger,
	}
}

// LoadTo uses a loader to initialize the machine with the system image.
func (img *SystemImage) LoadTo(loader *vm.Loader) (uint16, error) {
	var count uint16

	img.log.Debug("Loading trap handlers")
	for _, trap := range img.Traps {
		img.log.Debug("Generating code",
			"orig", trap.Orig,
			"size", len(trap.Code),
		)

		pc := trap.Orig
		obj := vm.ObjectCode{
			Orig: trap.Orig,
			Code: make([]vm.Word, 0, len(trap.Code)),
		}

		for _, op := range trap.Code {
			if op == nil {
				continue
			}

			encoded, err := op.Generate(img.Symbols, uint16(pc))
			if err != nil {
				return count, err
			}

			for i := range encoded {
				obj.Code = append(obj.Code, vm.Word(encoded[i]))
			}

			pc += 1
		}

		img.log.Debug("Loading code",
			"orig", trap.Orig,
			"vector", trap.Vector,
			"size", len(obj.Code),
		)

		if c, err := loader.LoadVector(trap.Vector, obj); err != nil {
			return count, err
		} else {
			count += c
		}
	}

	return count, nil
}
