package vm

import (
	"github.com/smoynes/elsie/internal/log"
)

// WithLogger is an option function that configures the VM to log to a particular logger.
func WithLogger(log *log.Logger) OptionFn {
	return func(vm *LC3, late bool) {
		if !late {
			vm.updateLogger(log)
		}
	}
}

// updateLogger changes the VM's logger.
// TODO: This is weird. Sub-components should be able to reference the global logger directly.
func (vm *LC3) updateLogger(logger *log.Logger) {
	vm.log = logger
	vm.Mem.log = logger.With(log.String("subsystem", "MEM"))
	vm.Mem.Devices.log = logger.With(log.String("subsystem", "MMIO"))
	vm.INT.log = logger.With(log.String("subsystem", "INTR"))
}

// LogValue formats a log record that describes the state of the VM.
func (vm *LC3) LogValue() log.Value {
	return log.GroupValue(
		log.String("PC", vm.PC.String()),
		log.String("IR", vm.IR.String()),
		log.String("PSR", vm.PSR.String()),
		log.String("USP", vm.USP.String()),
		log.String("SSP", vm.SSP.String()),
		log.String("MCR", vm.MCR.String()),
		log.String("MAR", vm.Mem.MAR.String()),
		log.String("MDR", vm.Mem.MDR.String()),
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
