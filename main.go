// cmd/elsie is the command-line interface to the ELSIE, an LC-3 simulator and tool suite.
package main

import (
	"context"
	"os"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/cli/cmd"
)

var (
	commands = []cli.Command{
		cmd.Demo(),
	}
)

// Entry point.
func main() {
	result :=
		cli.New(context.Background()).
			WithLogger(os.Stderr).
			WithCommands(commands).
			WithHelp(cmd.Help(commands)).
			Execute(os.Args[1:])

	os.Exit(result)
}
