package asm

// gen.go contains a code generation pass for our two-pass assembler.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/log"
)

// Generator controls the code generation pass of the assembler. The generator starts at the
// beginning of the parsed-syntax table, generates code for each operation, and then writes the
// bytes to the output (usually, a file).
//
// During the generation pass, any syntax or semantic errors that prevent generating machine code
// are immediately returned from WriteTo. The errors are wrapped SyntaxErrors and may be tested and
// retrieved using the errors package.
type Generator struct {
	pc      uint16
	symbols SymbolTable
	syntax  SyntaxTable

	log *log.Logger
}

// NewGenerator creates a code generator using the given symbol and syntax tables.
func NewGenerator(symbols SymbolTable, syntax SyntaxTable) *Generator {
	return &Generator{
		pc:      0x0000,
		symbols: symbols,
		syntax:  syntax,
		log:     log.DefaultLogger(),
	}
}

// WriteTo writes generated machine code to an output stream.
func (gen *Generator) WriteTo(out io.Writer) (int64, error) {
	var (
		encoded []uint16
		count   int64
		err     error
	)

	if len(gen.syntax) == 0 {
		return 0, nil
	}

	// Write the object-code header: the origin offset. The .ORIG directive should be the first
	// operation in the syntax table.
	if orig := origin(gen.syntax[0]); orig == nil {
		return 0, errors.New("gen: .ORIG directive must be the first operation")
	} else {
		gen.pc = orig.LITERAL
	}

	for i, code := range gen.syntax {
		if code == nil {
			continue
		} else if origin(code) != nil && i != 0 {
			err = errors.New("gen: .ORIG directive may only be the first operation")
			break
		}

		encoded, err = code.Generate(gen.symbols, gen.pc)

		if err != nil {
			src := code.Source()
			err = &SyntaxError{
				File: src.Filename,
				Loc:  gen.pc,
				Pos:  src.Pos,
				Line: src.Line,
				Err:  err,
			}

			break
		}

		if err = binary.Write(out, binary.BigEndian, encoded); err != nil {
			break
		}

		gen.pc += uint16(len(encoded))
		count += int64(len(encoded) * 2)
	}

	if err != nil {
		return count, fmt.Errorf("gen: %w", err)
	}

	return count, nil
}

func origin(op Operation) *ORIG {
	if op, ok := op.(*ORIG); ok {
		return op
	}

	return nil
}
