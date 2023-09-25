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
func (vm *LC3) withLogger(logger *log.Logger) {
	vm.log = logger
	vm.Mem.log = logger.With(log.String("subsytem", "MEM"))
	vm.Mem.Devices.log = logger.With(log.String("subsytem", "MMIO"))
	vm.INT.log = logger.With(log.String("subystem", "INTR"))
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

func (mmio *MMIO) WithLogger(logger *log.Logger) {
	mmio.log = logger.With("subsystem", "IO")

	for _, dev := range mmio.devs {
		if dev, ok := dev.(log.Loggable); ok {
			dev.WithLogger(logger.With("subsystem", "DEVICE"))
		}
	}
}
