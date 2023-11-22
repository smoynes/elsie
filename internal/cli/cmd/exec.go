package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/encoding"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/monitor"
	"github.com/smoynes/elsie/internal/vm"
)

func Executor() cli.Command {
	exec := &executor{log: log.DefaultLogger()}
	return exec
}

type executor struct {
	logLevel slog.Level
	log      *log.Logger
}

func (executor) Description() string {
	return "run a program"
}

func (executor) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `exec program.bin

Runs an executable in the emulator.`)

	return err
}

func (ex *executor) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	fs.Func("loglevel", "set log `level`", func(s string) error {
		return ex.logLevel.UnmarshalText([]byte(s))
	})

	return fs
}

// Run executes the program.
func (ex *executor) Run(ctx context.Context, args []string, stdout io.Writer, logger *log.Logger,
) int {
	log.LogLevel.Set(ex.logLevel)

	// Code translated is encoded in a hex-based encoding.
	code, err := ex.loadCode(args[0])
	if err != nil {
		logger.Error("Error loading code", "err", err)
		return -1
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	ctx, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTimeout()

	logger.Debug("Initializing machine")

	dispCh := make(chan rune, 1)

	machine := vm.New(
		vm.WithLogger(logger),
		monitor.WithDefaultSystemImage(),
		vm.WithDisplayListener(func(displayed uint16) {
			dispCh <- rune(displayed)
		}),
	)

	loader := vm.NewLoader(machine)
	count := uint16(0)

	for i := range code {
		n, err := loader.Load(code[i])
		count += n

		if err != nil {
			logger.Error(err.Error())
			return 1
		}
	}

	go func() {
		logger.Debug("Starting display")

		for {
			select {
			case disp := <-dispCh:
				fmt.Printf("%c", disp)
			case <-ctx.Done():
				return
			}
		}
	}()

	logger.Debug("Loaded program", "file", args[0], "loaded", count)

	go func(cancel context.CancelCauseFunc) {
		logger.Info("Starting machine")

		err := machine.Run(ctx)

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			logger.Warn("Demo timeout")
			return
		case err != nil:
			logger.Error(err.Error())
			cancel(err)

			return
		default:
			cancel(context.Canceled)
		}
	}(cancel)

	<-ctx.Done()

	close(dispCh)

	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		logger.Error("Exec timeout!")
		return 2
	} else if errors.Is(err, context.Canceled) {
		logger.Info("Program completed")
		return 0
	} else if err != nil {
		logger.Error("Program error", "ERR", err)
		return 2
	} else {
		logger.Info("Terminated")
		return 0
	}
}

func (ex executor) loadCode(fn string) ([]vm.ObjectCode, error) {
	ex.log.Debug("Loading executable", "file", fn)

	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	code, err := io.ReadAll(file)
	if err != nil {
		ex.log.Error(err.Error())
		return nil, err
	}

	ex.log.Debug("Loaded file", "bytes", len(code))

	hex := encoding.HexEncoding{}

	if err = hex.UnmarshalText(code); err != nil {
		ex.log.Error(err.Error())
		return nil, err
	}

	return hex.Code, nil
}
