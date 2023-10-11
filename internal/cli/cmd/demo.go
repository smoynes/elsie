package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

func Demo() cli.Command {
	return new(demo)
}

type demo struct {
	debug bool
}

func (demo) Description() string {
	return "run demo program"
}

func (d demo) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `
demo [-debug]

Run demonstration program while displaying VM state.`)

	return err
}

func (d *demo) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)

	fs.BoolVar(&d.debug, "debug", false, "enable debug logging")

	return fs
}

func (d demo) Run(ctx context.Context, args []string, out io.Writer, _ *log.Logger) int {
	if d.debug {
		log.LogLevel.Set(log.Debug)
	}

	logger := log.NewFormattedLogger(os.Stdout)
	log.SetDefault(logger)

	log.DefaultLogger = func() *log.Logger {
		return logger
	}

	logger.Info("Initializing machine")
	machine := vm.New(vm.WithLogger(logger))

	logger.Info("Loading trap handlers")

	loader := vm.NewLoader()
	haltHandler := vm.ObjectCode{
		Orig: 0x1000,
		Code: []vm.Instruction{
			/* 0x1000 */ vm.NewInstruction(vm.AND, 0x0020), // AND R0,R0,0  ; Clear R0.
			/* 0x1001 */ vm.NewInstruction(vm.LEA, 0x0201), // LEA R1,[MCR] ; Load MCR addr into R1.
			/* 0x1002 */ vm.NewInstruction(vm.STR, 0x0040), // STR R0,R1,#0 ; Write R0 to MCR addr.
			/* 0x1003 */ vm.Instruction(0xfffe), // ; MCR addr
		},
	}

	loader.Load(machine, haltHandler)
	var program vm.Register

	// TRAP HALT handler
	program = vm.Register(0x1000)
	machine.Mem.MAR = vm.Register(0x0025)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return 2
	}

	logger.Info("Loading program")

	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Instruction{vm.NewInstruction(vm.TRAP, uint16(vm.TrapHALT))},
	}
	loader.Load(machine, code)

	logger.Info("Starting machine")

	if err := machine.Run(ctx); err != nil {
		logger.Error(err.Error())
		return 2
	}

	logger.Info("Demo completed")

	return 0
}
