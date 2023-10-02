// inst.go implements parsing and code generation for each instruction opcode.

package asm

import (
	"reflect"
	"testing"
)

func TestAND_Parse(t *testing.T) {
	// I still not sure I like this style of table tests.
	ins := AND{}

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
			name: "immediate decimal",
			args: args{"AND", []string{"R0", "R1", "#123"}},
			want: &AND{
				Mode:   ImmediateMode,
				DR:     "R0",
				SR1:    "R1",
				OFFSET: uint16(123) | 0xffc0,
			},
			wantErr: false,
		},
		{
			name: "immediate hex",
			args: args{"AND", []string{"R0", "R2", "#x123"}},
			want: &AND{
				Mode:   ImmediateMode,
				DR:     "R0",
				SR1:    "R2",
				OFFSET: 0x123 & 0x001f,
			},
			wantErr: false,
		},
		{
			name: "immediate octal",
			args: args{"AND", []string{"R0", "R3", "#o123"}},
			want: &AND{
				Mode:   ImmediateMode,
				DR:     "R0",
				SR1:    "R3",
				OFFSET: 0o123 | 0xffe0,
			},
			wantErr: false,
		},
		{
			name: "immediate binary",
			args: args{"AND", []string{"R0", "R4", "#b111"}},
			want: &AND{
				Mode:   ImmediateMode,
				DR:     "R0",
				SR1:    "R4",
				OFFSET: 0b111,
			},
			wantErr: false,
		},
		{
			name: "immediate symbol",
			args: args{"AND", []string{"R0", "R4", "LABEL"}},
			want: &AND{
				Mode:   ImmediateMode,
				DR:     "R0",
				SR1:    "R4",
				SYMBOL: "LABEL",
			},
			wantErr: false,
		},
		{
			name: "register",
			args: args{"AND", []string{"R0", "R1", "R2"}},
			want: &AND{
				Mode: RegisterMode,
				DR:   "R0",
				SR1:  "R1",
				SR2:  "R2",
			},
			wantErr: false,
		},
		{
			name:    "no operands",
			args:    args{"AND", nil},
			wantErr: true,
		},
		{
			name:    "one operand",
			args:    args{"AND", []string{"R0"}},
			wantErr: true,
		},
		{
			name:    "two operands",
			args:    args{"AND", []string{"R0", "R0"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ins.Parse(tt.args.oper, tt.args.opers)

			if (err != nil) != tt.wantErr {
				t.Errorf("AND.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AND.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestBR_Parse(t *testing.T) {
	br := BR{}

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
			want:    &BR{NZP: 0x4, OFFSET: 0x12},
			wantErr: false,
		},
		{
			name:    "BRzn",
			args:    args{"BRzn", []string{"#b10011"}},
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
			want:    &BR{NZP: 0x5, SYMBOL: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRn",
			args:    args{"BRN", []string{"LABEL"}},
			want:    &BR{NZP: 0x2, SYMBOL: "LABEL"},
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
			got, err := br.Parse(tt.args.oper, tt.args.opers)
			if (err != nil) != tt.wantErr {
				t.Errorf("BR.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BR.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestAND_Generate(t *testing.T) {
	var instrs = []Operation{
		&AND{Mode: ImmediateMode, DR: "R0", SR1: "R7", OFFSET: 0x10},
		&AND{Mode: ImmediateMode, DR: "R0", SR1: "R7", SYMBOL: "LABEL"},
		&BR{NZP: 0x3, SYMBOL: "LABEL"},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3100,
	}

	if mc, err := instrs[0].Generate(symbols, pc); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[0], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}

		if mc != 0x61f0 {
			t.Errorf("bad maching code: %04X", mc)
		}
	}

	if mc, err := instrs[1].Generate(symbols, pc); err != nil {
		t.Fatalf("Code: %#v == error    ==> %s", instrs[1], err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[1], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}

		if mc != 0x71e0 {
			t.Errorf("bad maching code: %04x", mc)
		}
	}
}

func TestLD_Parse(t *testing.T) {
	ld := LD{}

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
			want:    nil,
			wantErr: true,
		},
		{
			name:    "LD label",
			args:    args{"LD", []string{"DR", "LABEL"}},
			want:    &LD{DR: "DR", OFFSET: 0, SYMBOL: "LABEL"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ld.Parse(tt.args.oper, tt.args.opers)

			if (err != nil) != tt.wantErr {
				t.Errorf("LD.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LD.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLD_Generate(t *testing.T) {
	var instrs = []Operation{
		&LD{DR: "R0", OFFSET: 0x10},
		&LD{DR: "R7", SYMBOL: "LABEL"},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3100,
	}

	if mc, err := instrs[0].Generate(symbols, pc); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[0], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}

		if mc != 0x2010 {
			t.Errorf("bad machine code: %04x", mc)
		}
	}

	if mc, err := instrs[1].Generate(symbols, pc); err != nil {
		t.Fatalf("Code: %#v == error    ==> %s", instrs[1], err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[1], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}

		if mc != 0x2f00 {
			t.Errorf("bad machine code: %04x", mc)
		}
	}
}
