// Package cli contains the command-line interface.
package cli

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/smoynes/elsie/internal/log"
)

type Flag = flag.Flag
type FlagSet = flag.FlagSet

func New(ctx context.Context) *Commander {
	return &Commander{
		ctx: ctx,
	}
}

type Commander struct {
	ctx context.Context
	log *log.Logger

	help     Command
	commands []Command
}

func (cli *Commander) Execute(args []string) int {
	if len(args) == 0 {
		flag.Parse()
		cli.help.Run(cli.ctx, nil, os.Stdout, cli.log)
		return 1
	}

	found := cli.help
	for _, cmd := range cli.commands {
		if args[0] == cmd.FlagSet().Name() {
			found = cmd
		}
	}

	fs := found.FlagSet()

	fs.Parse(args[1:])
	found.Run(cli.ctx, fs.Args(), os.Stdout, cli.log)

	// Initialize flags
	// Parse flags
	// Delegate execution
	return 0
}

func (cli *Commander) WithCommands(cmds []Command) *Commander {
	cli.commands = append([]Command(nil), cmds...)
	return cli
}

func (cli *Commander) WithHelp(cmd Command) *Commander {
	cli.help = cmd
	return cli
}

func (cli *Commander) WithLogger(out *os.File) *Commander {
	log := log.NewFormattedLogger(os.Stderr)
	cli.log = log

	slog.SetDefault(log)

	return cli
}

type Command interface {
	FlagSet() *flag.FlagSet
	Help() string
	Run(context.Context, []string, io.Writer, *log.Logger)
}
