package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/monitor"
	"github.com/smoynes/elsie/internal/vm"
)

// Demo is a demonstration command.
func Demo() cli.Command {
	return new(demo)
}

type demo struct {
	debug bool
	quiet bool
}

func (demo) Description() string {
	return "run demo program"
}

func (d demo) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `
demo [ -debug | -quiet ]

Run demonstration program while displaying VM state.`)

	return err
}

func (d *demo) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)

	fs.BoolVar(&d.debug, "debug", false, "enable debug logging")
	fs.BoolVar(&d.quiet, "quiet", false, "enable quiet output, machine display only")

	return fs
}

func (d demo) Run(ctx context.Context, args []string, out io.Writer, _ *log.Logger) int {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if d.quiet {
		log.LogLevel.Set(log.Error)
	}

	if d.debug {
		log.LogLevel.Set(log.Debug)
	}

	logger := log.NewFormattedLogger(os.Stdout)
	log.SetDefault(logger)
	log.DefaultLogger = func() *log.Logger {
		return logger
	}

	logger.Info("Initializing machine")

	dispCh := make(chan uint16)
	machine := vm.New(
		vm.WithLogger(logger),
		vm.WithDisplayListener(func(displayed uint16) {
			dispCh <- displayed
		}),
		monitor.WithDefaultSystemImage(),
	)

	logger.Info("Loading program")

	loader := vm.NewLoader(machine)
	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(vm.TRAP, uint16(vm.TrapHALT))),
		},
	}

	if _, err := loader.Load(code); err != nil {
		logger.Error("error loading code:", err)
		return 2
	}

	go func() {
		logger.Info("Starting display")

		timer := time.NewTicker(1200 * time.Millisecond)
		defer timer.Stop()

		for {
			select {
			case disp := <-dispCh:
				r := rune(disp)
				fmt.Printf("%c", r)
			case <-ctx.Done():
				return
			}

			<-timer.C
		}
	}()

	go func() {
		logger.Info("Starting machine")

		err := machine.Run(ctx)

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			logger.Warn("Demo timeout")
			return
		default:
			logger.Error(err.Error())
		}
	}()

	<-ctx.Done()

	close(dispCh)

	logger.Info("Demo completed")

	return 0
}
