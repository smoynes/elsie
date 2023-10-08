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
func (a *assembler) Run(ctx context.Context, args []string, stdout io.Writer, logger *log.Logger) int {
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

	logger.Debug("Parsed source",
		"symbols", parser.Symbols().Count(),
		"size", parser.Syntax().Size(),
		"err", parser.Err(),
	)

	if parser.Err() != nil {
		logger.Error("Parse error", "err", parser.Err())
		return 1
	}

	// Second pass: generate code.
	symbols := parser.Symbols()
	syntax := parser.Syntax()

	out, err := os.Create(a.output)
	if err != nil {
		logger.Error("open failed", "out", a.output, "err", err)
		return -1
	}

	logger.Debug("Writing object", "file", a.output)

	generator := asm.NewGenerator(symbols, syntax)

	wrote, err := generator.WriteTo(out)
	if err != nil {
		logger.Error("Compile error", "out", "a.o", "err", err)
		return -1
	}

	logger.Debug("Compiled object",
		"size", wrote,
		"symbols", symbols.Count(),
		"syntax", syntax.Size(),
	)

	return 0
}
