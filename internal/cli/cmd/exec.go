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
	"github.com/smoynes/elsie/internal/encoding"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/monitor"
	"github.com/smoynes/elsie/internal/vm"
)

func Executor() cli.Command {
	exec := &executor{
		logger: log.DefaultLogger(),
	}

	return exec
}

type executor struct {
	logger *log.Logger // Log destination
	log    string      // Log output path
	debug  string      // Debug log path
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

	fs.StringVar(&ex.log, "log", "", "write log to `file`")
	fs.StringVar(&ex.debug, "debug", "", "write debug log `file`")

	return fs
}

// Run executes the program.
func (ex *executor) Run(ctx context.Context, args []string, stdout io.Writer, logger *log.Logger,
) int {
	if len(args) == 0 {
		logger.Error("Missing object-code argument. Run elsie help exec for usage.")
		return -1
	}

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

	var (
		logFile  = os.Stderr
		logLevel = log.Error
	)

	if ex.debug != "" {
		if logFile, err = os.OpenFile(ex.debug, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err != nil {
			err = fmt.Errorf("%s: %w", ex.debug, err)
		}
	} else if ex.log != "" {
		if logFile, err = os.OpenFile(ex.log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err != nil {
			err = fmt.Errorf("%s: %w", ex.log, err)
		}
	}

	if err != nil {
		ex.logger.Error(err.Error())
		logger.Error(err.Error())

		return -1
	}

	defer logFile.Close()

	log.LogLevel.Set(logLevel)
	ex.logger = log.NewFormattedLogger(os.Stderr)
	ex.logger.Debug("Initializing machine")

	dispCh := make(chan rune, 1)
	machine := vm.New(
		vm.WithLogger(ex.logger),
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
			ex.logger.Error(err.Error())
			return 1
		}
	}

	go func() {
		ex.logger.Debug("Starting display")

		for {
			select {
			case disp := <-dispCh:
				fmt.Printf("%c", disp)
			case <-ctx.Done():
				return
			}
		}
	}()

	ex.logger.Debug("Loaded program", "file", args[0], "loaded", count)

	go func(cancel context.CancelCauseFunc) {
		ex.logger.Info("Starting machine")

		err := machine.Run(ctx)

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			ex.logger.Warn("Exec timeout")
			return
		case err != nil:
			ex.logger.Error(err.Error())
			cancel(err)

			return
		default:
			cancel(context.Canceled)
		}
	}(cancel)

	<-ctx.Done()

	close(dispCh)

	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		ex.logger.Error("Execution timeout")
		logger.Error("Execution timeout")

		return 2
	} else if errors.Is(err, context.Canceled) {
		ex.logger.Debug("Program completed")
		logger.Debug("Program completed")
		return 0
	} else if err != nil {
		ex.logger.Error("Program error", "ERR", err)
		logger.Error("Program error", "ERR", err)
		return 2
	} else {
		ex.logger.Info("Terminated")
		logger.Info("Terminated")
		return 0
	}
}

func (ex executor) loadCode(fn string) ([]vm.ObjectCode, error) {
	ex.logger.Debug("Loading executable", "file", fn)

	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	code, err := io.ReadAll(file)
	if err != nil {
		ex.logger.Error(err.Error())
		return nil, err
	}

	ex.logger.Debug("Loaded file", "bytes", len(code))

	hex := encoding.HexEncoding{}

	if err = hex.UnmarshalText(code); err != nil {
		ex.logger.Error(err.Error())
		return nil, err
	}

	return hex.Code, nil
}
