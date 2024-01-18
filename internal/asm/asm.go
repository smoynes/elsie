/*
Package asm implements a simple assembler for the machine.

The assembler generates LC-3 machine code from LC3ASM assembly language, an unnecessary dialect that
extends the Patt and Patel's with a few developer-friendly niceties.

	LABEL   AND R3,R3,R2
	        AND R1,R1,#-1
	        BRp LABEL

	       .ORIG x3010 ; comment
	IDENT  .FILL xff00
		   .END

	LABEL:
			AND R0, R0, R2

See |Grammar| for a more thorough description of syntax -- semantics are left as an exercise for
the reader.

Typically, one uses the "elsie asm" command to assemble source code:

	go run github.com/smoynes/elsie asm -o program PROGRAM.asm

See github.com/smoynes/internal/cli/cmd.Assembler for details using the command-line interface.

This package also provides APIs to parse source, create syntax and symbol tables, generate machine
code and even extend the language. See Parser and Generator for more.

# Bugs

There are ambiguities in the language grammar and the code could be a whole lot simpler. It is
debatable if ANTLR4 would be better.
*/
package asm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/smoynes/elsie/internal/vm"
)

// Grammar declares the syntax of LC3ASM in EBNF (with some liberties).
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
type SymbolTable map[string]vm.Word

// Count returns the number of symbols in the table.
func (s SymbolTable) Count() int {
	return len(s)
}

// Add adds a symbol to the symbol table.
func (s SymbolTable) Add(sym string, loc vm.Word) {
	if sym == "" {
		panic("empty symbol")
	}

	sym = strings.ToUpper(sym)
	s[sym] = loc
}

// Offset computes a n-bit PC-relative offset.
func (s SymbolTable) Offset(sym string, pc vm.Word, n uint8) (uint16, error) {
	sym = strings.ToUpper(sym)

	loc, ok := s[sym]
	if !ok {
		return badSymbol, &SymbolError{Symbol: sym, Loc: pc}
	}

	delta := int16(loc - pc)
	if delta >= (1<<n) || delta < -(1<<n) {
		return badSymbol, &OffsetRangeError{
			Range:  1 << n,
			Offset: uint16(delta),
		}
	}

	bottom := ^(-1 << n)

	return uint16(delta) & uint16(bottom), nil
}

const badSymbol uint16 = 0xffff

var (
	// ErrOpcode causes a SyntaxError if an opcode is invalid or incorrect.
	ErrOpcode = errors.New("opcode error")

	// ErrOperand causes a SyntaxError if an opcode's operands are invalid or incorrect.
	ErrOperand = errors.New("operand error")

	// ErrLiteral causes a SyntaxError if the literal operand is invalid.
	ErrLiteral = errors.New("literal error")
)

// SyntaxError is a wrapped error returned when the assembler encounters a syntax error. If fields
// are not known, they hold the zero value. For example, the filename is an empty string when the
// source code is not a file.
type SyntaxError struct {
	File string  // Source file name.
	Loc  vm.Word // Location counter.
	Pos  vm.Word // Line counter.
	Line string  // Source code line.
	Err  error   // Error cause.
}

func (se *SyntaxError) Error() string {
	if se.Err == nil && se.Line == "" {
		return fmt.Sprintf("syntax error: loc: %0#4x", se.Loc)
	} else if se.Err == nil && se.Line != "" {
		return fmt.Sprintf("syntax error: line: %q", se.Line)
	} else {
		return fmt.Sprintf("syntax error: %s: line: %0#4x %q", se.Err, se.Pos, se.Line)
	}
}

// Is checks if SyntaxError's error-tree matches a target error.
func (se *SyntaxError) Is(target error) bool {
	if errors.Is(se.Err, target) {
		return true
	} else if err, ok := target.(*SyntaxError); ok && errors.Is(err, err.Err) {
		return true
	} else {
		return se.Pos == err.Pos &&
			se.Line == err.Line &&
			se.Loc == err.Loc &&
			se.File == err.File
	}
}

// OffsetRangeError is a wrapped error returned when an offset value exceeds its range.
type OffsetRangeError struct {
	Offset uint16
	Range  uint16
}

func (oe *OffsetRangeError) Error() string {
	return fmt.Sprintf("offset error: %0#4x", oe.Offset)
}

// LiteralRangeError is a wrapped error returned when an offset value exceeds its range.
type LiteralRangeError struct {
	Literal string
	Range   uint8
}

func (le *LiteralRangeError) Error() string {
	return fmt.Sprintf("literal range error: %q (%0#4x, %0#4x)",
		le.Literal, -uint16(1<<(le.Range)), 1<<(le.Range-1))
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
	Loc    vm.Word
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
	Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error)
}

// SourceInfo wraps an operation to annotate it with parser metadata.
type SourceInfo struct {
	Filename string
	Pos      vm.Word
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

// Condition holds the condition flags for a BR opcode.
type Condition uint8

// Condition codes.
const (
	CondPositive = uint8(vm.ConditionPositive)
	CondZero     = uint8(vm.ConditionZero)
	CondNegative = uint8(vm.ConditionNegative)

	CondZP  = CondZero | CondPositive
	CondNZ  = CondNegative | CondZero
	CondNP  = CondNegative | CondPositive
	CondNZP = CondNegative | CondZero | CondPositive
)
