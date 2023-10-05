// Package cli contains the command-line interface.
package cli

import (
	"context"
	"flag"
	"io"
	"os"

	"github.com/smoynes/elsie/internal/log"
)

// Command represents a sub-command in the CLI. Each sub-command can have their own flags, config
// and action to perform.
type Command interface {
	// FlagSet returns a set of command options the command accepts.
	FlagSet() *flag.FlagSet

	// Description returns a brief description of the command's function.
	Description() string

	// Usage prints detailed command documentation.
	Usage(out io.Writer) error

	// Run executes the command with arguments. Command output should be written to |out|. It
	// returns an exit code. TODO: Should be an enum, instead of an exit code.
	Run(ctx context.Context, args []string, out io.Writer, logger *log.Logger) int
}

// Commander is a CLI command-runner that handles the life cycle of a CLI command execution.
type Commander struct {
	ctx context.Context
	log *log.Logger

	help     Command
	commands []Command
}

// New creates a new |Commander| that can start sub-commands.
func New(ctx context.Context) *Commander {
	return &Commander{
		ctx: ctx,
	}
}

// Execute runs a command, if configured.
func (cli *Commander) Execute(args []string) int {
	// If the CLI is started with no argumens, use the default "help" command.
	if len(args) == 0 {
		flag.Parse()
		cli.help.Run(cli.ctx, nil, os.Stdout, cli.log)

		return 1
	}

	// Find a command with the same name as the word on the CLI arguments.
	found := cli.help // Default, if no match.

	for _, cmd := range cli.commands {
		if args[0] == cmd.FlagSet().Name() {
			found = cmd
		}
	}

	// We found our command to run (or the help command). Now, we slice off the first argument, the
	// program name, and parse the command's flags.
	fs := found.FlagSet()
	args = args[1:]

	if err := fs.Parse(args); err != nil {
		cli.log.Error("parse error", "err", err)
		return 1
	}

	return found.Run(cli.ctx, fs.Args(), os.Stdout, cli.log)
}

// WithCommands adds a list of commands as sub-commands.
func (cli *Commander) WithCommands(cmds []Command) *Commander {
	cli.commands = append([]Command(nil), cmds...)
	return cli
}

// WithHelp configures the help message a command.
func (cli *Commander) WithHelp(cmd Command) *Commander {
	cli.help = cmd
	return cli
}

// WithLogger configures the logger for the CLI. Logs are written to os.Stderr to leave os.Stdout
// for program output.
func (cli *Commander) WithLogger(out *os.File) *Commander {
	logger := log.NewFormattedLogger(os.Stderr)
	cli.log = logger

	log.SetDefault(logger)

	return cli
}

// Type aliases from std lib.
type (
	Flag    = flag.Flag
	FlagSet = flag.FlagSet
)
