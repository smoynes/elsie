// inst.go implements parsing and code generation for each instruction opcode.

package asm

import (
	"reflect"
	"testing"
)

// I still don't quite like this style of table tests.
func TestAND_Parse(t *testing.T) {

	ins := AND{}

	type args struct {
		oper  string
		opers []string
	}
	tests := []struct {
		name    string
		args    args
		want    Instruction
		wantErr bool
	}{
		{
			name: "immediate decimal",
			args: args{"AND", []string{"R0", "R1", "#123"}},
			want: &AND{
				Mode: ImmediateMode,
				DR:   "R0",
				SR1:  "R1",
				LIT:  "123",
			},
			wantErr: false,
		},
		{
			name: "immediate hex",
			args: args{"AND", []string{"R0", "R2", "#x123"}},
			want: &AND{
				Mode: ImmediateMode,
				DR:   "R0",
				SR1:  "R2",
				LIT:  "x123",
			},
			wantErr: false,
		},
		{
			name: "immediate octal",
			args: args{"AND", []string{"R0", "R3", "#o123"}},
			want: &AND{
				Mode: ImmediateMode,
				DR:   "R0",
				SR1:  "R3",
				LIT:  "o123",
			},
			wantErr: false,
		},
		{
			name: "immediate binary",
			args: args{"AND", []string{"R0", "R4", "#b111"}},
			want: &AND{
				Mode: ImmediateMode,
				DR:   "R0",
				SR1:  "R4",
				LIT:  "b111",
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
		want    Instruction
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
			want:    &BR{NZP: 0x7, LIT: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BR offset",
			args:    args{"BR", []string{"#b1_0000"}},
			want:    &BR{NZP: 0x7, LIT: "b1_0000"},
			wantErr: false,
		},
		{
			name:    "BRNZP",
			args:    args{"BRNZP", []string{"#b1_0001"}},
			want:    &BR{NZP: 0x7, LIT: "b1_0001"},
			wantErr: false,
		},
		{
			name:    "BRz",
			args:    args{"BRz", []string{"#b1_0002"}},
			want:    &BR{NZP: 0x4, LIT: "b1_0002"},
			wantErr: false,
		},
		{
			name:    "BRzn",
			args:    args{"BRzn", []string{"#b1_0003"}},
			want:    &BR{NZP: 0x6, LIT: "b1_0003"},
			wantErr: false,
		},
		{
			name:    "BRnzp",
			args:    args{"BRnzp", []string{"LABEL"}},
			want:    &BR{NZP: 0x7, LIT: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRzp",
			args:    args{"BRzp", []string{"LABEL"}},
			want:    &BR{NZP: 0x5, LIT: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRn",
			args:    args{"BRN", []string{"LABEL"}},
			want:    &BR{NZP: 0x2, LIT: "LABEL"},
			wantErr: false,
		},
		{
			name:    "BRp",
			args:    args{"BRp", []string{"LABEL"}},
			want:    &BR{NZP: 0x1, LIT: "LABEL"},
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
	var instrs = []Instruction{
		&AND{Mode: ImmediateMode, DR: "R0", SR1: "R7", LIT: "#x0010"},
		&AND{Mode: ImmediateMode, DR: "R0", SR1: "R7", LIT: "LABEL"},
		&BR{NZP: 0x3, LIT: "LABEL"},
	}

	pc := uint16(0x3000)
	symbols := SymbolTable{
		"LABEL": 0x3100,
	}

	if mc, err := instrs[0].(*AND).Generate(symbols, pc); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[0], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}
	}

	if mc, err := instrs[1].(*AND).Generate(symbols, pc); err != nil {
		t.Fatalf("Code: %#v == error    ==> %s", instrs[1], err)
	} else {
		t.Logf("Code: %#v == generated ==> %0#4x", instrs[1], mc)

		if mc == 0xffff {
			t.Error("invalid machine code")
		}
	}
}
