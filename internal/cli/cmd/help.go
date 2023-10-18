package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
)

type help struct {
	cmd []cli.Command
}

var _ cli.Command = (*help)(nil)

func (help) Description() string {
	return "display help for commands"
}

func (h help) FlagSet() *cli.FlagSet {
	return flag.NewFlagSet("help", flag.ExitOnError)
}

func (h help) Run(_ context.Context, args []string, out io.Writer, log *log.Logger) int {
	if len(args) == 1 {
		for _, cmd := range h.cmd {
			if args[0] == cmd.FlagSet().Name() {
				h.printCommandHelp(cmd)
			}
		}
	} else {
		out := flag.CommandLine.Output()
		if err := h.Usage(out); err != nil {
			return 1
		}
	}

	return 0
}

func (h *help) Usage(out io.Writer) error {
	_, err := fmt.Fprintln(out, `
ELSIE is a virtual machine and programming tool for the LC-3 educational computer.

Usage:

        elsie <command> [option]... [arg]...

Commands:`)
	if err != nil {
		return err
	}

	for _, cmd := range h.cmd {
		fs := cmd.FlagSet()
		fmt.Fprintf(out, "  %-20s %s\n", fs.Name(), cmd.Description())
	}

	fmt.Fprintf(out, "  %-20s %s\n", h.FlagSet().Name(), h.Description())
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Use `elsie help <command>` to get help for a command.")

	return err
}

func (h *help) printCommandHelp(cmd cli.Command) {
	out := flag.CommandLine.Output()
	_ = cmd.FlagSet().Parse(nil)

	fmt.Fprint(out, "Usage:\n\n        elsie ")

	if err := cmd.Usage(out); err != nil {
		return
	}

	fmt.Fprintln(out, "\nOptions:")
	cmd.FlagSet().PrintDefaults()
}

func Help(cmd []cli.Command) *help {
	return &help{
		cmd: cmd,
	}
}
