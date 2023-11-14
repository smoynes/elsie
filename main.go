// 𝔼𝕃𝕊𝕀𝔼 is a virtual machine and programming tool for the LC-3 educational computer.
//
// # Usage
//
//	go run github.com/smoynes/elsie <command>
//
// Commands:
//   - exec
//   - asm
//   - demo
//   - help
package main // import "github.com/smoynes/elsie"

import (
	"context"
	"os"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/cli/cmd"
)

var commands = []cli.Command{
	cmd.Executor(),
	cmd.Assembler(),
	cmd.Demo(),
}

// Entry point.
func main() {
	result := cli.New(context.Background()).
		WithLogger(os.Stderr).
		WithCommands(commands).
		WithHelp(cmd.Help(commands)).
		Execute(os.Args[1:])

	os.Exit(result)
}
