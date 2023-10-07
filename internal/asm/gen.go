package asm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/log"
)

// gen.go contains a code generation pass for our two-pass assembler.

type Generator struct {
	pc      uint16
	symbols SymbolTable
	syntax  SyntaxTable

	log *log.Logger
}

func NewGenerator(symbols SymbolTable, syntax SyntaxTable) *Generator {
	return &Generator{
		pc:      0x0000,
		symbols: symbols,
		syntax:  syntax,
		log:     log.DefaultLogger(),
	}
}

func (gen *Generator) WriteTo(out io.Writer) (int64, error) {
	var (
		encoded []uint16
		count   int64
		err     error
	)

	if len(gen.syntax) == 0 {
		return 0, nil
	}

	gen.log.Debug("syntax", "syntax", gen.syntax)

	// Write the object-code header: the origin offset. The .ORIG directive should be the first
	// operation in the syntax table.
	if orig := origin(gen.syntax[0]); orig == nil {
		gen.log.Debug("first", "op", orig)
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
			break
		}

		if err = binary.Write(out, binary.LittleEndian, encoded); err != nil {
			break
		}

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
