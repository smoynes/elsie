package vm

import (
	"github.com/smoynes/elsie/internal/log"
)

// WithLogger is an option function that configures the VM to log to a particular logger.
func WithLogger(log *log.Logger) OptionFn {
	return func(vm *LC3) {
		vm.withLogger(log)
	}
}

// TODO: This is weird.
func (vm *LC3) withLogger(log *log.Logger) {
	vm.log = log
	vm.Mem.log = log
	vm.Mem.Devices.log = log
	vm.INT.log = log
}

func (vm *LC3) LogValue() log.Value {
	return log.GroupValue(
		log.String("PC", vm.PC.String()),
		log.String("IR", vm.IR.String()),
		log.String("PSR", vm.PSR.String()),
		log.String("USP", vm.USP.String()),
		log.String("SSP", vm.SSP.String()),
		log.String("MCR", vm.MCR.String()),
		log.Any("INT", vm.INT),
		log.Any("REG", vm.REG),
	)
}

func (mmio *MMIO) WithLogger(l *log.Logger) {
	mmio.log = l

	for _, dev := range mmio.devs {
		if dev, ok := dev.(log.Loggable); ok {
			dev.WithLogger(l)
		}
	}
}
