package asm

// gen.go contains a code generation pass for our two-pass assembler.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/encoding"
	"github.com/smoynes/elsie/internal/vm"
)

// Generator controls the code generation pass of the assembler and translates source code into byte code.
//
// The generator starts at the beginning of the parsed-syntax table, generates code for each
// operation, and then writes the generated code to the output (usually, a file). Use Encode to
// write as hex-encoded ASCII files or use WriteTo write to binary object-code to a buffer.
//
// During the generation pass, any syntax or semantic errors that prevent generating machine code
// are immediately returned. The errors are wrapped in SyntaxErrors and may be tested and retrieved
// using the errors package.
type Generator struct {
	pc       vm.Word
	symbols  SymbolTable
	syntax   SyntaxTable
	encoding encoding.HexEncoding
}

// NewGenerator creates a code generator using the given symbol and syntax tables.
func NewGenerator(symbols SymbolTable, syntax SyntaxTable) *Generator {
	return &Generator{
		pc:       0x0000,
		symbols:  symbols,
		syntax:   syntax,
		encoding: encoding.HexEncoding{},
	}
}

// Encode generates object code and encodes it as hex-encoded ASCII object code.
//
// Multiple sections are supported if the syntax table has multiple ORIG directives.
func (gen *Generator) Encode() ([]byte, error) {
	if len(gen.syntax) == 0 {
		return nil, nil
	}

	var (
		obj   vm.ObjectCode
		count int64
		err   error
	)

	// We expect the .ORIG directive to be the first operation in the syntax table. TODO: We should
	// be able to support multiple origins if the encoder does.
	if _, ok := origin(gen.syntax[0]); !ok {
		return nil, fmt.Errorf(".ORIG should be first operation; was: %T", gen.syntax[0])
	}

	for _, op := range gen.syntax {
		if op == nil {
			continue
		} else if orig, ok := origin(op); ok {
			if obj.Code != nil {
				gen.encoding.Code = append(gen.encoding.Code, obj)
			}

			gen.pc = orig.LITERAL
			obj = vm.ObjectCode{Orig: gen.pc}

			continue // We don't need to generate code.
		}

		genWords, genErr := op.Generate(gen.symbols, gen.pc+1)

		if genErr != nil {
			err = gen.annotate(op, genErr)
			break
		}

		obj.Code = append(obj.Code, genWords...)

		gen.pc += vm.Word(len(genWords))
		count += int64(len(genWords) * 2)
	}

	if err != nil {
		return nil, fmt.Errorf("gen: %w", err)
	}

	gen.encoding.Code = append(gen.encoding.Code, obj)

	if b, err := gen.encoding.MarshalText(); err != nil {
		return nil, fmt.Errorf("gen: %w", err)
	} else {
		return b, nil
	}
}

// WriteTo writes generated binary machine-code to an output stream. It implements io.WriteTo.
//
// Unlinke Encode, WriteTo does not support writing more than a single section of code.
func (gen *Generator) WriteTo(out io.Writer) (int64, error) {
	if len(gen.syntax) == 0 {
		return 0, nil
	}

	var (
		count int64
		err   error
	)

	// Write the origin offset as the leader of the object file. The .ORIG directive should be the
	// first operation in the syntax table.
	if orig, ok := origin(gen.syntax[0]); ok {
		gen.pc = orig.LITERAL
	} else {
		return 0, fmt.Errorf(".ORIG should be first operation; was: %T", gen.syntax[0])
	}

	for i, oper := range gen.syntax {
		if oper == nil {
			continue
		} else if _, ok := origin(oper); ok && i != 0 {
			err = errors.New(".ORIG directive may only be the first operation")
			break
		}

		generated, genErr := oper.Generate(gen.symbols, gen.pc) // TODO: should this be pc + 1

		if err != nil {
			err = gen.annotate(oper, genErr)
			break
		}

		if err = binary.Write(out, binary.BigEndian, generated); err != nil {
			break
		}

		gen.pc += vm.Word(len(generated))
		count += int64(len(generated) * 2)
	}

	if err != nil {
		return count, fmt.Errorf("gen: %w", err)
	}

	return count, nil
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
