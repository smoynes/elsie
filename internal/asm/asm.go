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
program        = { line } ;

line           = ';' comment
               | label ':' [ ';' comment ]
               | label [ ':' ] instruction [ ';' comment ]
               | '.' directive [ ';' comment ]
               | instruction   [ ';' comment ] ;

comment        = { char } ;

directive      = "ORIG" literal
               | "DW" literal
               | "FILL" literal
               | "BLKW" literal
               | "STRINGZ" literal
               | "END" ;

ident          = \p{Letter} { identchar } ;

label          = ident ;

instruction    = opcode [ operands ] ;

opcode         = ident ;

operands       = operand { ',' operand } ;

operand        = immediate
               | register
               | indirect ;

immediate      = '#' integer
               | 'x' hex { hex }
               | 'o' octal { octal }
               | 'b' binary { binary } ;

register       = 'R' octal ;

indirect       = '[' ( identifier | literal | register ) ']' ;

binary         = '0' | '1' | '_' ;

octal          = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '_' ;

decimal        = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | '_' ;

hex            = decimal
               | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' ;

integer        = [ '-' ] decimal { decimal } ;

identchar      = \p{Letter}
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
		return 0xffff, fmt.Errorf("%s: symbol not found", sym)
	}
	delta := loc - pc - 1
	bottom := ^(-1 << n)

	if delta >= (1 << n) {
		return delta, fmt.Errorf("%s: offset error", sym)
	}

	return delta & uint16(bottom), nil
}

// SyntaxError is a wrapped error returned when the parser encounters a syntax error.
type SyntaxError struct {
	Loc, Pos uint16
	Line     string
	Err      error
}

func (pe *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s: line: %d %q", pe.Err, pe.Pos, pe.Line)
}

// SyntaxTable is holds the parsed code and data indexed by its location counter.
type SyntaxTable []Operation

// Size returns the number of operations in the table.
func (s SyntaxTable) Size() int {
	n := 0
	for _, oper := range s {
		if oper != nil {
			n++
		}
	}
	return n
}

// Add puts an operation in a location in the table.
func (s SyntaxTable) Add(loc uint16, oper Operation) {
	if oper == nil {
		panic("nil operation")
	}

	s[loc] = oper
}

// Syntax returns the abstract syntax table (or, AST).
func (p *Parser) Syntax() SyntaxTable {
	return p.syntax
}
