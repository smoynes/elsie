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

func (help) Help() string { return "display help for commands" }

func (h help) FlagSet() *cli.FlagSet {
	return flag.NewFlagSet("help", flag.ExitOnError)
}

func (h help) Run(_ context.Context, args []string, out io.Writer, log *log.Logger) {
	if len(args) == 1 {
		for _, cmd := range h.cmd {
			if args[0] == cmd.FlagSet().Name() {
				h.printCommandHelp(cmd)
			}
		}
	} else {
		out := flag.CommandLine.Output()

		fmt.Fprintln(out, "ELSIE is a virtual machine for the LC-3 educational computer.")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "        elsie <command> [option]... [arg]...")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Commands:")

		for _, cmd := range h.cmd {
			fs := cmd.FlagSet()
			fmt.Fprintf(out, "  %-20s %s\n", fs.Name(), cmd.Help())
		}

		fmt.Fprintf(out, "  %-20s %s\n", h.FlagSet().Name(), h.Help())
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Use `elsie help <command>` to get help for a command.")
	}
}

func (h *help) printCommandHelp(cmd cli.Command) {
	out := flag.CommandLine.Output()
	_ = cmd.FlagSet().Parse(nil)

	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "        elsie demo [option]... [arg]...")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Run demonstration program while displaying VM state.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	cmd.FlagSet().PrintDefaults()
}

func Help(cmd []cli.Command) *help {
	return &help{
		cmd: cmd,
	}
}
