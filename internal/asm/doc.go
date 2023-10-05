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
