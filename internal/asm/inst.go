// inst.go implements parsing and code generation for each instruction opcode.

package asm

import (
	"errors"
	"fmt"
	"strings"
)

// Operators maps an opcode to a type which implements Instruction for the operator.
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

// AND is the bitwise AND binary operator with register and immediate modes.
//
//	AND REG, REG, REG
//	AND REG, REG, LITERAL
type AND struct {
	Mode AddressingMode

	DR, SR1 string
	SR2     string // Set when Mode is Register.
	LIT     string // Set when Mode is Immediate.
}

// AddressingMode represents how an address addresses its operands.
type AddressingMode uint8

//go:generate go run golang.org/x/tools/cmd/stringer -type AddressingMode -output strings_gen.go

// Addressing modes.
const (
	ImmediateMode AddressingMode = iota // IMM
	RegisterMode                        // REG
	IndirectMode                        // IND
)

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

func immediateOperand(oper string) string {
	if len(oper) > 1 && oper[0] == '#' {
		return oper[1:]
	} else {
		return oper
	}
}
