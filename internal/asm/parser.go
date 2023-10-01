package asm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/smoynes/elsie/internal/log"
)

// Parser reads source code and produces a symbol table, a parse table and a collection of errors,
// if any. The user calls |Parse| one or more times and then ask the Parser for the accumulated
// results. Some simple syntax checking is done during parsing, but it is not complete. The second
// pass does most of syntactic analysis as well as code generation.
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
	instrTable map[string]Instruction // Generators for each opcode.
	symbols    SymbolTable            // Symbolic references.
	instr      []Instruction          // Parsed instructions.

	fatal error   // Error causing parsing to halt.
	errs  []error // Syntax errors.

	log *log.Logger
}

// AddOperatorForTesting updates the operator table for the sake of testing the parser.
func AddOperatorForTesting(op string, ins Instruction) {
	instructionTable[op] = ins
	fmt.Printf("%#v\n%#v\n", op, instructionTable)
}

func NewParser(log *log.Logger) *Parser {
	return &Parser{
		symbols:    make(SymbolTable),
		instrTable: instructionTable,
		log:        log,
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
	p.symbols[sym] = loc
}

// Instructions returns the abstract syntax "tree".
func (p *Parser) Instructions() []Instruction {
	return p.instr
}

// Add instruction appends an instruction to the list of instructions.
func (p *Parser) AddInstruction(inst Instruction) {
	if inst == nil {
		panic("nil instruction")
	}
	p.instr = append(p.instr, inst)
}

// SyntaxError adds an error to the parser errors.
func (p *Parser) SyntaxError(loc uint16, pos uint16, line string, err error) {
	p.errs = append(p.errs, &SyntaxError{Loc: loc, Pos: pos, Line: line, Err: err})
}

// Err returns errors that occur during parsing. If a fatal error occurs that prevents parsing from
// continuing (e.g., a fs.PathError), the error is returned. Otherwise, the parser collects syntax
// errors during parsing and returns an error that wraps and joins them all. Callers can inspect the
// cause with the errors package.
func (p *Parser) Err() error {
	return errors.Join(p.errs...)
}

// Parse parses an input stream. The parser takes ownership of the stream and will close it.
func (p *Parser) Parse(in io.ReadCloser) {
	defer func() {
		_ = in.Close()
	}()

	lines := bufio.NewScanner(in)

	// Keep track of our location in memory, our line number in the source and any parsing errors.
	var (
		loc uint16 // Location counter.
		pos uint16 // Line number.
	)

scan:
	for {
		scanned := lines.Scan()
		line := lines.Text()
		pos++

		switch {
		case !scanned: // No more progress.
			break scan
		default:
			var fatal error
			loc, fatal = p.parseLine(loc, pos, line)

			if fatal != nil {
				p.fatal = fatal
				return
			}
		}
	}

	return
}

var (
	// Grammar terminals.
	text       = `(.*)`
	space      = `[\pZ\p{Cc}]*`
	ident      = `(\pL[\pL\p{Nd}\pM\p{Pc}\p{Pd}\pS]*)`
	literal    = `(^\p{Nd}+|^0[xob]\p{Nd}+|^'.*')`
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

// Parse line uses regular expressions to parse text.
func (p *Parser) parseLine(loc uint16, pos uint16, line string) (uint16, error) {
	var (
		label  string       // Label, if any.
		remain string       // Remaining unparsed line.
		next   uint16 = loc // Next location value.
		err    error
	)

	remain = strings.TrimSpace(line)

	if matched := commentPattern.FindStringIndex(remain); len(matched) > 1 {
		remain = remain[:matched[0]] // Discard comments.

	}

	if matched := labelPattern.FindStringSubmatchIndex(remain); len(matched) > 1 {
		var (
			matchEnd             = matched[1]
			labelStart, labelEnd = matched[2], matched[3]
		)

		label = remain[labelStart:labelEnd]
		label = strings.TrimSpace(label)
		label = strings.ToUpper(label)

		if p.isReservedKeyword(label) {
			label = ""
		} else {
			remain = remain[matchEnd:]
			p.symbols[label] = loc
		}

	}

	if matched := directivePattern.FindStringSubmatch(remain); len(matched) > 1 {
		ident := matched[1]
		ident = strings.TrimSpace(ident)
		ident = strings.ToUpper(ident)

		arg := matched[2]
		arg = strings.TrimSpace(arg)

		next, err = p.parseDirective(ident, arg, loc)
		if err != nil {
			err := fmt.Errorf("parser error: %w", err)
			p.SyntaxError(loc, pos, line, err)
		}

	}

	if matched := instructionPattern.FindStringSubmatch(remain); len(matched) > 2 {
		var inst Instruction

		operator := matched[1]
		operands := strings.Split(matched[2], ",")

		for i := range operands {
			operands[i] = strings.TrimSpace(operands[i])
		}

		inst, err = p.parseInstruction(operator, operands)
		if err != nil {
			p.SyntaxError(loc, pos, line, err)
		} else {
			p.AddInstruction(inst)
			next += 1
		}
	}

	return next, err
}

// parseInstruction dispatches parsing to an instruction parser based on the operator.
func (p *Parser) parseInstruction(operator string, operands []string) (Instruction, error) {
	proto, ok := p.instrTable[operator]
	if !ok {
		return nil, fmt.Errorf("operator: %s", operator)
	}

	inst, err := proto.Parse(operator, operands)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (p *Parser) isReservedKeyword(word string) bool {
	for i := range directives {
		if directives[i] == word {
			return true
		}
	}

	if _, instr := p.instrTable[word]; instr {
		return true
	}

	return false
}

// parseDirective parses a directive or pseudo-instructions from its identifier and argument. The
// directive may modify parser state by taking the location counter and returning new value.
func (p *Parser) parseDirective(ident string, arg string, loc uint16) (uint16, error) {
	switch ident {
	case ".ORIG":

		if len(arg) < 1 {
			return loc, errors.New("argument error")
		}

		if arg[0] == 'x' {
			arg = "0" + arg
		}

		if val, err := strconv.ParseInt(arg, 0, 16); err != nil {
			return loc, err
		} else if loc < 0 || loc > math.MaxUint16 {
			return loc, errors.New("directive error")
		} else {
			loc = uint16(val)

			return loc, nil
		}
	case ".DW":
		// TODO:??
		return loc + 1, nil
	case ".END":
		return loc, nil // TODO: stop parsing
	default:
		//return loc, errors.New("directive error")
		return loc, nil
	}
}

// SymbolTable maps a symbol reference to its location in object code.
type SymbolTable map[string]uint16

type SyntaxError struct {
	Loc, Pos uint16
	Line     string
	Err      error
}

func (pe *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s: line: %d %q", pe.Err, pe.Pos, pe.Line)
}
