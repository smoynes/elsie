// Package asm implements a simple assembler for the machine.
//
// The assembler generates LC-3 machine code from LCASM assembly language, an unnecessary dialect
// that extends the Pratt and Patel's with a few developer-friendly niceties.
//
//	LABEL   AND R3,R3,R2
//	        AND R1,R1,#-1
//	        BRp LABEL
//
//	       .ORIG x3010 ; comment
//	IDENT  .FILL xff00
//		   .END
//
//	LABEL:
//			AND R0, R0, R2
//
// See |Grammar| for a more thorough description of syntax -- semantics are left as an exercise for
// the reader.
//
// # Bugs
//
// There are ambiguities in the grammar and the code could be a whole lot simpler.
package asm

import (
	"fmt"
	"strings"
)

// Grammar declares the syntax of LCASM in EBNF (with some liberties).
var Grammar = (`
program      = { line } ;
line         = ';' comment
             | label ':' [ ';' comment ]
             | label [ ':' ] instruction [ ';' comment ]
             | '.' directive [ ';' comment ]
             | instruction   [ ';' comment ] ;
comment      = { char } ;
directive    = "ORIG" literal
             | "DW" literal
             | "FILL" literal
             | "BLKW" literal
             | "STRINGZ" literal
             | "END" ;
ident        = \p{Letter} { identchar } ;
label        = ident ;
instruction  = opcode [ operands ] ;
opcode       = ident ;
operands     = operand { ',' operand } ;
operand      = immediate
             | register
             | indirect ;
immediate    = '#' integer
             | 'x' hex { hex }
             | 'o' octal { octal }
             | 'b' binary { binary } ;
register     = 'R' octal ;
indirect     = '[' ( identifier | literal | register ) ']' ;
binary       = '0' | '1' | '_' ;
octal        = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '_' ;
decimal      = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | '_' ;
hex          = decimal
             | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' ;
integer      = [ '-' ] decimal { decimal } ;
identchar    = \p{Letter}
             | \p{Decimal Digits}
		     | \p{Marks}
		     | \p{Connector Punctuation}
		     | \p{Dash Punctuation}
		     | \p{Symbols} ;
`)

// SymbolTable maps a symbol reference to its location in object code.
type SymbolTable map[string]uint16

// Count returns the number of symbols in the table.
func (s SymbolTable) Count() int {
	return len(s)
}

// Add adds a symbol to the symbol table.
func (s SymbolTable) Add(sym string, loc uint16) {
	if sym == "" {
		panic("empty symbol")
	}

	sym = strings.ToUpper(sym)
	s[sym] = loc
}

// Offset computes a n-bit PC-relative offset.
func (s SymbolTable) Offset(sym string, pc uint16, n int) (uint16, error) {
	loc, ok := s[sym]
	if !ok {
		return 0xffff, &SymbolError{Symbol: sym, Loc: pc}
	}

	delta := int16(loc - pc)
	bottom := ^(-1 << n)

	if delta >= (1<<n) || delta < -(1<<n) {
		return badSymbol, &OffsetError{uint16(delta)}
	}

	return uint16(delta) & uint16(bottom), nil
}

const badSymbol uint16 = 0xffff

// SyntaxError is a wrapped error returned when the parser encounters a syntax error.
type SyntaxError struct {
	File string // Source file name.
	Loc  uint16 // Location counter.
	Pos  uint16 // Line counter, zero value if now known.
	Line string // Source code line, zero value if not known.
	Err  error  // Error cause.
}

func (pe *SyntaxError) Error() string {
	if pe.Err == nil && pe.Line == "" {
		return fmt.Sprintf("syntax error: loc: %0#4x", pe.Loc)
	} else if pe.Err == nil && pe.Line != "" {
		return fmt.Sprintf("syntax error: line: %q", pe.Line)
	} else {
		return fmt.Sprintf("syntax error: %s: line: %0#4X %q", pe.Err, pe.Pos, pe.Line)
	}
}

// OffsetError is a wrapped error returned when an offset value exceeds its range.
type OffsetError struct {
	Offset uint16
}

func (oe *OffsetError) Error() string {
	return fmt.Sprintf("offset error: %0#4x", oe.Offset)
}

// RegisterError is a wrapped error returned when an instruction names an invalid register.
type RegisterError struct {
	op  string
	Reg string
}

func (re *RegisterError) Error() string {
	return fmt.Sprintf("%s: register error: %s", re.op, re.Reg)
}

// Symbol is a wrapped error returned when a symbol could not be found in the symbol table.
type SymbolError struct {
	Loc    uint16
	Symbol string
}

func (se *SymbolError) Error() string {
	return fmt.Sprintf("symbol error: %q", se.Symbol)
}

func (se *SymbolError) Is(err error) bool {
	if _, ok := err.(*SymbolError); ok {
		return true
	}

	return false
}

// SyntaxTable is holds the parsed code and data indexed by its location counter.
type SyntaxTable []Operation

// Size returns the number of operations in the table.
func (s SyntaxTable) Size() int {
	return len(s)
}

// Add appends an operation to the syntax table.
func (s *SyntaxTable) Add(oper Operation) {
	if oper == nil {
		panic("nil operation")
	}

	*s = append(*s, oper)
}

// Operation is an assembly instruction or directive. It is parsed from source code during the
// assembler's first pass and encoded to object code in the second pass.
type Operation interface {
	// Parse initializes an assembly operation by parsing an opcode and its operands. An error is
	// returned if parsing the operands fails.
	Parse(operator string, operands []string) error

	// Generate encodes an operation as machine code. Using the values from Parse, the operation is
	// converted to one (or more) words.
	Generate(symbols SymbolTable, pc uint16) ([]uint16, error)
}

// SourceInfo wraps an operation to annotate it with parser metadata.
type SourceInfo struct {
	Filename string
	Pos      uint16
	Line     string

	Operation
}

// Unwrap returns the operation which the source info wraps.
func (si *SourceInfo) Unwrap() Operation {
	if si.Operation == nil {
		return nil
	}

	return si.Operation
}
