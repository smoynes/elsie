package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

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
	hex := encoding.HexEncoding{}

	code, err := ex.loadCode(args[0])
	if err != nil {
		logger.Error("Error loading code", "err", err)
		return -1
	}

	if err = hex.wUnmarshalText(code); err != nil {
		logger.Error(err.Error())
		return -2
	}

	logger.Debug("Initializing machine")
	machine := vm.New(
		vm.WithLogger(logger),
		monitor.WithDefaultSystemImage(),
	)

	loader := vm.NewLoader(machine)
	count := uint16(0)

	for i := range hex.Code {
		n, err := loader.Load(hex.Code[i])
		count += n

		if err != nil {
			logger.Error(err.Error())
			return 1
		}
	}

	logger.Debug("Loaded program", "file", args[0], "loaded", count)
	logger.Debug("Starting machine")

	err = machine.Run(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 2 // Exec error
	}

	return 0
}

func (ex executor) loadCode(fn string) (program []byte, err error) {
	ex.log.Debug("Loading executable", "file", fn)

	file, err := os.Open(fn)
	if err != nil {
		return
	}

	program, err = io.ReadAll(file)
	if err != nil {
		ex.log.Error(err.Error())
		return
	}

	ex.log.Debug("Loaded file", "bytes", len(program))

	return program, nil
}
