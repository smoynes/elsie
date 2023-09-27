// Package asm contains a simple two-pass assembler for the machine.
package asm

import (
	"bufio"
	"errors"
	"io"

	"github.com/smoynes/elsie/internal/log"
)

// A Parser reads source code and produces an intermediate representation (IR). The user provides
// one or more input streams and then asks the parser for the products: a symbol table, directives
// and similar. To accommodate multiple input files, parsing and I/O errors are reported separately
// from reads. Parser.Err returns the first fatal error encountered, similar to |bufio.Scanner|.
type Parser interface {
	// Read provides an input stream of source code to the parser. The parser takes ownership and is
	// responsible for reading, closing and handling I/O errors.
	Read(in io.ReadCloser) bool

	// Symbols returns the symbol table.
	Symbols() SymbolTable

	// Err reports all errors found during parsing.
	Err() error
}

// SymbolTable maps symbol literal to its location.
type SymbolTable map[string]int

type ParseError struct {
	errs []error
}

func (e ParseError) Error() string {
	var buf []byte // 'orrible, jus' 'orrible!
	for _, e := range e.errs {
		buf = append(buf, e.Error()...)
		buf = append(buf, '\n')
	}

	return string(buf)
}

func (e ParseError) Unwrap() []error {
	return e.errs
}

type parser struct {
	symbols SymbolTable
	errs    []error
	log     *log.Logger
}

func NewParser(log *log.Logger) Parser {
	return &parser{
		log:     log,
		symbols: make(SymbolTable),
	}
}

var _ Parser = (*parser)(nil)

func (p *parser) Read(in io.ReadCloser) bool {
	defer func() { _ = in.Close() }()

	if len(p.errs) > 0 {
		return false
	}

	// Create scanner that produces lines.
	lines := bufio.NewScanner(in)

tokenize:
	// Our strategy is straightforward: scan each line in the source, split the line into tokens,
	// and build the symbol table.
	for {
		scanned := lines.Scan()
		tok := lines.Text()

		p.log.Debug("parsed token", "token", tok, "scanned", scanned)

		switch {
		case !scanned:
			break tokenize

		case tok == "":
			// Should never (?) get empty tokens.
			p.errs = append(p.errs, errors.New("empty token"))
			break tokenize
		}
	}

	return false
}

func (p *parser) Symbols() SymbolTable {
	return p.symbols
}

func (p *parser) Err() error {
	if len(p.errs) == 0 {
		return nil
	} else if len(p.errs) == 1 {
		return p.errs[0]
	} else {
		return &ParseError{
			errs: p.errs,
		}
	}
}
