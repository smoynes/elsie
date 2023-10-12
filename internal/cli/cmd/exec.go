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

func Executor() cli.Command {
	exec := &executor{
		log: log.DefaultLogger(),
	}
	return exec
}

type executor struct {
	debug bool
	log   *log.Logger
}

func (executor) Description() string {
	return "run a program"
}

func (executor) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `exec program.bin

Run program.`)

	return err
}

func (a *executor) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	fs.BoolVar(&a.debug, "debug", false, "enable debug logging")

	return fs
}

// Run executes the program.
func (ex *executor) Run(ctx context.Context, args []string, stdout io.Writer,
	logger *log.Logger,
) int {

	if ex.debug {
		log.LogLevel.Set(log.Debug)
	}

	code, err := ex.loadCode(args[0])
	if err != nil {
		logger.Error("Error loading code", "err", err)
		return 1
	}

	logger.Debug("Initializing machine")
	machine := vm.New(
		vm.WithLogger(logger),
		vm.WithTrapHandlers(),
	)

	loader := vm.NewLoader()
	count, err := loader.Load(machine, code)

	if err != nil {
		logger.Error(err.Error())
		return 1
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

func (ex executor) loadCode(fn string) (code vm.ObjectCode, err error) {
	ex.log.Debug("Loading executable", "file", fn)

	file, err := os.Open(fn)
	if err != nil {
		return
	}

	program, err := io.ReadAll(file)
	if err != nil {
		ex.log.Error(err.Error())
		return
	}

	_, err = code.Read(program)
	if err != nil {
		return
	}

	return code, nil
}
