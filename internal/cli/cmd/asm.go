package cmd

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/smoynes/elsie/internal/asm"
	"github.com/smoynes/elsie/internal/cli"
	"github.com/smoynes/elsie/internal/log"
)

// Assembler is the command that translates LCASM source code into executable object code.
//
//	elsie asm -o a.o FILE.asm
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
	_, err = fmt.Fprintln(out, `asm [-o file.o] file.asm

Assemble source into object code.`)

	return err
}

func (a *assembler) FlagSet() *cli.FlagSet {
	fs := flag.NewFlagSet("asm", flag.ExitOnError)
	fs.BoolVar(&a.debug, "debug", false, "enable debug logging")
	fs.StringVar(&a.output, "o", "a.o", "output `filename`")

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
			logger.Error("Parse error", "err", err)
			return 1
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

	out, err := os.Create(a.output)
	if err != nil {
		logger.Error("open failed", "out", a.output, "err", err)
		return -1
	}

	// Second pass: generate code.
	symbols := parser.Symbols()
	syntax := parser.Syntax()
	generator := asm.NewGenerator(symbols, syntax)

	logger.Debug("Writing object", "file", a.output)

	buf := bufio.NewWriter(out)

	objCode, err := generator.Encode()
	if err != nil {
		logger.Error("Compile error", "out", a.output, "err", err)
		return -1
	}

	wrote, err := io.Copy(buf, bytes.NewBuffer(objCode))
	if err != nil {
		logger.Error("I/O error", "out", a.output, "err", err)
		return -1
	}

	if err := buf.Flush(); err != nil {
		logger.Error("I/O error", "out", a.output, "err", err)
		return -1
	}

	logger.Debug("Compiled object",
		"out", a.output,
		"size", wrote,
		"symbols", symbols.Count(),
		"syntax", syntax.Size(),
	)

	return 0
}
