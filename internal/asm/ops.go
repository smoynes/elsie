package asm

// ops.go implements parsing and code generation for all opcodes and instructions.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
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

// Operation is an assembly instruction or directive. It is parsed from source code during the
// assembler's first pass and encoded to object code in the second pass.
type Operation interface {
	// Parse initializes an assembly operation by parsing an opcode and its operands. An error is
	// returned if parsing the operands fails.
	Parse(operator string, operands []string) error

	// Generate encodes an operation as machine code. Using the values from Parse, the operation is
	// converted to one (or more) words.
	//// TODO: should allow (or more) words.
	Generate(symbols SymbolTable, pc uint16) (uint16, error)
}

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

func (br *BR) Parse(oper string, opers []string) error {
	var nzp uint8

	if len(opers) != 1 {
		return errors.New("br: invalid operands")
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
		return fmt.Errorf("unknown opcode: %s", oper)
	}

	off, sym, err := parseImmediate(opers[0], 9)
	if err != nil {
		return fmt.Errorf("br: operand error: %w", err)
	}

	*br = BR{
		NZP:    nzp,
		SYMBOL: sym,
		OFFSET: off,
	}

	return nil
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
		return 0xffff, fmt.Errorf("gen: br: operand: %w", err)
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
//	| 0101 | DR | SR1 | 1 | IMM5 |
//	|------+----+-----+---+------|
//	|15  12|11 9|8   6| 5 |4    0|
type AND struct {
	Mode    AddressingMode
	DR, SR1 string

	SR2    string // Set when mode is register.
	OFFSET uint16 // Set when mode is immediate and value is a literal value.
	SYMBOL string // Set when mode is immediate and value is a reference.
}

func (and AND) String() string { return fmt.Sprintf("AND(%#v)", and) }

// Parse parses an AND instruction from its opcode and operands.
func (and *AND) Parse(oper string, opers []string) error {
	if len(opers) != 3 {
		return errors.New("and: operands")
	}

	*and = AND{
		DR:  parseRegister(opers[0]),
		SR1: parseRegister(opers[1]),
	}

	if sr2 := parseRegister(opers[2]); sr2 != "" {
		and.Mode = RegisterMode
		and.SR2 = sr2

		return nil
	}

	and.Mode = ImmediateMode

	off, sym, err := parseImmediate(opers[2], 5)
	if err != nil {
		return fmt.Errorf("and: operand error: %w", err)
	}

	and.OFFSET = off
	and.SYMBOL = sym

	return nil
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
				return 0xffff, fmt.Errorf("and: symbol error: %q", and.SYMBOL)
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

// LD: Load from memory, PC-relative..
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

func (ld *LD) Parse(opcode string, operands []string) error {
	var err error

	if strings.ToUpper(opcode) != "LD" {
		return errors.New("ld: opcode error")
	} else if len(operands) != 2 {
		return errors.New("ld: operand error")
	}

	*ld = LD{
		DR: operands[0],
	}

	ld.OFFSET, ld.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return fmt.Errorf("ld: operand error: %w", err)
	}

	return nil
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
			return 0xffff, fmt.Errorf("ld: symbol error: %q", ld.SYMBOL)
		}

		code |= pc - (loc & 0x1ff) // TODO ??

		return code, nil
	default:
		code |= ld.OFFSET & 0x001ff
		return code, nil
	}
}

// LDR: Load from memory, register-relative.
//
//	LDR DR,SR,#LITERAL
//	LDR DR,SR,LABEL
//
//	| 0110 | DR | SR | OFFSET6 |
//	|------+----+----+---------|
//	|15  12|11 9|8  6|5       0|
type LDR struct {
	DR     string
	SR     string
	OFFSET uint16
	SYMBOL string
}

func (ldr LDR) String() string { return fmt.Sprintf("LDR(%#v)", ldr) }

func (ldr *LDR) Parse(opcode string, operands []string) error {
	if opcode != "LDR" {
		return errors.New("ldr: opcode error")
	}

	*ldr = LDR{
		DR: operands[0],
		SR: operands[1],
	}

	off, sym, err := parseImmediate(operands[2], 6)
	if err != nil {
		return fmt.Errorf("ldr: operand error: %w", err)
	}

	ldr.OFFSET = off
	ldr.SYMBOL = sym

	return nil
}

func (ldr LDR) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	dr := registerVal(ldr.DR)

	if dr == BadRegister {
		return 0xffff, fmt.Errorf("ldr: register error")
	}

	code := 0b0010<<12 | dr<<9

	switch {
	case ldr.SYMBOL != "":
		loc, ok := symbols[ldr.SYMBOL]
		if !ok {
			return 0xffff, fmt.Errorf("ldr: symbol error: %q", ldr.SYMBOL)
		}

		code |= (pc - loc) & 0x001f

		return code, nil
	default:
		code |= ldr.OFFSET & 0x001f
		return code, nil
	}
}

// ADD: Arithmetic addition operator.
//
//	ADD DR,SR1,SR2
//	ADD DR,SR1,#LITERAL
//
//	| 0001 | DR | SR1 | 0 | 00 | SR2 | (register mode)
//	|------+----+-----+---+----+-----|
//	|15  12|11 9|8   6| 5 |4  3|2   0|
//
//	| 0001 | DR | SR1 | 1 |   IMM5   | (immediate mode)
//	|------+----+-----+---+----------|
//	|15  12|11 9|8  6 | 5 |4        0|
//
// .
type ADD struct {
	DR      string
	SR1     string
	SR2     string // Not empty when register mode.
	SYMBOL  string // Not empty with symbol, immediate mode.
	LITERAL uint16 // Literal value otherwise, immediate mode.
}

func (add ADD) String() string { return fmt.Sprintf("%#v", add) }

func (add *ADD) Parse(opcode string, operands []string) error {
	if opcode != "ADD" {
		return errors.New("add: opcode error")
	} else if len(operands) != 3 {
		return errors.New("add: operand error")
	}

	*add = ADD{
		DR:  parseRegister(operands[0]),
		SR1: parseRegister(operands[1]),
	}

	if sr2 := parseRegister(operands[2]); sr2 != "" {
		add.SR2 = sr2
	} else {
		off, sym, err := parseImmediate(operands[2], 5)
		if err != nil {
			return fmt.Errorf("add: operand error: %w", err)
		}

		add.LITERAL = off & 0x1f
		add.SYMBOL = sym
	}

	return nil
}

func (add ADD) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	var code uint16 = 0b0001 << 12

	sr1 := registerVal(add.SR1)
	if sr1 == BadRegister {
		return 0, errors.New("add: register error")
	}

	code |= sr1 << 6

	if add.SR2 != "" {
		sr2 := registerVal(add.SR2)
		code |= sr2

		return code, nil
	}

	code |= 1 << 5
	code |= add.LITERAL & 0x001f

	return code, nil
}

// NOT: Bitwise complement.
//
//	NOT DR,SR ;; DR <- ^(SR)
//
//	| 1001 | DR | SR | 1 1111 |
//	|------+----+----+--------|
//	|15  12|11 9|8  6| 5     0|
//
// .
type NOT struct {
	DR string
	SR string
}

func (not NOT) String() string { return fmt.Sprintf("%#v", not) }

func (not *NOT) Parse(opcode string, operands []string) error {
	if opcode != "NOT" {
		return errors.New("not: opcode error")
	} else if len(operands) != 2 {
		return errors.New("not: operand error")
	}

	*not = NOT{
		DR: parseRegister(operands[0]),
		SR: parseRegister(operands[1]),
	}

	return nil
}

func (not *NOT) Generate(symbols SymbolTable, pc uint16) (uint16, error) {
	var code uint16 = 0b1001 << 12

	if not.DR == "" || not.SR == "" {
		return 0xffff, fmt.Errorf("gen: not: bad operand")
	}

	code |= registerVal(not.DR) << 9
	code |= registerVal(not.SR) << 6
	code |= 0x001f

	return code, nil
}

// .FILL: Allocate and initialize one word of data.
//
//	.FILL x1234
//	.FILL 0
type FILL struct {
	LITERAL uint16 // Literal constant.
}

func (fill *FILL) Parse(opcode string, operands []string) error {
	val, err := parseLiteralConstant(operands[0])
	fill.LITERAL = val
	return err
}

func (fill *FILL) Generate(_ SymbolTable, pc uint16) (uint16, error) {
	return fill.LITERAL, nil
}

// .BLKW: Data allocation directive.
//
//	.BLKW 1
type BLKW struct {
	ALLOC uint16 // Number of words allocated.
}

func (blkw *BLKW) Parse(opcode string, operands []string) error {
	val, err := parseLiteralConstant(operands[0])
	blkw.ALLOC = val
	return err
}

func (blkw *BLKW) Generate(_ SymbolTable, pc uint16) (uint16, error) {
	return 0x2361, nil // TODO: un-init memory
}

// .ORIG: Origin directive. Sets the location counter to the value.
//
//	.ORIG x1234
//	.ORIG 0
type ORIG struct {
	LITERAL uint16 // Literal constant.
}

func (orig *ORIG) Parse(opcode string, operands []string) error {
	if len(operands) != 1 {
		return errors.New("argument error")
	}

	arg := operands[0]

	switch arg[0] {
	case 'x', 'b', 'o':
		arg = "0" + arg
	}

	val, err := strconv.ParseUint(arg, 0, 16)
	numError := &strconv.NumError{}

	if errors.As(err, &numError) {
		return fmt.Errorf("parse error: %s (%s)", numError.Num, numError.Err.Error())
	} else if val > math.MaxUint16 {
		return errors.New("argument error")
	}

	orig.LITERAL = uint16(val)

	return nil
}

func (orig *ORIG) Generate(_ SymbolTable, pc uint16) (uint16, error) {
	return 0x0000, nil
}

// registerVal returns the registerVal encoded as an integer or BadRegister if the register does not
// exist.
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
		return BadRegister
	}
}

// BadRegister is returned when a named register is invalid.
const BadRegister uint16 = 0xffff

func symbolVal(oper string, sym SymbolTable, _ uint16) (uint16, error) {
	// TODO
	if val, ok := sym[oper]; ok {
		return val, nil
	} else {
		return 0xffff, errors.New("codegen: symbolic ref: not implemented")
	}
}
