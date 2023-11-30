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
	"github.com/smoynes/elsie/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

func Executor() cli.Command {
	exec := &executor{
		logger: log.DefaultLogger(),
	}

	return exec
}

type executor struct {
	logger *log.Logger
	log    string
	debug  string
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

	fs.StringVar(&ex.log, "log", "", "enable logging to `file``")
	fs.StringVar(&ex.debug, "debug", "", "enable debug logging to `file`")

	return fs
}

// Run executes the program.
func (ex *executor) Run(ctx context.Context, args []string, stdout io.Writer, logger *log.Logger,
) int {
	var (
		err       error
		logOutput *log.Logger
	)

	if ex.log != "" {
		logOutput, err = createFileLogger(log.Info, ex.log)
	} else if ex.debug != "" {
		logOutput, err = createFileLogger(log.Debug, ex.debug)
	} else {
		log.LogLevel.Set(log.Error)
		logOutput = logger
	}

	if err != nil {
		logger.Error("Error opening log file", "err", err)
		return -1
	} else {
		logger = logOutput
	}

	logger.Info("Starting machine")

	// Translate from hex-based text-encoding to object-code sections.
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

	keyboard := vm.NewKeyboard()
	display := vm.NewDisplay()
	displayDriver := vm.NewDisplayDriver(display)

	ctx, _, cancelConsole := tty.WithConsole(ctx, keyboard, display) // TODO: logging
	defer cancelConsole()

	machine := vm.New(
		vm.WithLogger(logger),
		monitor.WithDefaultSystemImage(),
		vm.WithKeyboard(keyboard),
		vm.WithDisplay(displayDriver),
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

	logger.Debug("Loaded program", "file", args[0], "loaded", count)

	go func(cancel context.CancelCauseFunc) {
		logger.Info("Starting machine")

		err := machine.Run(ctx)

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return
		case err != nil:
			cancel(err)
			return
		default:
			cancel(context.Canceled)
		}
	}(cancel)

	<-ctx.Done()

	cancelConsole()
	cancel(context.Canceled)

	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		logger.Error("Exec timeout!")
		return 2
	} else if errors.Is(err, context.Canceled) {
		logger.Info("Program completed")
		return 0
	} else if err != nil {
		logger.Error("Program error", "ERR", err)
		return 2
	} else if perr := recover(); perr != nil {
		logger.Error("Panic!", "err", perr)
		return 2
	} else {
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

func createFileLogger(level log.Level, filename string) (*log.Logger, error) {
	println("opening", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_SYNC|os.O_CREATE|os.O_APPEND, 0o660)
	if err != nil {
		return nil, err
	}

	handler := log.NewHandler(file)
	logger := log.New(handler)
	log.SetDefault(logger)
	log.LogLevel.Set(level)

	return log.DefaultLogger(), nil
}
