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

// Demo is a demonstration command. It serves as a smoke test for the VM and an example for
// developers.
func Demo() cli.Command {
	return new(demo)
}

type demo struct {
	log   bool
	debug bool
}

func (demo) Description() string {
	return "run demo program"
}

func (d demo) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `
demo [ -log | -debug ]

Run demonstration program.`)

	return err
}

func (d *demo) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)

	fs.BoolVar(&d.log, "log", false, "log execution state")
	fs.BoolVar(&d.debug, "debug", false, "verbose execution state")

	return fs
}

func (d demo) Run(ctx context.Context, args []string, out io.Writer, _ *log.Logger) int {
	// When the context is cancelled the machine will stop running.
	ctx, done := context.WithCancel(ctx)
	defer done()

	// We expect it to take much less than 1 second to run the demo. If it takes much longer,
	// something is wrong.
	ctx, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)
	defer cancelTimeout()

	// For the demo, we log to the error stream.
	logger := d.configureLogger(os.Stderr)

	logger.Info("Initializing machine")

	// Use a channel to send displayed values to a background thread.
	dispCh := make(chan uint16)

	// Create virtual machine.
	machine := vm.New(
		// Use default BIOS.
		monitor.WithDefaultSystemImage(),

		// Log using the configured logger.
		vm.WithLogger(logger),

		// Write displayed values to the display channel.
		vm.WithDisplayListener(func(displayed uint16) {
			dispCh <- displayed
		}),
	)

	logger.Info("Loading program")

	// Load the demo program.
	loader := vm.NewLoader(machine)
	machine.REG[vm.R0] = 0x2364 // ⍤
	code := vm.ObjectCode{
		Orig: 0x3000,
		Code: []vm.Word{
			vm.Word(vm.NewInstruction(vm.TRAP, uint16(vm.TrapOUT))),
			vm.Word(vm.NewInstruction(vm.TRAP, uint16(vm.TrapOUT))),
			vm.Word(vm.NewInstruction(vm.TRAP, uint16(vm.TrapHALT))),
		},
	}

	if _, err := loader.Load(code); err != nil {
		logger.Error("error loading code:", err)
		return 2
	}

	// Start a background thread to displays each character after a brief delay.
	go func() {
		logger.Info("Starting display")

		timer := time.NewTicker(80 * time.Millisecond)
		defer timer.Stop()

		for {
			select {
			case disp := <-dispCh:
				r := rune(disp)
				fmt.Printf("%c", r)
				<-timer.C
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		logger.Info("Starting machine")

		err := machine.Run(ctx)

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			logger.Warn("Demo timeout")
			return
		case err != nil:
			logger.Error(err.Error())
		}

		done()
	}()

	<-ctx.Done()

	close(dispCh)

	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		logger.Error("Demo timeout!")
	} else if err != nil {
		logger.Error("Demo error!", "ERR", err)
	} else {
		logger.Info("Demo completed")
	}

	return 0
}

func (d demo) configureLogger(out io.Writer) *log.Logger {
	logger := log.NewFormattedLogger(out)
	log.SetDefault(logger)
	log.DefaultLogger = func() *log.Logger {
		return logger
	}

	switch {
	case d.debug == true:
		log.LogLevel.Set(log.Debug)
	case d.log == true:
		log.LogLevel.Set(log.Info)
	default:
		log.LogLevel.Set(log.Error)
	}

	return logger
}
