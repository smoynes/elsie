package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/monitor"
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

	logger.Info("Loading trap handler")

	loader := vm.NewLoader()
	halt := monitor.TrapHalt{}
	vector, handler := halt.Vector()
	_, err := loader.LoadVector(machine, vector, handler)

	logger.Info("Loading program")

	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(vm.TRAP, uint16(vm.TrapHALT))),
		},
	}

	_, err = loader.Load(machine, code)
	if err != nil {
		logger.Error(err.Error())
		return 2
	}

	logger.Info("Starting machine")

	if err := machine.Run(ctx); err != nil {
		logger.Error(err.Error())
		return 2
	}

	logger.Info("Demo completed")

	return 0
}
