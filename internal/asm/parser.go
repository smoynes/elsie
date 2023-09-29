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

// Parser reads source code and produces a symbol table, a parse table and a collection of errors,
// if any. The user calls |Parse| one or more times and then ask the parser for the results. The
// caller may parse multiple streams and results are accumulated.
//
// Some basic syntax checking is done during parsing, but it is not complete. The second pass does
// most of analysis and code generation.
type Parser interface {
	// Parse parses an input stream. The parser takes ownership of the stream and will close it.
	Parse(in io.ReadCloser)

	// Symbols returns the symbol table.
	Symbols() SymbolTable

	// Instructions returns the parsed instructions.
	Instructions() Instructions

	// Err returns errors that occur during parsing. If an error occurs that prevents parsing from
	// continuing, for example fs.PathError, the first such error is returned. Otherwise, an error
	// is returned that callers unwrap into a slice of errors; these, in turn, may be unwrapped into
	// a SyntaxError.
	Err() error
}

type parser struct {
	symbols SymbolTable
	instr   Instructions
	fatal   error
	errs    []error

	operators map[string]Instruction
	log       *log.Logger
}

// Operators maps an opcode to a type which implements Instruction for the operator.
var operators = map[string]Instruction{
	"AND": &iAnd{},
}

func AddOperatorForTesting(op string, ins Instruction) {
	operators[op] = ins
}

func NewParser(log *log.Logger) Parser {
	return &parser{
		operators: operators,
		symbols:   make(SymbolTable),
		log:       log,
	}
}

func (p *parser) Symbols() SymbolTable {
	return p.symbols
}

func (p *parser) AddSymbol(sym string, loc int) {
	p.symbols[sym] = loc
}

func (p *parser) Instructions() Instructions {
	return p.instr
}

func (p *parser) AddInstruction(inst Instruction) {
	if inst == nil {
		panic("nil instruction")
	}
	p.instr = append(p.instr, inst)
}

func (p *parser) SyntaxError(loc int, pos int, line string) {
	p.errs = append(p.errs, &SyntaxError{Loc: loc, Pos: pos, Line: line})
}

func (p *parser) Err() error {
	return errors.Join(p.errs...)
}

func (p *parser) Parse(in io.ReadCloser) {
	defer func() {
		_ = in.Close()
	}()

	lines := bufio.NewScanner(in)

	// Keep track of our location in memory, our line number in the source and any parsing errors.
	var (
		loc int // Location counter.
		pos int // Line number.
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
	text      = `(.*)`
	space     = `[\pZ\p{Cc}]*`
	directive = `(ORIG|DW|FILL|BLKW|STRINGZ|END)`
	ident     = `(\pL[\pL\p{Nd}\pM\p{Pc}\p{Pd}\pS]*)`
	literal   = `(^\p{Nd}+|^0[xob]\p{Nd}+|^'.*')`

	commentPattern     = regexp.MustCompile(space + ";+" + text + "$")
	labelPattern       = regexp.MustCompile("^" + space + ident + space + ":")
	directivePattern   = regexp.MustCompile("^" + space + `\.` + directive + space + text + "$")
	instructionPattern = regexp.MustCompile("^" + space + ident + space + literal + "*")
)

// Parse line uses regular expressions to parse a line of source code.
func (p *parser) parseLine(loc int, pos int, line string) (int, error) {
	var (
		label  string        // Label, if any.
		remain string = line // Remaining unparsed line.
	)

	if matched := commentPattern.FindStringIndex(remain); len(matched) > 1 {
		remain = remain[:matched[0]]
	}

	if matched := directivePattern.FindStringSubmatchIndex(remain); len(matched) > 1 {
		ident := remain[matched[2]:matched[3]]
		arg := remain[matched[4]:matched[5]]
		ident = strings.ToUpper(ident)

		switch ident {
		case "ORIG":
			arg = strings.TrimSpace(arg)
			val, err := strconv.ParseInt(arg, 0, 16)
			if err != nil {
				p.errs = append(p.errs, err)
				return 0, nil
			}

			return int(val), nil
		case "DW":
			return loc + 1, nil
		}

		return loc, nil
	}

	if matched := labelPattern.FindStringSubmatchIndex(remain); len(matched) > 1 {
		start, end := matched[2], matched[3]
		label = remain[start:end]
		remain = remain[end:]

		defer func() {
			p.symbols[label] = loc
		}()
	}

	if matched := instructionPattern.FindStringSubmatch(remain); len(matched) > 2 {
		operator, operands := matched[1], matched[2]
		inst, err := p.parseInstruction(operator, operands)

		p.log.Debug("parse result", "inst", inst, "err", err, log.String("type", fmt.Sprintf("%T", err)))

		if err == nil {
			p.AddInstruction(inst)
		} else {
			p.SyntaxError(loc, pos, line)
		}

		return loc + 1, nil
	}

	if label == "" && remain != "" {
		p.SyntaxError(loc, pos, line)
	}

	return loc, nil
}

// parseInstruction parses strings for an operator and its operands and returns an instruction.
func (p *parser) parseInstruction(oper string, operands string) (Instruction, error) {
	opers := strings.Split(operands, ",")
	for i := range opers {
		opers[i] = strings.TrimSpace(opers[i])
	}

	if proto, ok := p.operators[oper]; ok {
		return proto.Parse(oper, opers)
	} else {
		return nil, fmt.Errorf("unknown operator")
	}
}

// iAnd is an AND instruction.
type iAnd struct{}

func (ins iAnd) String() string { return "AND" }

func (ins iAnd) Parse(oper string, opers []string) (Instruction, error) {
	return iAnd{}, nil
}

// SymbolTable maps symbol literal to its location.
type SymbolTable map[string]int

// Syntax is a list of the parsed instructions.
type Instructions []Instruction

type Instruction interface {
	Parse(operator string, operands []string) (Instruction, error)

	fmt.Stringer
}

type SyntaxError struct {
	Loc, Pos int
	Line     string
}

func (pe *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %d: %q", pe.Pos, pe.Line)
}
