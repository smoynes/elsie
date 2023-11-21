package asm

// gen.go contains a code generation pass for our two-pass assembler.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/encoding"
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
	pc       uint16
	symbols  SymbolTable
	syntax   SyntaxTable
	encoding encoding.HexEncoding
	log      *log.Logger
}

// NewGenerator creates a code generator using the given symbol and syntax tables.
func NewGenerator(symbols SymbolTable, syntax SyntaxTable) *Generator {
	return &Generator{
		pc:       0x0000,
		symbols:  symbols,
		syntax:   syntax,
		encoding: encoding.HexEncoding{},
		log:      log.DefaultLogger(),
	}
}

// WriteTo writes generated machine code to an output stream. // TODO: encode output
func (gen *Generator) WriteTo(out io.Writer) (int64, error) {
	var (
		generated []uint16
		count     int64
		err       error
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

		generated, err = code.Generate(gen.symbols, gen.pc)

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

		if err = binary.Write(out, binary.BigEndian, generated); err != nil {
			break
		}

		gen.pc += uint16(len(generated))
		count += int64(len(generated) * 2)
	}

	if err != nil {
		return count, fmt.Errorf("gen: %w", err)
	}

	return count, nil
}

// Encode generates object code and encodes it as an object code file.
func (gen *Generator) Encode() ([]byte, error) {
	gen.log.Debug("encoding", "count", len(gen.syntax), "symbols", len(gen.symbols))

	if len(gen.syntax) == 0 {
		return nil, nil
	}

	var (
		obj   vm.ObjectCode
		count int64
		err   error
	)

	// We expect the .ORIG directive to be the first operation in the syntax table.
	if orig, ok := origin(gen.syntax[0]); ok {
		gen.log.Debug("object offset", "ORIG", fmt.Sprintf("%0#4x", orig.LITERAL))
		gen.pc = orig.LITERAL
		obj.Orig = vm.Word(orig.LITERAL)

	} else {
		return nil, fmt.Errorf(".ORIG should be first operation; was: %T", gen.syntax[0])
	}

	for _, op := range gen.syntax {
		if op == nil {
			continue
		} else if _, ok := origin(op); ok {
			continue
		}

		genWords, genErr := op.Generate(gen.symbols, gen.pc+1)

		if genErr != nil {
			err = gen.annotate(op, genErr)
			break
		}

		for i := range genWords {
			obj.Code = append(obj.Code, vm.Word(genWords[i]))
		}

		gen.pc += uint16(len(genWords))
		count += int64(len(genWords) * 2)
	}

	if err != nil {
		return nil, fmt.Errorf("gen: %w", err)
	}

	gen.encoding.Code = append(gen.encoding.Code, obj)
	b, err := gen.encoding.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("gen: %w", err)
	}

	return b, nil
}

// annotate wraps errors with source code information.
func (gen *Generator) annotate(code Operation, err error) error {
	if err == nil {
		return nil
	} else if src, ok := code.(*SourceInfo); ok {
		err := &SyntaxError{
			File: src.Filename,
			Loc:  gen.pc,
			Pos:  src.Pos,
			Line: src.Line,
			Err:  err,
		}
		return err
	} else {
		return nil
	}
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

// origin unwraps and returns an .ORIG directive.
func origin(oper Operation) (orig *ORIG, ok bool) {
	orig, ok = unwrap(oper).(*ORIG)
	return
}
