package monitor

import (
	"fmt"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

type vec struct {
	vector vm.Word
	orig   vm.Word
	code   []asm.Operation
}

// SystemImage holds the initial state of memory for the machine.
type SystemImage struct {
	Symbols    asm.SymbolTable // System symbol table.
	Traps      []vec           // System calls.
	ISRs       []vec           // Interrupt service routines.
	Exceptions []vec           // Exception handlers.
	log        *log.Logger
}

// NewSystemImage creates a default system image including basic I/O system calls and exception
// handlers.
func NewSystemImage() *SystemImage {
	logger := log.DefaultLogger()
	return &SystemImage{
		Symbols:    asm.SymbolTable{},
		Traps:      []vec{TrapHalt},
		ISRs:       []vec{},
		Exceptions: []vec{},
		log:        logger,
	}
}

// LoadTo uses a loader to initialize the machine with the system image.
func (img *SystemImage) LoadTo(loader *vm.Loader) (uint16, error) {
	var (
		err   error
		count uint16
	)

	img.log.Debug("Loading traps handlers")
	for _, trap := range img.Traps {
		pc := trap.orig

		obj := vm.ObjectCode{
			Orig: trap.orig,
			Code: make([]vm.Word, 0, len(trap.code)),
		}

		img.log.Debug("Generating code",
			"orig", fmt.Sprintf("%0#4x", trap.orig),
			"size", len(trap.code),
		)

		for _, op := range trap.code {
			if op == nil {
				continue
			}

			encoded, err := op.Generate(img.Symbols, uint16(pc))
			if err != nil {
				break
			}

			for i := range encoded {
				obj.Code = append(obj.Code, vm.Word(encoded[i]))
			}
		}

		img.log.Debug("Loading code",
			"orig", trap.orig,
			"vector", trap.vector,
			"size", len(obj.Code),
		)

		loader.LoadVector(trap.vector, obj)

		if c, err := loader.Load(obj); err != nil {
			break
		} else {
			count += c
			pc += 1
		}

	}

	if err != nil {
		return count, err
	}

	return count, nil
}
