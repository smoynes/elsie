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
//	BR    [ LABEL | #OFFSET9 ]
//	BRn   [ LABEL | #OFFSET9 ]
//	BRnz  [ LABEL | #OFFSET9 ]
//	BRz   [ LABEL | #OFFSET9 ]
//	BRzp  [ LABEL | #OFFSET9 ]
//	BRp   [ LABEL | #OFFSET9 ]
//	BRnzp [ LABEL | #OFFSET9 ]
//
//	| 0000 | NZP | OFFSET9 |
//	|------+-----+---------|
//	|15  12|11  9|8       0|
type BR struct {
	NZP    uint8
	SYMBOL string
	OFFSET uint16
}

func (br BR) String() string { return fmt.Sprintf("%#v", br) }

// Parse parses all variations of the BR* instruction based on the opcode.
func (br *BR) Parse(opcode string, opers []string) error {
	var nzp uint16

	if len(opers) != 1 {
		return ErrOperand
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
		return err
	}

	*br = BR{
		NZP:    uint8(nzp),
		SYMBOL: sym,
		OFFSET: off,
	}

	return nil
}

func (br BR) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := vm.NewInstruction(vm.BR, uint16(br.NZP)<<9)

	if br.SYMBOL != "" {
		offset, err := symbols.Offset(br.SYMBOL, pc, 9)

		if err != nil {
			return nil, fmt.Errorf("br: %w", err)
		}

		code.Operand(offset & 0x01ff)
	} else {
		if br.OFFSET > 0x01ff {
			return nil, &OffsetRangeError{
				Range:  1 << 9,
				Offset: br.OFFSET,
			}
		}

		code.Operand(vm.Word(br.OFFSET) & 0x01ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// AND: Bitwise AND binary operator.
//
//	AND DR,SR1,SR2                    ; (register mode)
//
//	| 0101 | DR | SR1 | 0 | 00 | SR2 |
//	|------+----+-----+---+----+-----|
//	|15  12|11 9|8   6| 5 |4  3|2   0|
//
//	AND DR,SR1,#IMM5                  ; (immediate mode)
//
//	| 0101 | DR | SR1 | 1 |   IMM5   |
//	|------+----+-----+---+----------|
//	|15  12|11 9|8   6| 5 |4        0|
type AND struct {
	DR      string
	SR1     string
	SR2     string // Register mode.
	SYMBOL  string // Symbolic reference.
	LITERAL uint16 // Otherwise.
}

func (and AND) String() string { return fmt.Sprintf("%#v", and) }

// Parse parses an AND instruction from its opcode and operands.
func (and *AND) Parse(oper string, opers []string) error {
	if len(opers) != 3 {
		return ErrOperand
	}

	*and = AND{
		DR:  parseRegister(opers[0]),
		SR1: parseRegister(opers[1]),
	}

	if sr2 := parseRegister(opers[2]); sr2 != "" {
		and.SR2 = sr2

		return nil
	}

	off, sym, err := parseImmediate(opers[2], 5)
	if err != nil {
		return err
	}

	and.LITERAL = off
	and.SYMBOL = sym

	return nil
}

// Generate returns the machine code for an AND instruction.
func (and AND) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
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

		code.Operand(vm.Word(sr2))
	case and.SYMBOL != "":
		code.Operand(1 << 5)

		offset, err := symbols.Offset(and.SYMBOL, pc, 5)
		if err != nil {
			return nil, fmt.Errorf("and: %w", err)
		}

		code.Operand(offset & 0x001f)
	default:
		code.Operand(1 << 5)

		if and.LITERAL > 0x001f {
			err := &OffsetRangeError{Offset: and.LITERAL, Range: 0x001f}
			return nil, fmt.Errorf("and: %w", err)
		}

		code.Operand(vm.Word(and.LITERAL) & 0x001f)
	}

	return []vm.Word{code.Encode()}, nil
}

// LD: Load from memory.
//
//	LD DR,LABEL
//	LD DR,#OFFSET9
//
//	| 0010 | DR | OFFSET9 |
//	|------+----+---------|
//	|15  12|11 9|8       0|
type LD struct {
	DR     string
	OFFSET uint16
	SYMBOL string
}

func (ld LD) String() string { return fmt.Sprintf("%#v", ld) }

func (ld *LD) Parse(opcode string, operands []string) error {
	var err error

	if strings.ToUpper(opcode) != "LD" {
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	*ld = LD{
		DR: operands[0],
	}

	ld.OFFSET, ld.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return err
	}

	return nil
}

func (ld LD) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
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
		code.Operand(vm.Word(ld.OFFSET) & 0x0ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// LDR: Load from memory, register-relative.
//
//	LDR DR,SR,LABEL
//	LDR DR,SR,#OFFSET6
//
//	| 0110 | DR | SR | OFFSET6 |
//	|------+----+----+---------|
//	|15  12|11 9|8  6|5       0|
//
// .
type LDR struct {
	DR     string
	SR     string
	OFFSET uint16
	SYMBOL string
}

func (ldr LDR) String() string { return fmt.Sprintf("%#v", ldr) }

func (ldr *LDR) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "LDR" {
		return ErrOpcode
	} else if len(operands) != 3 {
		return ErrOperand
	}

	*ldr = LDR{
		DR: operands[0],
		SR: operands[1],
	}

	ldr.OFFSET, ldr.SYMBOL, err = parseImmediate(operands[2], 6)
	if err != nil {
		return err
	}

	return nil
}

func (ldr LDR) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
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
		code.Operand(vm.Word(ldr.OFFSET) & 0x003f)
	}

	return []vm.Word{code.Encode()}, nil
}

// LEA: Load effective address.
//
//	LDR DR,LABEL
//	LDR DR,#OFFSET9
//
//	| 1110 | DR | OFFSET9 |
//	|------+----+---------|
//	|15  12|11 9|8       0|
type LEA struct {
	DR     string
	SYMBOL string
	OFFSET uint16
}

func (lea LEA) String() string { return fmt.Sprintf("%#v", lea) }

func (lea *LEA) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "LEA" {
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	*lea = LEA{
		DR: operands[0],
	}

	lea.OFFSET, lea.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return err
	}

	return nil
}

func (lea LEA) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	dr := registerVal(lea.DR)

	if dr == badGPR {
		return nil, &RegisterError{"lea", lea.DR}
	}

	code := vm.NewInstruction(vm.LEA, dr<<9)

	switch {
	case lea.SYMBOL != "":
		offset, err := symbols.Offset(lea.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(lea.OFFSET) & 0x01ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// LDI: Load indirect
//
//	LDI SR,LABEL
//	LDI SR,#OFFSET9
//
//	| 1010 | SR  | OFFSET9 |
//	|------+-----+---------|
//	|15  12|11  9|8       0|
//
// .
type LDI struct {
	DR     string
	SYMBOL string
	OFFSET uint16
}

func (ldi LDI) String() string { return fmt.Sprintf("%#v", ldi) }

func (ldi *LDI) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "LDI" {
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	*ldi = LDI{
		DR: operands[0],
	}

	ldi.OFFSET, ldi.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return err
	}

	return nil
}

func (ldi LDI) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	sr := registerVal(ldi.DR)

	if sr == badGPR {
		return nil, &RegisterError{"ldi", ldi.DR}
	}

	code := vm.NewInstruction(vm.LDI, sr<<9)

	switch {
	case ldi.SYMBOL != "":
		offset, err := symbols.Offset(ldi.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(ldi.OFFSET) & 0x01ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// ST: Store word in memory.
//
//	ST SR,LABEL
//	ST SR,#OFFSET9
//
//	| 0011 | SR  | OFFSET9 |
//	|------+-----+---------|
//	|15  12|11  9|8       0|

type ST struct {
	SR     string
	SYMBOL string
	OFFSET uint16
}

func (st ST) String() string { return fmt.Sprintf("%#v", st) }

func (st *ST) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "ST" {
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	*st = ST{
		SR: operands[0],
	}

	st.OFFSET, st.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return fmt.Errorf("st: operand error: %w", err)
	}

	return nil
}

func (st ST) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	dr := registerVal(st.SR)

	if dr == badGPR {
		return nil, &RegisterError{"st", st.SR}
	}

	code := vm.NewInstruction(vm.ST, dr<<9)

	switch {
	case st.SYMBOL != "":
		offset, err := symbols.Offset(st.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("st: %w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(st.OFFSET) & 0x01ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// STI: Store Indirect.
//
//	STI SR,LABEL
//	STI SR,#OFFSET9
//
//	| 1011 | SR  | OFFSET9 |
//	|------+-----+---------|
//	|15  12|11  9|8       0|
//
// .
type STI struct {
	SR     string
	SYMBOL string
	OFFSET uint16
}

func (sti STI) String() string { return fmt.Sprintf("%#v", sti) }

func (sti *STI) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "STI" {
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	*sti = STI{
		SR: operands[0],
	}

	sti.OFFSET, sti.SYMBOL, err = parseImmediate(operands[1], 9)
	if err != nil {
		return err
	}

	return nil
}

func (sti STI) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	dr := registerVal(sti.SR)

	if dr == badGPR {
		return nil, &RegisterError{"sti", sti.SR}
	}

	code := vm.NewInstruction(vm.STI, dr<<9)

	switch {
	case sti.SYMBOL != "":
		offset, err := symbols.Offset(sti.SYMBOL, pc, 9)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(sti.OFFSET) & 0x01ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// STR: Store Relative.
//
//	STR SR1,SR2,LABEL
//	STR SR1,SR2,#OFFSET6
//
//	| 0111 | SR1 | SR2 | OFFSET6 |
//	|------+-----+-----+---------|
//	|15  12|11  9|8   6|5       0|
//
// .
type STR struct {
	SR1    string
	SR2    string
	SYMBOL string
	OFFSET uint16
}

func (str STR) String() string { return fmt.Sprintf("%#v", str) }

func (str *STR) Parse(opcode string, operands []string) error {
	var err error

	if opcode != "STR" {
		return ErrOpcode
	} else if len(operands) != 3 {
		return ErrOperand
	}

	*str = STR{
		SR1: operands[0],
		SR2: operands[1],
	}

	str.OFFSET, str.SYMBOL, err = parseImmediate(operands[2], 6)
	if err != nil {
		return err
	}

	return nil
}

func (str STR) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	sr1 := registerVal(str.SR1)
	sr2 := registerVal(str.SR2)

	if sr1 == badGPR {
		return nil, &RegisterError{"str", str.SR1}
	} else if sr2 == badGPR {
		return nil, &RegisterError{"str", str.SR2}
	}

	code := vm.NewInstruction(vm.STR, sr1<<9|sr2<<6)

	switch {
	case str.SYMBOL != "":
		offset, err := symbols.Offset(str.SYMBOL, pc, 5)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(str.OFFSET) & 0x003f)
	}

	return []vm.Word{code.Encode()}, nil
}

// JMP: Unconditional relative branch.
//
//	JMP SR
//
//	| 1100 | 000 | SR | 00 0000 |
//	|------+-----+----+---------|
//	|15  12|11  9|8  6|5       0|
//
// .
type JMP struct {
	SR string
}

func (jmp *JMP) String() string { return fmt.Sprintf("%#v", jmp) }

func (jmp *JMP) Parse(opcode string, operands []string) error {
	if opcode != "JMP" {
		return ErrOpcode
	} else if len(operands) != 1 {
		return ErrOperand
	}

	*jmp = JMP{
		SR: operands[0],
	}

	return nil
}

func (jmp JMP) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	sr := registerVal(jmp.SR)

	if sr == badGPR {
		return nil, &RegisterError{"str", jmp.SR}
	}

	code := vm.NewInstruction(vm.JMP, sr<<6)

	return []vm.Word{code.Encode()}, nil
}

// RET: Return from subroutine.
//
//	RET
//
//	| 1100 | 000 | 111 | 00 0000 |
//	|------+-----+-----+---------|
//	|15  12|11  9|8   6|5       0|
//
// .
type RET struct{}

func (ret *RET) String() string { return fmt.Sprintf("%#v", ret) }

func (ret *RET) Parse(opcode string, operands []string) error {
	if opcode != "RET" {
		return ErrOpcode
	} else if len(operands) > 0 {
		return ErrOperand
	}

	*ret = RET{}

	return nil
}

func (ret RET) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := vm.NewInstruction(vm.RET, uint16(vm.RETP)<<6)

	return []vm.Word{code.Encode()}, nil
}

// ADD: Arithmetic addition operator.
//
//	ADD DR,SR1,SR2
//	ADD DR,SR1,#IMM5
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
	LITERAL uint16 // Literal value otherwise, immediate mode.
}

func (add ADD) String() string { return fmt.Sprintf("%#v", add) }

func (add *ADD) Parse(opcode string, operands []string) error {
	if opcode != "ADD" {
		return ErrOpcode
	} else if len(operands) != 3 {
		return ErrOperand
	}

	dr := parseRegister(operands[0])
	sr1 := parseRegister(operands[1])

	*add = ADD{
		DR:  dr,
		SR1: sr1,
	}

	if sr2 := parseRegister(operands[2]); sr2 != "" {
		add.SR2 = sr2
	} else {
		off, _, err := parseImmediate(operands[2], 5)
		if err != nil {
			return err
		}

		add.LITERAL = off & 0x1f
	}

	return nil
}

func (add ADD) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
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

		code.Operand(vm.Word(sr2))
	} else {
		code.Operand(1 << 5)
		code.Operand(vm.Word(add.LITERAL) & 0x001f)
	}

	return []vm.Word{code.Encode()}, nil
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
	LITERAL uint16
}

func (trap TRAP) String() string { return fmt.Sprintf("%#v", trap) }

func (trap *TRAP) Parse(opcode string, operands []string) error {
	switch {
	case opcode == "HALT":
		if len(operands) != 0 {
			return fmt.Errorf("HALT: %w", ErrOperand)
		}

		*trap = TRAP{LITERAL: uint16(vm.TrapHALT)}

		return nil
	case opcode == "GETC":
		if len(operands) != 0 {
			return fmt.Errorf("GETC: %w", ErrOperand)
		}

		*trap = TRAP{LITERAL: uint16(vm.TrapGETC)}

		return nil
	case opcode == "OUT":
		if len(operands) != 0 {
			return fmt.Errorf("OUT: %w", ErrOperand)
		}

		*trap = TRAP{LITERAL: uint16(vm.TrapOUT)}

		return nil
	case opcode == "TRAP":
		if len(operands) != 1 {
			return ErrOperand
		}
	default:
		return ErrOperand
	}

	lit, err := parseLiteral(operands[0], 8)
	if err != nil {
		return err
	}

	*trap = TRAP{
		LITERAL: lit,
	}

	return nil
}

func (trap TRAP) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := vm.NewInstruction(vm.TRAP, trap.LITERAL&0x00ff)
	return []vm.Word{code.Encode()}, nil
}

// RTI: Return from Trap or Interrupt
//
//	RTI
//
//	| 1000 | 0000 0000 0000 |
//	|------+----------------|
//	|15  12|11             0|
//
// .
type RTI struct{}

func (rti RTI) String() string { return fmt.Sprintf("%#v", rti) }

func (rti *RTI) Parse(opcode string, operands []string) error {
	if opcode != "RTI" {
		return errors.New("operator error")
	} else if len(operands) != 0 {
		return ErrOperand
	}

	return nil
}

func (rti RTI) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := vm.NewInstruction(vm.RTI, 0x000)
	return []vm.Word{code.Encode()}, nil
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
		return ErrOpcode
	} else if len(operands) != 2 {
		return ErrOperand
	}

	dr := parseRegister(operands[0])
	sr := parseRegister(operands[1])

	*not = NOT{
		DR: dr,
		SR: sr,
	}

	return nil
}

func (not NOT) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	if not.DR == "" || not.SR == "" {
		return nil, fmt.Errorf("not: bad operand")
	}

	dr := registerVal(not.DR)
	sr := registerVal(not.SR)

	if dr == badGPR {
		return nil, &RegisterError{"not", not.DR}
	} else if sr == badGPR {
		return nil, &RegisterError{"not", not.SR}
	}

	code := vm.NewInstruction(vm.NOT, dr<<9|sr<<6|0x003f)

	return []vm.Word{code.Encode()}, nil
}

// JSR: Jump to subroutine.
//
//	JSR LABEL
//	JSR #OFFSET11
//
//	| 0100 |  1 | OFFSET11 |
//	|------+----+----------|
//	|15  12| 11 |10       0|
type JSR struct {
	SYMBOL string
	OFFSET uint16
}

func (jsr *JSR) String() string { return fmt.Sprintf("%#v", jsr) }

func (jsr *JSR) Parse(opcode string, operands []string) error {
	if opcode != "JSR" {
		return ErrOpcode
	} else if len(operands) != 1 {
		return ErrOperand
	}

	off, sym, err := parseImmediate(operands[0], 11)
	if err != nil {
		return err
	}

	*jsr = JSR{
		OFFSET: off,
		SYMBOL: sym,
	}

	return nil
}

func (jsr JSR) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := vm.NewInstruction(vm.JSR, 1<<11)

	switch {
	case jsr.SYMBOL != "":
		offset, err := symbols.Offset(jsr.SYMBOL, pc, 11)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		code.Operand(offset)
	default:
		code.Operand(vm.Word(jsr.OFFSET) & 0x03ff)
	}

	return []vm.Word{code.Encode()}, nil
}

// JSRR: Jump to subroutine, register mode.
//
//	JSRR SR
//
//	| 0100 |  0 | 00 | SR | 0 0000 |
//	|------+----+----+----+--------|
//	|15  12| 11 |10 9|8  6|5      0|
//
// .
type JSRR struct {
	SR string
}

func (jsrr *JSRR) String() string { return fmt.Sprintf("%#v", jsrr) }

func (jsrr *JSRR) Parse(opcode string, operands []string) error {
	if opcode != "JSRR" {
		return errors.New("jsrr: opcode error")
	} else if len(operands) != 1 {
		return ErrOperand
	}

	*jsrr = JSRR{
		SR: operands[0],
	}

	return nil
}

func (jsrr JSRR) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	reg := registerVal(jsrr.SR)
	if reg == badGPR {
		return nil, &RegisterError{"jsrr", jsrr.SR}
	}

	code := vm.NewInstruction(vm.JSRR, reg<<6)

	return []vm.Word{code.Encode()}, nil
}

// .FILL: Allocate and initialize one word of data.
//
//	.FILL x1234
//	.FILL 0
type FILL struct {
	LITERAL uint16 // Literal constant.
}

func (fill *FILL) Parse(opcode string, operands []string) error {
	val, err := parseLiteral(operands[0], 16)
	fill.LITERAL = val

	return err
}

func (fill FILL) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	return []vm.Word{vm.Word(fill.LITERAL)}, nil
}

// .BLKW: Data allocation directive.
//
//	.BLKW 1
type BLKW struct {
	ALLOC vm.Word // Number of words allocated.
}

func (blkw *BLKW) String() string { return fmt.Sprintf("%#v", blkw) }

func (blkw *BLKW) Parse(opcode string, operands []string) error {
	val, err := parseLiteral(operands[0], 16)
	blkw.ALLOC = vm.Word(val)

	return err
}

func (blkw BLKW) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	code := make([]vm.Word, blkw.ALLOC)
	for i := vm.Word(0); i < blkw.ALLOC; i++ {
		code[i] = 0x2361
	}

	return code, nil
}

// .ORIG: Origin directive. Sets the location counter to a literal value.
//
//	.ORIG x1234
//	.ORIG 0
type ORIG struct {
	LITERAL vm.Word // Literal constant.
}

func (orig *ORIG) Is(target Operation) bool {
	if _, ok := target.(*ORIG); ok {
		return true
	} else if target, ok := target.(interface{ Is(Operation) bool }); ok {
		return target.Is(orig)
	}

	return false
}

func (orig *ORIG) Parse(opcode string, operands []string) error {
	if opcode != ".ORIG" {
		return ErrOpcode
	} else if len(operands) != 1 {
		return ErrOperand
	}

	arg := operands[0]

	switch arg[0] {
	case 'x', 'b', 'o':
		arg = "0" + arg
	}

	val, err := strconv.ParseUint(arg, 0, 16)

	if numError := (&strconv.NumError{}); errors.As(err, &numError) {
		// TODO: err types
		return fmt.Errorf("parse error: %s (%s)", numError.Num, numError.Err.Error())
	} else if val > math.MaxUint16 {
		return errors.New("argument error")
	}

	orig.LITERAL = vm.Word(val)

	return nil
}

// Generate encodes the origin as the entry point in machine code. It should only be called as the
// first operation when generating code.
func (orig ORIG) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	return []vm.Word{orig.LITERAL}, nil
}

// .STRINGZ: A directive to allocate a ASCII-encoded, zero-terminated string.
//
//	HELLO .STRINGZ "Hello, world!"
type STRINGZ struct {
	LITERAL string // Literal constant.
}

func (s *STRINGZ) Parse(opcode string, val []string) error {
	return s.ParseString(opcode, val[0])
}

func (s *STRINGZ) ParseString(opcode string, val string) error {
	s.LITERAL = strings.Trim(val, `"`)
	return nil
}

func (s STRINGZ) Generate(symbols SymbolTable, pc vm.Word) ([]vm.Word, error) {
	encoded := append(utf16.Encode([]rune(s.LITERAL)), 0)
	code := make([]vm.Word, len(encoded))

	for i := range encoded {
		code[i] = vm.Word(encoded[i])
	}

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
