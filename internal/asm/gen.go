package asm

// gen.go contains a code generation pass for our two-pass assembler.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
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

	// Write the origin offset as the leader of the object file. The .ORIG directive should be the
	// first operation in the syntax table. However, operations may be wrapped, so we unwrap to the
	// base case, first.
	if orig, ok := unwrap(gen.syntax[0]).(*ORIG); ok {
		gen.pc = orig.LITERAL
		gen.log.Debug("Wrote object header", "ORIG", fmt.Sprintf("%0#4x", orig.LITERAL))
	} else {
		return 0, fmt.Errorf(".ORIG should be first operation; was: %T", gen.syntax[0])
	}

	for i, code := range gen.syntax {
		if code == nil {
			continue
		} else if _, ok := (unwrap(code)).(*ORIG); ok && i != 0 {
			err = errors.New(".ORIG directive may only be the first operation")
			break
		}

		encoded, err = code.Generate(gen.symbols, gen.pc)

		if err != nil {
			// If code generation caused an error, we try to get the source of the operation and
			// covert annotate the error with the source code annotation.
			if src, ok := code.(*SourceInfo); ok {
				err = &SyntaxError{
					File: src.Filename,
					Loc:  gen.pc,
					Pos:  src.Pos,
					Line: src.Line,
					Err:  err,
				}
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

// Encode generates operations encoded as word values. Unlike WriteTo, Encode does not handle
// directives, just instruction and data values.
func (gen *Generator) Encode() ([]vm.Word, error) {
	encoded := make([]vm.Word, 0, len(gen.syntax))

	for _, val := range gen.syntax {
		if code, err := val.Generate(gen.symbols, gen.pc); err != nil {
			return nil, fmt.Errorf("gen: %w", err)
		} else if len(code) != 0 {
			panic(code)
		} else {
			encoded = append(encoded, vm.Word(code[0]))
		}
	}

	return encoded, nil
}

// unwrap returns the base operation from possibly wrapped operation.
func unwrap(oper Operation) Operation {
	for {
		if wrap, ok := oper.(interface{ Unwrap() Operation }); ok {
			oper = wrap.Unwrap()
		} else {
			return oper
		}
	}
}
