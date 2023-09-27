package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
)

func Assembler() cli.Command {
	return new(assembler)
}

type assembler struct {
	debug  bool
	output string
}

func (assembler) Description() string {
	return "assemble source code into object code"
}

func (assembler) Usage(out io.Writer) error {
	var err error
	_, err = fmt.Fprintln(out, `asm file...

Run demonstration program while displaying VM state.`)

	return err
}

func (a *assembler) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("asm", flag.ExitOnError)
	fs.BoolVar(&a.debug, "debug", false, "enable debug logging")
	fs.StringVar(&a.output, "o", "a.out", "output `filename`")

	return fs
}

func (a *assembler) Run(ctx context.Context, args []string, out io.Writer, logger *log.Logger) int {
	return 1
}
