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
	return new(executor)
}

type executor struct {
	debug bool
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
func (a *executor) Run(ctx context.Context, args []string, stdout io.Writer, logger *log.Logger) int {
	if a.debug {
		log.LogLevel.Set(log.Debug)
	}

	logger.Debug("Loading executable", "file", args[0])

	file, err := os.Open(args[0])
	if err != nil {
		logger.Error(err.Error())
		return 1 // I/O error
	}

	program, err := io.ReadAll(file)
	if err != nil {
		logger.Error(err.Error())
		return 1 // I/O error
	}

	code := vm.ObjectCode{}
	read, err := code.Read(program)

	if err != nil {
		logger.Error(err.Error())
		return 2 // Code load error
	}

	logger.Debug("Initializing machine")
	machine := vm.New()

	loader := vm.NewLoader()
	count, err := loader.Load(machine, code)

	if err != nil {
		logger.Error(err.Error())
		return 2 // Code load error
	}

	logger.Debug("Loaded program", "file", args[0], "bytes", read, "loaded", count)
	logger.Debug("Starting machine")

	err = machine.Run(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 3 // Exec error
	}

	return 0
}
