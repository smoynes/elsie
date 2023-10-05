package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/smoynes/elsie/internal/asm"
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
	_, err = fmt.Fprintln(out, `asm [-o file.out] file.asm

Assemble source into object code.`)

	return err
}

func (a *assembler) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("asm", flag.ExitOnError)
	fs.BoolVar(&a.debug, "debug", false, "enable debug logging")
	fs.StringVar(&a.output, "o", "a.out", "output `filename`")

	return fs
}

// Run calls the assembler to assemble the assembly.
func (a *assembler) Run(ctx context.Context, args []string, out io.Writer, logger *log.Logger) int {
	if a.debug {
		log.LogLevel.Set(log.Debug)
	}

	// First pass: parse source and create symbol table.
	parser := asm.NewParser(logger)

	for i := range args {
		fn := args[i]

		f, err := os.Open(fn)
		if err != nil {
			logger.Error("Parse error: %s: %s", fn, err)
		}

		parser.Parse(f)
	}

	logger.Info("Parsed source", "symbols", parser.Symbols(), "err", parser.Err())

	if parser.Err() != nil {
		return 1
	}

	syntax := parser.Syntax()

	for pc, code := range syntax {
		if code == nil {
			continue
		}

		mc, err := code.Generate(parser.Symbols(), uint16(pc))
		logger.Info("Parsed",
			"pc", fmt.Sprintf("%0#4x", uint16(pc)),
			"code", fmt.Sprintf("%#v", code),
			"code", code,
			"gen", fmt.Sprintf("%0#4x", mc),
			"err", err)
	}

	// TODO: second pass
	return 0
}
