package asm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/smoynes/elsie/internal/log"
)

// Parser reads source code and produces a symbol table, a syntax table and a collection of errors,
// if any. The user calls |Parse| one or more times and then asks the Parser for the accumulated
// results. Some simple syntax checking is done during parsing, but it is not complete. The second
// pass does most of semantic analysis in addition to code generation.
//
//	p := NewParser(logger)
//	_ = p.Parse(os.Open("file1.asm"))
//	_ = p.Parse(os.Open("file2.asm"))
//	_ = p.Parse(os.Open("file3.asm"))
//
//	err := err.Err()
//	println(errors.Is(err, SyntaxError{})) // true
//	for _, err := range err.(interface { Unwrap() []error }).Unwrap() {
//		println(err.Error()) // SyntaxError
//	}
type Parser struct {
	loc     uint16      // Location counter.
	pos     uint16      // Line number in source file.
	symbols SymbolTable // Symbolic references.
	syntax  []Operation // Parsed code and data indexed by its address in memory.

	fatal error   // Error causing parsing to halt, i.e., I/O errors.
	errs  []error // Syntax errors.

	// Stub opcode and instruction for testing.
	probeOpcode string
	probeInstr  Operation

	log *log.Logger
}

func NewParser(log *log.Logger) *Parser {
	return &Parser{
		symbols: make(SymbolTable),
		syntax:  make([]Operation, 0xffff),
		log:     log,
	}
}

// Symbols returns the symbol table constructed so far.
func (p *Parser) Symbols() SymbolTable {
	return p.symbols
}

// AddSymbol adds a new symbol to the symbol table.
func (p *Parser) AddSymbol(sym string, loc uint16) {
	if sym == "" {
		panic("empty symbol")
	}

	sym = strings.ToUpper(sym)
	p.symbols[sym] = loc
}

// Syntax returns the abstract syntax "tree". Syntax holds the parsed code and data and is used by
// the second pass to create generate code and memory layout.
func (p *Parser) Syntax() []Operation {
	return p.syntax
}

// Add instruction appends an instruction to the list of instructions.
func (p *Parser) AddInstruction(inst Operation) {
	if inst == nil {
		panic("nil instruction")
	}

	p.syntax[p.loc] = inst
}

// SyntaxError adds an error to the parser errors.
func (p *Parser) SyntaxError(loc uint16, pos uint16, line string, err error) {
	p.errs = append(p.errs, &SyntaxError{Loc: loc, Pos: pos, Line: line, Err: err})
}

// Err returns errors that occur during parsing. If a fatal error occurs that prevents parsing from
// continuing (e.g., a fs.PathError), that error is returned. Otherwise, the parser collects syntax
// errors during parsing and returns an error that wraps and joins them all. Callers can inspect the
// cause with the errors package.
func (p *Parser) Err() error {
	if p.fatal != nil {
		return p.fatal
	}

	return errors.Join(p.errs...)
}

// Probe adds a stub instruction to the parser for the sake of testing.
func (p *Parser) Probe(opcode string, ins Operation) {
	p.probeOpcode = strings.ToUpper(opcode)
	p.probeInstr = ins
}

// SymbolTable maps a symbol reference to its location in object code.
type SymbolTable map[string]uint16

// SyntaxError is a wrapped error returned when the parser encounters a syntax error.
type SyntaxError struct {
	Loc, Pos uint16
	Line     string
	Err      error
}

func (pe *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s: line: %d %q", pe.Err, pe.Pos, pe.Line)
}

// Parse parses an input stream. The parser takes ownership of the stream and will close it.
func (p *Parser) Parse(in io.ReadCloser) {
	defer func() {
		_ = in.Close()
	}()

	lines := bufio.NewScanner(in)
	if err := lines.Err(); err != nil {
		p.fatal = err
		return
	}

	for {
		scanned := lines.Scan()

		if err := lines.Err(); err != nil {
			p.fatal = err
			break
		}

		line := lines.Text()
		p.pos++

		if !scanned {
			break
		}

		if err := p.parseLine(line); err != nil {
			// Assume descendant accumulated syntax errors and that any errors returned are
			// therefore fatal.
			p.fatal = err
			return
		}
	}
}

// Parse line uses regular expressions to parse text. Based on the which patterns match, the text is
// parsed and the parser state is updated.
func (p *Parser) parseLine(line string) error {
	remain := strings.TrimSpace(line) // Remaining, unparsed line.

	if matched := commentPattern.FindStringIndex(remain); len(matched) > 1 {
		remain = remain[:matched[0]] // Discard comments.
	}

	if matched := labelPattern.FindStringSubmatchIndex(remain); len(matched) > 1 {
		var (
			matchEnd             = matched[1]
			labelStart, labelEnd = matched[2], matched[3]
		)

		label := remain[labelStart:labelEnd]
		label = strings.TrimSpace(label)
		label = strings.ToUpper(label)

		if !p.isReservedKeyword(label) {
			remain = remain[matchEnd:]
			p.symbols[label] = p.loc
		}
	}

	if matched := directivePattern.FindStringSubmatch(remain); len(matched) > 1 {
		ident := matched[1]
		ident = strings.TrimSpace(ident)
		ident = strings.ToUpper(ident)

		arg := matched[2]
		arg = strings.TrimSpace(arg)

		if err := p.parseDirective(ident, arg); err != nil {
			p.fatal = err
			return err
		}
	}

	if matched := instructionPattern.FindStringSubmatch(remain); len(matched) > 2 {
		operator := matched[1]
		operands := strings.Split(matched[2], ",")

		for i := range operands {
			operands[i] = strings.TrimSpace(operands[i])
		}

		if err := p.parseInstruction(operator, operands); err != nil {
			p.SyntaxError(p.loc, p.pos, line, err)
		}
	}

	return nil
}

// Parser regular expressions.
var (
	// Grammar terminals.
	text       = `(.*)`
	space      = `[\pZ\p{Cc}]*`
	ident      = `(\pL[\pL\p{Nd}\pM\p{Pc}\p{Pd}\pS]*)`
	directives = []string{
		`\.ORIG`,
		`\.DW`,
		`\.FILL`,
		`\.BLKW`,
		`\.STRINGZ`,
		`\.END`,
	}

	// Grammar patterns.
	commentPattern   = regexp.MustCompile(space + `;` + text + `$`)
	labelPattern     = regexp.MustCompile(`^` + ident + space + `:?` + space)
	directivePattern = regexp.MustCompile(
		`^(` + strings.Join(directives, `|`) + `)` + space + text + `$`)
	instructionPattern = regexp.MustCompile(`^` + space + ident + space + text + `$`)
)

// parseInstruction dispatches parsing to an instruction parser based on the opcode. Parsing the
// operands is delegated to the dispatched parser.
func (p *Parser) parseInstruction(opcode string, operands []string) error {
	oper := p.parseOperator(opcode)
	if oper == nil {
		return errors.New("parse: operator error")
	}

	err := oper.Parse(opcode, operands)
	if err != nil {
		return fmt.Errorf("opcode: %s: %w", opcode, err)
	}

	p.AddInstruction(oper)
	p.loc++

	return nil
}

// parseOperator returns the operation for the given opcode or an error if there is no such
// operation.
func (p *Parser) parseOperator(opcode string) Operation {
	switch strings.ToUpper(opcode) {
	case "ADD":
		return &ADD{}
	case "AND":
		return &AND{}
	case "BR", "BRNZP", "BRN", "BRZ", "BRP", "BRZN", "BRNP", "BRZP":
		return &BR{}
	case "NOT":
		return &NOT{}
	case "LD":
		return &LD{}
	case "LDR":
		return &LDR{}
	case "TRAP":
		return &TRAP{}
	case p.probeOpcode:
		return p.probeInstr
	default:
		return nil
	}
}

// Returns true if word is a reserved keyword: an opcode, a directive or an otherwise invalid symbol
// name.
func (p *Parser) isReservedKeyword(word string) bool {
	for i := range directives {
		if directives[i] == word {
			return true
		}
	}

	return p.parseOperator(word) != nil
}

// parseDirective parses a directive, or pseudo-instruction, by its identifier and argument.
func (p *Parser) parseDirective(ident string, arg string) error {
	var err error

	switch ident {
	case ".ORIG":
		orig := ORIG{}

		err = orig.Parse(ident, []string{arg})
		if err != nil {
			break
		}

		p.loc = orig.LITERAL
	case ".BLKW":
		blkw := BLKW{}

		err = blkw.Parse(ident, []string{arg})
		if err != nil {
			break
		}

		p.syntax[p.loc] = &blkw
		p.loc += blkw.ALLOC

	case ".FILL", ".DW":
		fill := FILL{}

		err = fill.Parse(ident, []string{arg})
		if err != nil {
			break
		}

		p.syntax[p.loc] = &fill
		p.loc++
	case ".END":
		// TODO: signal end
	default:
		return fmt.Errorf("directive error: %s", ident)
	}

	if err != nil {
		return fmt.Errorf("%s: %w", ident, err)
	}

	return nil
}

// parseRegister returns the register name from an operand or an empty value if the register does
// not exist.
func parseRegister(oper string) string {
	switch oper {
	case
		"R0", "R1", "R2", "R3",
		"R4", "R5", "R6", "R7":
		return oper
	default:
		return ""
	}
}

// parseImmediate returns a constant literal value or a symbolic reference from an operand. The
// value is taken as n bits long. Literals can take the forms:
//
//	#123
//	#-1
//	#x123
//	#o123
//	#b0101
//
// References may be in the forms:
//
//	LABEL
//	[LABEL]
func parseImmediate(oper string, n uint8) (uint16, string, error) {
	if len(oper) > 1 && oper[0] == '#' { // Immediate-mode prefix.
		val, err := parseLiteral(oper[1:], n)
		return val, "", err
	} else if len(oper) > 2 && oper[0] == '[' && oper[len(oper)-1] == ']' { // [LABEL]
		return 0, oper[1 : len(oper)-2], nil
	} else if len(oper) > 1 {
		val, err := parseLiteral(oper, n)
		if err != nil {
			return 0, oper, nil //nolint:nilerr
		}

		return val, oper, nil

	} else { // oh no
		return 0xffff, "", errors.New("operand error")
	}
}

// parseLiteral converts an operand as literal text to an n-bit integer value. If the literal cannot
// be parsed, or if the value exceeds 2ⁿ bits, an error is returned. Accepts operands in the
// forms:
//
// - x0000
// - o000
// - b01011010
// - 0
// - -1
func parseLiteral(operand string, n uint8) (uint16, error) {
	if len(operand) == 0 {
		return 0xffff, fmt.Errorf("literal error: %s", operand)
	}

	prefix := operand[0]
	literal := operand

	switch {
	case prefix == 'x':
		literal = "0" + operand
	case prefix == 'o':
		literal = "0" + operand
	case prefix == 'b':
		literal = "0" + operand
	}

	// The parsed value must not exceed n bits, i.e. its range is [0, 2ⁿ). Using strconv.Uint16
	// seems like the thing to do. However, it does not accept negative literals, e.g.
	// ADD R1,R1,#-1. So, we use n+1 here, giving us the range [-2ⁿ, 2ⁿ], and checking for overflow
	// and converting to unsigned.
	val64, err := strconv.ParseInt(literal, 0, int(n)+1)
	if err != nil {
		return 0xffff, fmt.Errorf("literal error: %s %d", operand, val64)
	}

	var bitmask int64 = 1<<n - 1

	if val64 < -bitmask || val64 > bitmask { // 0 <= val64 < 2ⁿ
		return 0xffff, fmt.Errorf("literal error: max: %x %s %x", 1<<n, operand, val64)
	}

	val16 := uint16(val64) & uint16(bitmask)

	return val16, nil
}
