package vm

import (
	"log"
)

type logger = *log.Logger

var defaultLogger = makeLogger

func makeLogger() logger {
	l := log.Default()
	//l := log.New(io.Discard, "", 0)
	return l
}

func (vm *LC3) withLogger(l logger) {
	vm.log = l
	vm.Mem.log = l
	vm.Mem.device.log = l
}

func (mmio *MMIO) WithLogger(l logger) {
	mmio.log = l

	for _, dev := range mmio.devs {
		if dev, ok := dev.(interface{ WithLogger(logger) }); ok {
			dev.WithLogger(l)
		}
	}
}

func WithLogger(log logger) OptionFn {
	return func(vm *LC3) {
		vm.withLogger(log)
	}
}
