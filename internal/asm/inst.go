package asm

// gen.go implements code generation for each instruction opcode.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/smoynes/elsie/internal/vm"
)

// AddressingMode represents how an instruction addresses its operands.
type AddressingMode uint8

//go:generate go run golang.org/x/tools/cmd/stringer -type AddressingMode -output strings_gen.go

const (
	ImmediateMode AddressingMode = iota // IMM
	RegisterMode                        // REG
	IndirectMode                        // IND
)

// Operators maps an opcode to a type which implements Instruction.
var operators = map[string]Instruction{
	"AND": _AND,
	"BR":  _BR,
	"BRN": _BR, "BRZ": _BR, "BRP": _BR,
	"BRNZ": _BR, "BRNP": _BR, "BRZP": _BR,
}

var (
	_AND Instruction = &AND{}
	_BR  Instruction = &AND{}
)

// AND: Bitwise AND binary operator
//
//	AND DR,SR1,SR2                 ; (register mode)
//
//	| 0101 | DR | SR1 | 0 | 00 | SR2 |
//	|------+----+-----+---+----+-----|
//	|15  12|11 9|8   6| 5 |4  3|2   0|
//
//	AND DR,SR1,[ LITERAL | IDENT ] ; (immediate mode)
//
//	| 0101 | DR | SR1 | 1 |   IMM5   |
//	|------+----+-----+---+----------|
//	|15  12|11 9|8   6| 5 |4        0|
type AND struct {
	Mode AddressingMode

	DR, SR1 string
	SR2     string // Set when Mode is Register.
	LIT     string // Set when Mode is Immediate.
}

type Opcode = vm.Opcode // TODO: move opcode

func (and AND) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	var code uint16

	dr := register(and.DR)
	sr1 := register(and.SR1)

	code |= 0o6 << 12
	code |= dr << 9
	code |= sr1 << 6

	switch and.Mode {
	case RegisterMode:
		sr2 := register(and.SR2)
		code |= sr2

		if code != 0xffff {
			return 0xffff, errors.New("codegen: no register")
		}

		return code, nil
	case ImmediateMode:
		imm5, litErr := literalVal(and.LIT, 5)
		loc, symErr := symbolVal(and.LIT, symbols, pc)

		switch {
		case litErr == nil:
			code |= 1 << 5
			code |= imm5 & 0x001f

			return code, nil

		case litErr != nil && symErr == nil:
			code = loc
			code = pc

			return code, nil
		default:
			return 0xffff, fmt.Errorf("codegen: immediate mode operand: %s %s", litErr, symErr)
		}

	default:
		return 0xffff, errors.New("codegen: address mode error")
	}
}

func (ins AND) String() string { return "AND" }

func (ins AND) Parse(oper string, opers []string) (Instruction, error) {
	if len(opers) != 3 {
		return nil, errors.New("and: operands")
	}

	and := AND{
		DR:  opers[0],
		SR1: opers[1],
	}
	if sr2 := registerOperand(opers[2]); sr2 != "" {
		and.Mode = RegisterMode
		and.SR2 = sr2
	} else if lit := immediateOperand(opers[2]); lit != "" {
		and.Mode = ImmediateMode
		and.LIT = lit
	} else {
		return nil, errors.New("and: invalid mode")
	}

	return &and, nil
}

// BR is the branch instruction and implements several operator variations. All operators use immediate mode addressing.
//
//	BRz   [ IDENT | LITERAL ]
//	BRzn  [ IDENT | LITERAL ]
//	BRznp [ IDENT | LITERAL ]
//	BRn   [ IDENT | LITERAL ]
//	BRnp  [ IDENT | LITERAL ]
//	BRp   [ IDENT | LITERAL ]
type BR struct {
	NZP uint8
	LIT string
}

func (br BR) String() string { return "BR" }

func (BR) Parse(oper string, opers []string) (Instruction, error) {
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

	br := &BR{
		NZP: nzp,
		LIT: immediateOperand(opers[0]),
	}

	return br, nil
}

func register(reg string) uint16 {
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

func registerOperand(oper string) string {
	switch oper {
	case
		"R0", "R1", "R2", "R3",
		"R4", "R5", "R6", "R7":
		return oper
	default:
		return ""
	}
}

// TODO: Refactor to extract
type Word = vm.Word

func literalVal(oper string, n uint8) (uint16, error) {
	if len(oper) < 2 {
		return 0xffff, fmt.Errorf("codegen: literal error: %s", oper[2:])
	}

	switch pref, lit := oper[:2], oper[2:]; pref {
	case "#x":
		i, err := strconv.ParseUint(lit, 16, 16)
		if err != nil {
			return 0xffff, fmt.Errorf("codegen: literal error: %s (%s)", err, lit)
		}

		val := Word(i)
		val.Sext(n)

		return uint16(val), nil
	default:
		return 0xffff, fmt.Errorf("codegen: literal error: %s", lit)
	}

}

func symbolVal(oper string, sym SymbolTable, pc uint16) (uint16, error) {
	return 0xffff, errors.New("codegen: symbolic ref: not implemented")
}

func immediateOperand(oper string) string {
	if len(oper) > 1 && oper[0] == '#' {
		return oper[1:]
	} else {
		return oper
	}
}
