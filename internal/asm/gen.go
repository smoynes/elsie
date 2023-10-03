package asm

// gen.go implements parsing and code generation for all opcodes and instructions.

import (
	"errors"
	"fmt"
	"strings"
)

// AddressingMode represents how an instruction addresses its operands.
type AddressingMode uint8

//go:generate go run golang.org/x/tools/cmd/stringer -type AddressingMode -output strings_gen.go

const (
	ImmediateMode AddressingMode = iota // IMM
	RegisterMode                        // REG
	IndirectMode                        // IND
)

// An Operation represents a machine-code instruction or an assembler directive. It is parsed from
// source code during the assembler's first pass and encoded to object code in the second pass.
type Operation interface {
	// Parse creates a new instruction by parsing an operator and its operands as represented in
	// source code. An error is returned if parsing the operands fails. The returned instruction may
	// not be semantically or even syntactically correct.
	Parse(operator string, operands []string) (Operation, error)

	// Generate encodes an instruction to machine code. // TODO: should allow 0..n words
	Generate(symbols SymbolTable, pc uint16) (uint16, error)
}

// instructionTable maps assembly-language opcodes to code generators. TODO: this could be a static
// switch statement.
var instructionTable = map[string]Operation{
	"AND": _AND,
	"BR":  _BR, "BRZNP": _BR,
	"BRN": _BR, "BRZ": _BR, "BRP": _BR,
	"BRNZ": _BR, "BRNP": _BR, "BRZP": _BR,
	"LD": _LD,
}

var (
	_BR  Operation = &BR{}
	_AND Operation = &AND{}
	_LD  Operation = &LD{}
)

// BR: Conditional branch.
//
//	BR    [ IDENT | LITERAL ]
//	BRn   [ IDENT | LITERAL ]
//	BRnz  [ IDENT | LITERAL ]
//	BRz   [ IDENT | LITERAL ]
//	BRzp  [ IDENT | LITERAL ]
//	BRp   [ IDENT | LITERAL ]
//	BRnzp [ IDENT | LITERAL ]
//
//	| 0000 | NZP | OFFSET9 |
//	|------+-----+---------|
//	|15  12|11  9|8       0|
type BR struct {
	NZP    uint8
	SYMBOL string
	OFFSET uint16
}

func (br BR) String() string { return fmt.Sprintf("BR(%#v)", br) }

func (BR) Parse(oper string, opers []string) (Operation, error) {
	var nzp uint8

	if len(opers) != 1 {
		return nil, errors.New("br: invalid operands")
	}

	switch strings.ToUpper(oper) {
	case "BR", "BRNZP":
		nzp = 0o7
	case "BRP":
		nzp = 0o1
	case "BRN":
		nzp = 0o2
	case "BRNP":
		nzp = 0o3
	case "BRZ":
		nzp = 0o4
	case "BRZP":
		nzp = 0o5
	case "BRZN":
		nzp = 0o6
	default:
		return &BR{}, fmt.Errorf("unknown opcode: %s", oper)
	}

	off, sym, err := parseImmediate(opers[0], 9)
	if err != nil {
		return nil, fmt.Errorf("br: operand error: %s", err)
	}

	br := &BR{
		NZP:    nzp,
		SYMBOL: sym,
		OFFSET: off,
	}

	return br, nil
}

func (br *BR) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	var (
		code uint16
		loc  uint16
		err  error
	)

	code |= 0o0 << 12
	code |= uint16(br.NZP) << 9

	if br.SYMBOL != "" {
		loc, err = symbolVal(br.SYMBOL, symbols, pc)
		loc = loc - pc
	} else {
		loc = pc + br.OFFSET&0x01ff
	}

	if err != nil {
		return 0xffff, fmt.Errorf("gen: br: operand: %s", err)
	}

	code |= loc

	return code, nil
}

// AND: Bitwise AND binary operator.
//
//	AND DR,SR1,SR2                    ; (register mode)
//
//	| 0101 | DR | SR1 | 0 | 00 | SR2 |
//	|------+----+-----+---+----+-----|
//	|15  12|11 9|8   6| 5 |4  3|2   0|
//
//	AND DR,SR1,#LITERAL               ; (immediate mode)
//	AND DR,SR1,LABEL                  ;
//
//	| 0101 | DR | SR1 | 1 |   IMM5   |
//	|------+----+-----+---+----------|
//	|15  12|11 9|8   6| 5 |4        0|
type AND struct {
	Mode    AddressingMode
	DR, SR1 string

	SR2    string // Set when mode is register.
	OFFSET uint16 // Set when mode is immediate and value is a literal value.
	SYMBOL string // Set when mode is immediate and value is a reference.
}

func (and AND) String() string { return fmt.Sprintf("AND(%#v)", and) }

// Parse parses an AND instruction from its opcode and operands.
func (and AND) Parse(oper string, opers []string) (Operation, error) {
	if len(opers) != 3 {
		return nil, errors.New("and: operands")
	}

	operation := AND{
		DR:  opers[0],
		SR1: opers[1],
	}

	if sr2 := parseRegister(opers[2]); sr2 != "" {
		operation.Mode = RegisterMode
		operation.SR2 = sr2

		return &operation, nil
	}

	operation.Mode = ImmediateMode

	off, sym, err := parseImmediate(opers[2], 5)
	if err != nil {
		return nil, fmt.Errorf("and: operand error: %s", err)
	}

	operation.OFFSET = off
	operation.SYMBOL = sym

	return &operation, nil
}

// Generate returns the machine code for an AND instruction.
func (and AND) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	var code uint16

	dr := registerVal(and.DR)
	sr1 := registerVal(and.SR1)

	code |= 0o6 << 12
	code |= dr << 9
	code |= sr1 << 6

	switch and.Mode {
	case IndirectMode:
		return 0xffff, errors.New("and: addressing error")
	case RegisterMode:
		sr2 := registerVal(and.SR2)
		code |= sr2

		if code != 0xffff {
			return 0xffff, errors.New("and: register error")
		}

		return code, nil
	case ImmediateMode:
		code |= 1 << 5

		switch {
		case and.SYMBOL != "":
			loc, ok := symbols[and.SYMBOL]
			if !ok {
				return 0xffff, fmt.Errorf("and: symbol not found: %q", and.SYMBOL)
			}

			code |= pc - (loc & 0x1f)

			return code, nil
		default:
			code |= and.OFFSET & 0x001f
			return code, nil
		}

	default:
		return 0xffff, errors.New("codegen: address mode error")
	}
}

// LD: Load (immediate) from memory.
//
//	LD DR,LABEL
//	LD DR,#LITERAL
//
//	| 0010 | DR | OFFSET9 |
//	|------+----+---------|
//	|15  12|11 9|8       0|
type LD struct {
	DR     string
	OFFSET uint16
	SYMBOL string
}

func (ld LD) String() string { return fmt.Sprintf("LD(%#v)", ld) }

func (_ LD) Parse(opcode string, operands []string) (Operation, error) {
	var err error

	if strings.ToUpper(opcode) != "LD" {
		return nil, errors.New("ld: opcode error")
	} else if len(operands) != 2 {
		return nil, errors.New("ld: operand error")
	}

	operation := LD{
		DR: operands[0],
	}

	operation.OFFSET, operation.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return nil, fmt.Errorf("ld: operand error: %s", err)
	}

	return &operation, nil
}

func (ld LD) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	dr := registerVal(ld.DR)

	if dr == 0xffff {
		return 0xffff, fmt.Errorf("ld: register error")
	}

	var code uint16 = 0o2 << 12
	code |= dr << 9

	switch {
	case ld.SYMBOL != "":
		loc, ok := symbols[ld.SYMBOL]
		if !ok {
			return 0xffff, fmt.Errorf("ld: symbol not found: %q", ld.SYMBOL)
		}

		code |= pc - (loc & 0x1ff) // TODO ??

		return code, nil
	default:
		code |= ld.OFFSET & 0x001ff
		return code, nil
	}
}

// registerVal returns the registerVal encoded as an integer or 0xffff if the register does not exist.
func registerVal(reg string) uint16 {
	switch reg {
	case "R0":
		return 0
	case "R1":
		return 1
	case "R2":
		return 2
	case "R3":
		return 3
	case "R4":
		return 4
	case "R5":
		return 5
	case "R6":
		return 6
	case "R7":
		return 7
	default:
		return 0xffff
	}
}

func symbolVal(oper string, sym SymbolTable, pc uint16) (uint16, error) {
	// TODO
	if val, ok := sym[oper]; ok {
		return val, nil
	} else {
		return 0xffff, errors.New("codegen: symbolic ref: not implemented")
	}
}
