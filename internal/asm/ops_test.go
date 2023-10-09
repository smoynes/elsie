package asm

import (
	"errors"
	"reflect"
	"testing"
)

type parserCase struct {
	name     string
	opcode   string
	operands []string

	want    Operation
	wantErr error
}

func TestAND_Parse(t *testing.T) {
	// I still not sure I like this style of table tests.
	tests := []struct {
		name     string
		opcode   string
		operands []string
		want     Operation
		wantErr  bool
	}{
		{
			name:     "immediate decimal",
			opcode:   "AND",
			operands: []string{"R0", "R1", "#12"},
			want: &AND{
				DR:     "R0",
				SR1:    "R1",
				OFFSET: uint16(12),
			},
			wantErr: false,
		},
		{
			name:   "immediate hex",
			opcode: "AND", operands: []string{"R0", "R2", "#x1f"},
			want: &AND{
				DR:     "R0",
				SR1:    "R2",
				OFFSET: 0x1f,
			},
			wantErr: false,
		},
		{
			name:   "immediate octal",
			opcode: "AND", operands: []string{"R0", "R3", "#o12"},
			want: &AND{
				DR:     "R0",
				SR1:    "R3",
				OFFSET: 0o12,
			},
			wantErr: false,
		},
		{
			name:   "immediate binary",
			opcode: "AND", operands: []string{"R0", "R4", "#b01111"},
			want: &AND{
				DR:     "R0",
				SR1:    "R4",
				OFFSET: 0b1111,
			},
			wantErr: false,
		},
		{
			name:   "immediate symbol",
			opcode: "AND", operands: []string{"R0", "R4", "LABEL"},
			want: &AND{
				DR:     "R0",
				SR1:    "R4",
				SYMBOL: "LABEL",
			},
			wantErr: false,
		},
		{
			name:   "register",
			opcode: "AND", operands: []string{"R0", "R1", "R2"},
			want: &AND{
				DR:  "R0",
				SR1: "R1",
				SR2: "R2",
			},
			wantErr: false,
		},
		{
			name:   "no operands",
			opcode: "AND", operands: nil,
			wantErr: true,
		},
		{
			name:   "one operand",
			opcode: "AND", operands: []string{"R0"},
			wantErr: true,
		},
		{
			name:   "two operands",
			opcode: "AND", operands: []string{"R0", "R0"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &AND{}

			err := got.Parse(tt.opcode, tt.operands)

			if (err != nil) != tt.wantErr {
				t.Errorf("AND.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AND.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestBR_Parse(t *testing.T) {
	type args struct {
		oper  string
		opers []string
	}

	tests := []struct {
		name    string
		args    args
		want    Operation
		wantErr bool
	}{
		{
			name:    "bad oper",
			args:    args{"OP", []string{"IDENT"}},
			want:    &BR{},
			wantErr: true,
		},
		{
			name:    "BR label",
			args:    args{"BR", []string{"LABEL"}},
			want:    &BR{NZP: 0x7, SYMBOL: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BR offset",
			args:    args{"BR", []string{"#b10000"}},
			want:    &BR{NZP: 0x7, OFFSET: 0b0001_0000},
			wantErr: false,
		},
		{
			name:    "BRNZP",
			args:    args{"BRNZP", []string{"#b00010001"}},
			want:    &BR{NZP: 0x7, OFFSET: 0b1_0001},
			wantErr: false,
		},
		{
			name:    "BRz",
			args:    args{"BRz", []string{"#b10010"}},
			want:    &BR{NZP: 0x2, OFFSET: 0x12},
			wantErr: false,
		},
		{
			name:    "BRnz",
			args:    args{"BRnz", []string{"#b10011"}},
			want:    &BR{NZP: 0x6, OFFSET: 0x13},
			wantErr: false,
		},
		{
			name:    "BRnzp symbol",
			args:    args{"BRnzp", []string{"LABEL"}},
			want:    &BR{NZP: 0x7, SYMBOL: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRzp",
			args:    args{"BRzp", []string{"LABEL"}},
			want:    &BR{NZP: 0x3, SYMBOL: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRn",
			args:    args{"BRN", []string{"LABEL"}},
			want:    &BR{NZP: 0x4, SYMBOL: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRp",
			args:    args{"BRp", []string{"LABEL"}},
			want:    &BR{NZP: 0x1, SYMBOL: "LABEL"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &BR{}

			err := got.Parse(tt.args.oper, tt.args.opers)
			if (err != nil) != tt.wantErr {
				t.Errorf("BR.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BR.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLD_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "LD label",
			opcode: "LD", operands: []string{"DR", "LABEL"},
			want:    &LD{DR: "DR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &LD{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("LD.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LD.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLDR_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "LDR label",
			opcode: "LDR", operands: []string{"DR", "SR", "LABEL"},
			want:    &LDR{DR: "DR", SR: "SR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "LDR literal",
			opcode: "LDR", operands: []string{"DR", "SR", "#-1"},
			want:    &LDR{DR: "DR", SR: "SR", OFFSET: 0x3f},
			wantErr: nil,
		},
		{
			name:   "LDR literal too large",
			opcode: "LDR", operands: []string{"DR", "SR", "#x40"},
			want:    &LDR{DR: "DR", SR: "SR", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "LDR literal too negative",
			opcode: "LDR", operands: []string{"DR", "SR", "#-64"},
			want:    &LDR{DR: "DR", SR: "SR", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &LDR{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("LDR.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDR.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLEA_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "LEA label",
			opcode: "LEA", operands: []string{"DR", "LABEL"},
			want:    &LEA{DR: "DR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "LEA literal",
			opcode: "LEA", operands: []string{"DR", "#-1"},
			want:    &LEA{DR: "DR", OFFSET: 0x01ff},
			wantErr: nil,
		},
		{
			name:   "LEA literal too large",
			opcode: "LEA", operands: []string{"DR", "SR", "#x40"},
			want:    &LEA{DR: "DR", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "LEA literal too negative",
			opcode: "LEA", operands: []string{"DR", "SR", "#-64"},
			want:    &LEA{DR: "DR", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &LEA{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("LEA.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LEA.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLDI_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "LDI label",
			opcode: "LDI", operands: []string{"SR", "LABEL"},
			want:    &LDI{SR: "SR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "LDI literal",
			opcode: "LDI", operands: []string{"SR", "#-1"},
			want:    &LDI{SR: "SR", OFFSET: 0x01ff},
			wantErr: nil,
		},
		{
			name:   "LDI literal too large",
			opcode: "LDI", operands: []string{"SR", "#x0200"},
			want:    &LDI{SR: "SR", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "LDI literal too negative",
			opcode: "LDI", operands: []string{"SR", "#xff00"},
			want:    &LDI{SR: "SR", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &LDI{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("LDI.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LDI.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestST_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "ST label",
			opcode: "ST", operands: []string{"SR", "LABEL"},
			want:    &ST{SR: "SR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "ST literal",
			opcode: "ST", operands: []string{"SR", "#-1"},
			want:    &ST{SR: "SR", OFFSET: 0x01ff},
			wantErr: nil,
		},
		{
			name:   "ST literal too large",
			opcode: "ST", operands: []string{"SR", "#x0200"},
			want:    &ST{SR: "SR", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "ST literal too negative",
			opcode: "ST", operands: []string{"SR", "#xff00"},
			want:    &ST{SR: "SR", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &ST{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("ST.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ST.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSTI_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "STI label",
			opcode: "STI", operands: []string{"SR", "LABEL"},
			want:    &STI{SR: "SR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "STI literal",
			opcode: "STI", operands: []string{"SR", "#-1"},
			want:    &STI{SR: "SR", OFFSET: 0x01ff},
			wantErr: nil,
		},
		{
			name:   "STI literal too large",
			opcode: "STI", operands: []string{"SR", "#x0200"},
			want:    &STI{SR: "SR", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "STI literal too negative",
			opcode: "STI", operands: []string{"SR", "#xff00"},
			want:    &STI{SR: "SR", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &STI{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("STI.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("STI.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSTR_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "STR label",
			opcode: "STR", operands: []string{"SR", "SR2", "LABEL"},
			want:    &STR{SR1: "SR", SR2: "SR2", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: nil,
		},
		{
			name:   "STR literal",
			opcode: "STR", operands: []string{"SR", "SR2", "#-1"},
			want:    &STR{SR1: "SR", SR2: "SR2", OFFSET: 0x03f},
			wantErr: nil,
		},
		{
			name:   "STR literal too large",
			opcode: "STR", operands: []string{"SR", "SR2", "#x0200"},
			want:    &STR{SR1: "SR", SR2: "SR2", OFFSET: 0x00},
			wantErr: &SyntaxError{},
		},
		{
			name:   "STR literal too negative",
			opcode: "STR", operands: []string{"SR", "SR2", "#xff00"},
			want:    &STR{SR1: "SR", SR2: "SR2", OFFSET: 0x3f},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &STR{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)

				t.Errorf("STR.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("STR.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestJMP_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "JMP register",
			opcode: "JMP", operands: []string{"SR"},
			want: &JMP{SR: "SR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &JMP{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Errorf("JMP.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JMP.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestADD_Parse(t *testing.T) {
	tcs := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "ADD register",
			opcode: "ADD", operands: []string{"R0", "R0", "R1"},
			want:    &ADD{DR: "R0", SR1: "R0", SR2: "R1"},
			wantErr: nil,
		},
		{
			name:   "ADD literal",
			opcode: "ADD", operands: []string{"R0", "R1", "#-1"},
			want:    &ADD{DR: "R0", SR1: "R1", LITERAL: 0x001f},
			wantErr: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := &ADD{}
			err := got.Parse(tc.opcode, tc.operands)

			if (tc.wantErr != nil && err == nil) || err != nil && tc.wantErr == nil {
				t.Errorf("ADD.Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ADD.Parse() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestNOT_Parse(t *testing.T) {
	tcs := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"IDENT"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "too few operands",
			opcode: "NOT", operands: []string{"DR"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "too many operands",
			opcode: "NOT", operands: []string{"OP", "DR", "SR1", "SR2"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "NOT register",
			opcode: "NOT", operands: []string{"R6", "R2"},
			want:    &NOT{DR: "R6", SR: "R2"},
			wantErr: nil,
		},
		{
			name:   "NOT label",
			opcode: "NOT", operands: []string{"R7", "R0", "LABEL"},
			want:    &NOT{DR: "R7", SR: "R0"},
			wantErr: &SyntaxError{},
		},
		{
			name:   "NOT literal",
			opcode: "NOT", operands: []string{"R0", "0x0"},
			want:    &NOT{DR: "R0", SR: ""},
			wantErr: nil, // This is a semantic, not syntactic error. 🤔
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := &NOT{}
			err := got.Parse(tc.opcode, tc.operands)

			if (tc.wantErr != nil && err == nil) || err != nil && tc.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tc.wantErr)

				t.Errorf("NOT.Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if (err == nil) && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NOT.Parse() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestTRAP_Parse(t *testing.T) {
	tests := []parserCase{
		{
			name:   "bad oper",
			opcode: "OP", operands: []string{"x21"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "too few operands",
			opcode: "TRAP", operands: []string{},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "too many operands",
			opcode: "TRAP", operands: []string{"x25", "x21"},
			want:    nil,
			wantErr: &SyntaxError{},
		},
		{
			name:   "TRAP",
			opcode: "TRAP", operands: []string{"x25"},
			want:    &TRAP{LITERAL: 0x0025},
			wantErr: nil,
		},
		{
			name:   "TRAP literal too big",
			opcode: "TRAP", operands: []string{"x100"},
			want:    &TRAP{LITERAL: 0x00ff},
			wantErr: &SyntaxError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &TRAP{}
			err := got.Parse(tt.opcode, tt.operands)

			if (tt.wantErr != nil && err == nil) || err != nil && tt.wantErr == nil {
				t.Fatalf("not expected: %#v, want: %#v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil && errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err: %#v, got: %#v", tt.wantErr, err)
			}

			if (err == nil) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NOT.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestTRAP_Generate(t *testing.T) {
	tcs := []struct {
		op   Operation
		want uint16
	}{
		{
			op:   &TRAP{LITERAL: 0x00ff},
			want: 0xf0ff,
		},
		{
			op:   &TRAP{LITERAL: 0x0025},
			want: 0xf025,
		},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{}

	for tc := range tcs {
		op, exp := tcs[tc].op, tcs[tc].want

		mc, err := op.Generate(symbols, pc)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if mc == nil {
			t.Error("invalid machine code")
		}

		if len(mc) != 1 {
			t.Errorf("incorrect machine code: %d bytes", len(mc))
		}

		if mc[0] != exp {
			t.Errorf("incorrect machine code: want: %0#4x, got: %0#4x", exp, mc)
		}
	}
}
