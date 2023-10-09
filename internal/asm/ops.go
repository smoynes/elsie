package asm

// ops.go implements parsing and code generation for all opcodes and instructions.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/smoynes/elsie/internal/vm"
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
	SourceInfo
	NZP    uint8
	SYMBOL string
	OFFSET uint16
}

func (br BR) String() string { return fmt.Sprintf("BR(%#v)", br) }

// Parse parses all variations of the BR* instruction based on the opcode.
func (br *BR) Parse(opcode string, opers []string) error {
	var nzp uint16

	if len(opers) != 1 {
		return errors.New("br: invalid operands")
	}

	switch strings.ToUpper(opcode) {
	case "BR", "BRNZP":
		nzp = uint16(vm.ConditionNegative | vm.ConditionZero | vm.ConditionPositive)
	case "BRP":
		nzp = uint16(vm.ConditionPositive)
	case "BRZ":
		nzp = uint16(vm.ConditionZero)
	case "BRZP":
		nzp = uint16(vm.ConditionZero | vm.ConditionPositive)
	case "BRN":
		nzp = uint16(vm.ConditionNegative)
	case "BRNP":
		nzp = uint16(vm.ConditionNegative | vm.ConditionPositive)
	case "BRNZ":
		nzp = uint16(vm.ConditionNegative | vm.ConditionZero)
	default:
		return fmt.Errorf("unknown opcode: %s", opcode)
	}

	off, sym, err := parseImmediate(opers[0], 9)
	if err != nil {
		return fmt.Errorf("br: operand error: %w", err)
	}

	*br = BR{
		SourceInfo: br.SourceInfo,
		NZP:        uint8(nzp),
		SYMBOL:     sym,
		OFFSET:     off,
	}

	return nil
}

func (br *BR) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	code := vm.NewInstruction(vm.BR, uint16(br.NZP)<<9)

	if br.SYMBOL != "" {
		offset, err := symbols.Offset(br.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("and: %w", err)
		}

		code.Operand(offset)
	} else {
		code.Operand(br.OFFSET & 0x01ff)
	}

	return []uint16{code.Encode()}, nil
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
	SourceInfo
	DR     string
	SR1    string
	SR2    string // Register mode.
	SYMBOL string // Symbolic reference.
	OFFSET uint16 // Otherwise.
}

func (and AND) String() string { return fmt.Sprintf("AND(%#v)", and) }

// Parse parses an AND instruction from its opcode and operands.
func (and *AND) Parse(oper string, opers []string) error {
	if len(opers) != 3 {
		return errors.New("and: operands")
	}

	*and = AND{
		SourceInfo: and.SourceInfo,
		DR:         parseRegister(opers[0]),
		SR1:        parseRegister(opers[1]),
	}

	if sr2 := parseRegister(opers[2]); sr2 != "" {
		and.SR2 = sr2

		return nil
	}

	off, sym, err := parseImmediate(opers[2], 5)
	if err != nil {
		return fmt.Errorf("and: operand error: %w", err)
	}

	and.OFFSET = off
	and.SYMBOL = sym

	return nil
}

// Generate returns the machine code for an AND instruction.
func (and *AND) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	dr := registerVal(and.DR)
	sr1 := registerVal(and.SR1)

	if dr == badGPR {
		return nil, &RegisterError{"and", and.DR}
	} else if sr1 == badGPR {
		return nil, &RegisterError{"and", and.SR1}
	}

	code := vm.NewInstruction(vm.AND, dr<<9|sr1<<6)

	switch {
	case and.SR2 != "":
		sr2 := registerVal(and.SR2)
		if sr2 == badGPR {
			return nil, &RegisterError{"and", and.SR2}
		}

		code.Operand(sr2)
	case and.SYMBOL != "":
		code.Operand(1 << 5)

		offset, err := symbols.Offset(and.SYMBOL, pc, 5)
		if err != nil {
			return nil, fmt.Errorf("and: %w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(1 << 5)
		code.Operand(and.OFFSET & 0x001f)
	}

	return []uint16{code.Encode()}, nil
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
	SourceInfo
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
		SourceInfo: ld.SourceInfo,
		DR:         operands[0],
	}

	ld.OFFSET, ld.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return fmt.Errorf("ld: operand error: %w", err)
	}

	return nil
}

func (ld LD) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	dr := registerVal(ld.DR)
	if dr == badGPR {
		return nil, &RegisterError{op: "ld", Reg: ld.DR}
	}

	code := vm.NewInstruction(vm.LD, dr<<9)

	switch {
	case ld.SYMBOL != "":
		offset, err := symbols.Offset(ld.SYMBOL, pc, 8)
		if err != nil {
			return nil, fmt.Errorf("and: %w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(ld.OFFSET & 0x0ff)
	}

	return []uint16{code.Encode()}, nil
}

// LDR: Load from memory, register-relative.
//
//	LDR DR,SR,LABEL
//	LDR DR,SR,#LITERAL
//
//	| 0110 | DR | SR | OFFSET6 |
//	|------+----+----+---------|
//	|15  12|11 9|8  6|5       0|
//
// .
type LDR struct {
	SourceInfo
	DR     string
	SR     string
	OFFSET uint16
	SYMBOL string
}

func (ldr LDR) String() string { return fmt.Sprintf("LDR(%#v)", ldr) }

func (ldr *LDR) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "LDR" {
		return errors.New("ldr: opcode error")
	} else if len(operands) != 3 {
		return errors.New("ldr: operand error")
	}

	*ldr = LDR{
		SourceInfo: ldr.SourceInfo,
		DR:         operands[0],
		SR:         operands[1],
	}

	ldr.OFFSET, ldr.SYMBOL, err = parseImmediate(operands[2], 6)
	if err != nil {
		return fmt.Errorf("ldr: operand error: %w", err)
	}

	return nil
}

func (ldr LDR) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	dr := registerVal(ldr.DR)
	sr := registerVal(ldr.SR)

	if dr == badGPR {
		return nil, &RegisterError{"ldr", ldr.DR}
	} else if sr == badGPR {
		return nil, &RegisterError{"ldr", ldr.SR}
	}

	code := vm.NewInstruction(vm.LDR, dr<<9|sr<<6)

	switch {
	case ldr.SYMBOL != "":
		offset, err := symbols.Offset(ldr.SYMBOL, pc, 6)
		if err != nil {
			return nil, fmt.Errorf("ldr: %w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(ldr.OFFSET & 0x003f)
	}

	return []uint16{code.Encode()}, nil
}

// LEA: Load effective address.
//
//	LDR DR,LABEL
//	LDR DR,#LITERAL
//
//	| 1110 | DR | OFFSET9 |
//	|------+----+---------|
//	|15  12|11 9|8       0|
//
// .
type LEA struct {
	SourceInfo
	DR     string
	SYMBOL string
	OFFSET uint16
}

func (lea LEA) String() string { return fmt.Sprintf("%#v", lea) }

func (lea *LEA) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "LEA" {
		return errors.New("lea: opcode error")
	} else if len(operands) != 2 {
		return errors.New("lea: operand error")
	}

	*lea = LEA{
		SourceInfo: lea.SourceInfo,
		DR:         operands[0],
	}

	lea.OFFSET, lea.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return fmt.Errorf("lea: operand error: %w", err)
	}

	return nil
}

func (lea LEA) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	dr := registerVal(lea.DR)

	if dr == badGPR {
		return nil, &RegisterError{"lea", lea.DR}
	}

	code := vm.NewInstruction(vm.LEA, dr<<9)

	switch {
	case lea.SYMBOL != "":
		offset, err := symbols.Offset(lea.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("lea: %w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(lea.OFFSET & 0x01ff)
	}

	return []uint16{code.Encode()}, nil
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
	SourceInfo
	DR      string
	SR1     string
	SR2     string // Not empty when register mode.
	LITERAL uint16 // Literal value otherwise, immediate mode.
}

func (add ADD) String() string { return fmt.Sprintf("%#v", add) }

func (add *ADD) Parse(opcode string, operands []string) error {
	if opcode != "ADD" {
		return errors.New("add: opcode error")
	} else if len(operands) != 3 {
		return errors.New("add: operand error")
	}

	dr := parseRegister(operands[0])
	sr1 := parseRegister(operands[1])

	*add = ADD{
		SourceInfo: add.SourceInfo,
		DR:         dr,
		SR1:        sr1,
	}

	if sr2 := parseRegister(operands[2]); sr2 != "" {
		add.SR2 = sr2
	} else {
		off, _, err := parseImmediate(operands[2], 5)
		if err != nil {
			return fmt.Errorf("add: operand error: %w", err)
		}

		add.LITERAL = off & 0x1f
	}

	return nil
}

func (add ADD) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	dr := registerVal(add.DR)
	sr1 := registerVal(add.SR1)

	if dr == badGPR {
		return nil, &RegisterError{"and", add.DR}
	} else if sr1 == badGPR {
		return nil, &RegisterError{"and", add.SR1}
	}

	code := vm.NewInstruction(vm.ADD, dr<<9|sr1<<6)

	if add.SR2 != "" {
		sr2 := registerVal(add.SR2)
		if sr2 == badGPR {
			return nil, &RegisterError{"and", add.SR2}
		}

		code.Operand(sr2)
	} else {
		code.Operand(1 << 5)
		code.Operand(add.LITERAL & 0x001f)
	}

	return []uint16{code.Encode()}, nil
}

// TRAP: System call or software interrupt.
//
//	TRAP x25
//
//	| 1111 | 0000 | VECTOR8 |
//	|------+------+---------|
//	|15  12|11   8|7       0|
//
// .
type TRAP struct {
	SourceInfo
	LITERAL uint16
}

func (trap TRAP) String() string { return fmt.Sprintf("%#v", trap) }

func (trap *TRAP) Parse(opcode string, operands []string) error {
	if opcode != "TRAP" {
		return errors.New("trap: operator error")
	} else if len(operands) != 1 {
		return errors.New("trap: operand error")
	}

	lit, err := parseLiteral(operands[0], 8)
	if err != nil {
		return fmt.Errorf("trap: operand error: %w", err)
	}

	*trap = TRAP{
		SourceInfo: trap.SourceInfo,
		LITERAL:    lit,
	}

	return nil
}

func (trap TRAP) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	code := uint16(vm.TRAP)<<12 | trap.LITERAL&0x00ff
	return []uint16{code}, nil
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
	SourceInfo
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

	dr := parseRegister(operands[0])
	sr := parseRegister(operands[1])

	*not = NOT{
		SourceInfo: not.SourceInfo,
		DR:         dr,
		SR:         sr,
	}

	return nil
}

func (not *NOT) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	if not.DR == "" || not.SR == "" {
		return nil, fmt.Errorf("gen: not: bad operand")
	}

	dr := registerVal(not.DR)
	sr := registerVal(not.SR)

	if dr == badGPR {
		return nil, &RegisterError{"not", not.DR}
	} else if sr == badGPR {
		return nil, &RegisterError{"not", not.SR}
	}

	code := vm.NewInstruction(vm.NOT, dr<<9|sr<<6|0x003f)

	return []uint16{code.Encode()}, nil
}

// .FILL: Allocate and initialize one word of data.
//
//	.FILL x1234
//	.FILL 0
type FILL struct {
	SourceInfo
	LITERAL uint16 // Literal constant.
}

func (fill *FILL) Parse(opcode string, operands []string) error {
	val, err := parseLiteral(operands[0], 16)
	fill.LITERAL = val

	return err
}

func (fill *FILL) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	return []uint16{fill.LITERAL}, nil
}

// .BLKW: Data allocation directive.
//
//	.BLKW 1
type BLKW struct {
	SourceInfo
	ALLOC uint16 // Number of words allocated.
}

func (blkw *BLKW) Parse(opcode string, operands []string) error {
	val, err := parseLiteral(operands[0], 16)
	blkw.ALLOC = val

	return err
}

func (blkw *BLKW) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	return []uint16{0x2361}, nil // TODO: un-init memory ?
}

// .ORIG: Origin directive. Sets the location counter to the value.
//
//	.ORIG x1234
//	.ORIG 0
type ORIG struct {
	SourceInfo
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

	if numError := (&strconv.NumError{}); errors.As(err, &numError) {
		return fmt.Errorf("parse error: %s (%s)", numError.Num, numError.Err.Error())
	} else if val > math.MaxUint16 {
		return errors.New("argument error")
	}

	orig.LITERAL = uint16(val)

	return nil
}

// Generate encodes the origin as the entry point in machine code. It should only be called as the
// first operation when generating code.
func (orig *ORIG) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	return []uint16{orig.LITERAL}, nil
}

// .STRINGZ: A directive to allocate a ASCII-encoded, zero-terminated string.
//
//	HELLO .STRINGZ "Hello, world!"
type STRINGZ struct {
	SourceInfo
	LITERAL string // Literal constant.
}

func (s *STRINGZ) Parse(opcode string, val []string) error {
	return s.ParseString(opcode, val[0])
}

func (s *STRINGZ) ParseString(opcode string, val string) error {
	s.LITERAL = strings.Trim(val, `"`)
	return nil
}

func (s *STRINGZ) Generate(symbols SymbolTable, pc uint16) ([]uint16, error) {
	code := append(utf16.Encode([]rune(s.LITERAL)), 0) // null terminate value.
	return code, nil
}

// badGPR is returned when a value is invalid because it is more noticeable than a zero value.
const badGPR = uint16(vm.BadGPR)

// registerVal returns the registerVal encoded as an integer or badGPR if the register does not
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
		return uint16(vm.BadGPR)
	}
}
