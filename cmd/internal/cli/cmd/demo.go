package cmd

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/smoynes/elsie/cmd/internal/cli"
	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

func Demo() cli.Command {
	return new(demo)
}

type demo struct {
	debug bool
}

func (demo) Help() string { return "run demo program" }

func (d *demo) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)

	fs.BoolVar(&d.debug, "debug", false, "enable debug logging")

	return fs
}

func (d demo) Run(ctx context.Context, args []string, out io.Writer, _ *log.Logger) {
	if d.debug {
		log.LogLevel.Set(slog.LevelDebug)
	}

	logger := log.NewFormattedLogger(os.Stdout)
	slog.SetDefault(logger)

	log.DefaultLogger = func() *log.Logger {
		return logger
	}

	logger.Info("Initializing machine")
	machine := vm.New(vm.WithLogger(logger))

	logger.Info("Loading trap handlers")

	var program vm.Register

	// TRAP HALT handler
	program = vm.Register(0x1000)
	machine.Mem.MAR = vm.Register(0x0025)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	// AND R0,R0,0 ; clear R0
	program = vm.Register(vm.Word(vm.AND) | 0x0020)
	machine.Mem.MAR = vm.Register(0x1000)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	// LEA R1,[MCR] ; load MCR addr into R1
	program = vm.Register(vm.Word(vm.LEA) | 0x0201)
	machine.Mem.MAR = vm.Register(0x1001)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	// STR R0,R1,0
	program = vm.Register(vm.Word(vm.STR) | 0x0040)
	machine.Mem.MAR = vm.Register(0x1002)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	// Store MCR addr
	machine.Mem.MAR = vm.Register(0x1003)
	machine.Mem.MDR = vm.Register(0xfffe)

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("Loading program")

	// TRAP HALT
	program = vm.Register(vm.Word(vm.TRAP) | vm.TrapHALT)
	machine.Mem.MAR = vm.Register(machine.PC)
	machine.Mem.MDR = program

	if err := machine.Mem.Store(); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("Starting machine")

	if err := machine.Run(ctx); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("Demo completed")
}
