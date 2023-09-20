package vm

// ops.go defines the CPU operations and their semantics.

import (
	"fmt"
)

// An Opcode identifies the instruction to be executed by the CPU. The ISA has
// 15 distinct opcodes, plus one reserved value that is undefined.
type Opcode uint16

// Opcode constants.
const (
	BR Opcode = iota << 12
	ADD
	LD
	ST
	JSR
	AND
	LDR
	STR
	RTI
	NOT
	LDI
	STI
	JMP
	RESV
	LEA
	TRAP

	// Two synthetic opcodes used for printing special cases of the above.
	JSRR = Opcode(JSR | 0x0f00)
	RET  = Opcode(JMP | 0x0f00)
)

type mo struct { // no, mo is NOT a monad. ( ._.)
	vm  *LC3
	err error
}

func (op mo) Err() error      { return op.err }
func (op *mo) Fail(err error) { op.err = err }

func (op mo) String() string {
	return fmt.Sprintf("ins: %s", op.vm.IR.Opcode())
}

// BR: Conditional branch
//
// | 0000 | NZP | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type br struct {
	mo
	cond   Condition
	offset Word
}

func (op br) String() string {
	return fmt.Sprintf("%s[cond:%s offset:%s]", op.mo.String(), op.cond.String(), op.offset.String())
}

var (
	_ executable = &br{}
)

func (op *br) Decode(vm *LC3) {
	*op = br{
		mo:     mo{vm: vm},
		cond:   vm.IR.Cond(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *br) Execute() {
	if op.vm.PSR.Any(op.cond) {
		op.vm.PC = ProgramCounter(int16(op.vm.PC) + int16(op.offset))
	}
}

// NOT: Bitwise complement operation
//
// | 1001 | DR | SR | 1 | 1 1111 |
// |------+----+----+---+--------|
// |15  12|11 9|8  6| 5 |4      0|
type not struct {
	mo
	dr GPR
	sr GPR
}

var (
	_ executable = &not{}
)

func (op *not) Decode(vm *LC3) {
	*op = not{
		mo: mo{vm: vm},
		sr: vm.IR.SR1(),
		dr: vm.IR.DR(),
	}
}

func (op *not) Execute() {
	op.vm.REG[op.dr] = op.vm.REG[op.sr] ^ 0xffff
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

// AND: Bitwise AND binary operator (registers)
//
// | 0101 | DR | SR1 | 0 | 00 | SR2 |
// |------+----+-----+---+----+-----|
// |15  12|11 9|8   6| 5 |4  3|2   0|
//
// | 0101 | DR  | SR | 1 | IMM5 | (immediate)
// |------+-----+----+---+------|
// |15  12|11  9|8  6| 5 |4    0|
type and struct {
	mo
	dest GPR
	sr1  GPR
	sr2  GPR
}

func (op *and) String() string {
	return fmt.Sprintf("%s[dr:%s sr1:%s sr2: %v]", op.mo.String(), op.dest.String(), op.sr1, op.sr2)
}

func (a *and) Decode(vm *LC3) {
	*a = and{
		mo:   mo{vm: vm},
		dest: vm.IR.DR(),
		sr1:  vm.IR.SR1(),
		sr2:  vm.IR.SR2(),
	}
}

func (op *and) Execute() {
	op.vm.REG[op.dest] = op.vm.REG[op.sr1]
	op.vm.REG[op.dest] &= op.vm.REG[op.sr2]
	op.vm.PSR.Set(op.vm.REG[op.dest])
}

type andImm struct {
	mo
	dr  GPR
	sr  GPR
	lit Word
}

func (op *andImm) String() string {
	return fmt.Sprintf("%s[dr:%s sr:%s lit: %v]", op.mo.String(), op.dr.String(), op.sr, op.lit)
}

func (a *andImm) Decode(vm *LC3) {
	*a = andImm{
		mo:  mo{vm: vm},
		dr:  vm.IR.DR(),
		sr:  vm.IR.SR1(),
		lit: vm.IR.Literal(IMM5),
	}
}

func (op *andImm) Execute() {
	op.vm.REG[op.dr] = op.vm.REG[op.sr] & Register(op.lit)
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

// ADD: Arithmetic addition operator
//
// | 0001 | DR | SR1 | 000 | SR2 |  (register mode)
// |------+----+-----+-----+-----|
// |15  12|11 9|8   6| 5  3|2   0|
//
// ADD: Arithmetic addition operator (immediate mode)
//
// | 0001 | DR  | SR | 1 | 11111 |
// |------+-----+----+---+-------|
// |15  12|11  9|8  6| 5 |4     0|
// .
type add struct {
	mo
	dr  GPR
	sr1 GPR
	sr2 GPR
}

var (
	_ executable = &add{}
)

func (op *add) Decode(vm *LC3) {
	*op = add{
		mo:  mo{vm: vm},
		dr:  vm.IR.DR(),
		sr1: vm.IR.SR1(),
		sr2: vm.IR.SR2(),
	}
}

func (op *add) Execute() {
	op.vm.REG[op.dr] = Register(int16(op.vm.REG[op.sr1]) + int16(op.vm.REG[op.sr2]))
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

type addImm struct {
	mo
	dr  GPR
	sr  GPR
	lit Word
}

var (
	_ executable = &addImm{}
)

func (op *addImm) Decode(vm *LC3) {
	*op = addImm{
		mo:  mo{vm: vm},
		dr:  vm.IR.DR(),
		sr:  vm.IR.SR1(),
		lit: vm.IR.Literal(IMM5),
	}
}

func (op *addImm) Execute() {
	op.vm.REG[op.dr] = Register(int16(op.vm.REG[op.sr]) + int16(op.lit))
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

// LD: Load word from memory.
//
// | 0010 | DR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type ld struct {
	mo
	dr     GPR
	offset Word
}

var (
	_ addressable = &ld{}
	_ fetchable   = &ld{}
)

func (op *ld) Decode(vm *LC3) {
	*op = ld{
		mo:     mo{vm: vm},
		dr:     vm.IR.DR(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *ld) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.PC) + int16(op.offset))
}

func (op *ld) FetchOperands() {
	op.vm.REG[op.dr] = op.vm.Mem.MDR
}

func (op *ld) Execute() {
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

// LDI: Load indirect
//
// | 1010 | DR | OFFSET9 |
// |------+--------------|
// |15  12|11 9|8       0|
type ldi struct {
	mo
	dr     GPR
	offset Word
}

var (
	_ addressable = &ldi{}
	_ fetchable   = &ldi{}
)

func (op *ldi) Decode(vm *LC3) {
	*op = ldi{
		mo:     mo{vm: vm},
		dr:     vm.IR.DR(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *ldi) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.PC) + int16(op.offset))
}

func (op *ldi) FetchOperands() {
	op.vm.Mem.MAR = op.vm.Mem.MDR

	if err := op.vm.Mem.Fetch(); err != nil {
		op.Fail(err)
		return
	}

	op.vm.REG[op.dr] = op.vm.Mem.MDR
}

func (op *ldi) Execute() {
	op.vm.PSR.Set(op.vm.Mem.MDR)
}

func (op *ldi) String() string {
	return fmt.Sprintf("OP: LDI (%s+%s)", op.dr, op.offset)
}

// LDR: Load Relative
//
// | 0110 | DR | BASE | OFFSET6 |
// |------+----+------+---------|
// |15  12|11 9|8    6|5       0|
type ldr struct {
	mo
	dr     GPR
	base   GPR
	offset Word
}

var (
	_ addressable = &ldr{}
	_ fetchable   = &ldr{}
)

func (op *ldr) Decode(vm *LC3) {
	*op = ldr{
		mo:     mo{vm: vm},
		dr:     vm.IR.DR(),
		base:   vm.IR.SR1(),
		offset: vm.IR.Offset(OFFSET6),
	}
}

func (op *ldr) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.REG[op.base]) + int16(op.offset))
}

func (op *ldr) FetchOperands() {
	op.vm.REG[op.dr] = op.vm.Mem.MDR
}

func (op *ldr) Execute() {
	op.vm.PSR.Set(op.vm.REG[op.dr])
}

// LEA: Load effective address
//
// | 1110 | DR | OFFSET9 |
// |------+--------------|
// |15  12|11 9|8       0|
type lea struct {
	mo
	dr     GPR
	offset Word
}

var (
	_ fetchable = &lea{}
)

func (op *lea) Decode(vm *LC3) {
	*op = lea{
		mo:     mo{vm: vm},
		dr:     vm.IR.DR(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *lea) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.PC) + int16(op.offset))
}

func (op *lea) FetchOperands() {
	op.vm.REG[op.dr] = op.vm.Mem.MDR
}

// ST: Store word in memory.
//
// | 0011 | SR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type st struct {
	mo
	sr     GPR
	offset Word
}

var (
	_ addressable = &st{}
	_ storable    = &st{}
)

func (op *st) Decode(vm *LC3) {
	*op = st{
		mo:     mo{vm: vm},
		sr:     vm.IR.SR(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *st) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.PC) + int16(op.offset))
}

func (op *st) Execute() {
	op.vm.Mem.MDR = op.vm.REG[op.sr]
}

func (op *st) StoreResult() {} // ?

// STI: Store Indirect.
//
// | 1011 | SR  | OFFSET9 |
// |------+-----+---------|
// |15  12|11  9|8       0|
type sti struct {
	mo
	sr     GPR
	offset Word
}

var (
	_ addressable = &sti{}
	_ fetchable   = &sti{}
	_ storable    = &sti{}
)

func (op *sti) Decode(vm *LC3) {
	*op = sti{
		mo:     mo{vm: vm},
		sr:     vm.IR.SR(),
		offset: vm.IR.Offset(OFFSET9),
	}
}

func (op *sti) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.PC) + int16(op.offset))
}

func (op *sti) FetchOperands() {
	op.vm.Mem.MAR = op.vm.Mem.MDR
}

func (op *sti) Execute() {
	op.vm.Mem.MDR = op.vm.REG[op.sr]
}

func (op *sti) StoreResult() {}

// STR: Store Relative.
//
// | 0111 | SR | GPR | OFFSET6 |
// |------+----+-----+---------|
// |15  12|11 9|8   6|5       0|
type str struct {
	mo
	sr     GPR
	base   GPR
	offset Word
}

var (
	_ addressable = &str{}
	_ storable    = &str{}
)

func (op *str) Decode(vm *LC3) {
	*op = str{
		mo:     mo{vm: vm},
		sr:     vm.IR.SR(),
		base:   vm.IR.SR1(),
		offset: vm.IR.Offset(OFFSET6),
	}
}

func (op *str) EvalAddress() {
	op.vm.Mem.MAR = Register(int16(op.vm.REG[op.base]) + int16(op.offset))
}

func (op *str) Execute() {
	op.vm.Mem.MDR = op.vm.REG[op.sr]
}

func (op *str) StoreResult() {}

// JMP: Unconditional branch
//
// | 1100 | 000 | SR | 00 00000 |
// |------+-----+----+----------|
// |15  12|11  9|8  6|5        0|
//
// RET: Return from subroutine
// | 1100 | 111 | SR | 00 00000 |
// |------+-----+----+----------|
// |15  12|11  9|8  6|5        0|
type jmp struct {
	mo
	sr GPR
}

var (
	_ executable = &jmp{}
)

func (op *jmp) Decode(vm *LC3) {
	*op = jmp{
		mo: mo{vm: vm},
		// TODO
		sr: GPR(vm.IR & 0x01e0 >> 6),
	}
}

func (op *jmp) Execute() {
	op.vm.PC = ProgramCounter(op.vm.REG[op.sr])
}

// JSR: Jump to subroutine (relative mode)
//
// | 0100 |  1 | OFFSET11 |
// |------+----+----------|
// |15  12| 11 |10       0|
//
// JSRR: Jump to subroutine (register mode)
//
// | 0100 |  0 | SR | 00 0000 |
// |------+----+----+---------|
// |15  12| 11 |8  6|5       0|
// .
type jsr struct {
	mo
	offset Word
}

var (
	_ executable = &jsr{}
)

func (op *jsr) Decode(vm *LC3) {
	*op = jsr{
		mo:     mo{vm: vm},
		offset: Word(vm.IR & 0x07ff),
	}
	op.offset.Sext(11)
}

func (op *jsr) Execute() {
	op.vm.REG[RETP] = Register(op.vm.PC)
	op.vm.PC = ProgramCounter(int16(op.vm.PC) + int16(op.offset))
}

type jsrr struct {
	mo
	sr GPR
}

var (
	_ executable = &jsrr{}
)

func (op *jsrr) Decode(vm *LC3) {
	*op = jsrr{
		mo: mo{vm: vm},
		sr: vm.IR.SR1(),
	}
}

func (op *jsrr) Execute() {
	op.vm.REG[RETP] = Register(op.vm.PC)
	op.vm.PC = ProgramCounter(op.vm.REG[op.sr])
}

// TRAP: System call or software interrupt.
//
// | 1111 | 0000 | VECTOR8 |
// |------+------+---------|
// |15  12|11   8|7       0|
type trap struct {
	mo
	vec  Word
	addr Word
}

func (op *trap) String() string {
	return fmt.Sprintf("TRAP: %s (%s)", op.vec, op.addr)
}

var (
	_ executable = &trap{}
)

func (op *trap) Decode(vm *LC3) {
	*op = trap{
		mo:  mo{vm: vm},
		vec: vm.IR.Vector(VECTOR8),
	}
}

type trapErr struct {
	interrupt
}

func (op *trap) Execute() {
	op.err = &trapErr{
		interrupt{
			table: TrapTable,
			vec:   op.vec,
			pc:    op.vm.PC,
			psr:   op.vm.PSR,
		},
	}
}

func (err *trapErr) Error() string {
	return fmt.Sprintf("INT: TRAP (%s:%s)", err.table, err.vec)
}

func (err *trapErr) Handle(cpu *LC3) error {
	// Switch from the user to the system stack and system privilege level
	// if it is a user trap.
	if cpu.PSR.Privilege() == PrivilegeUser {
		cpu.USP = cpu.REG[SP]
		cpu.REG[SP] = cpu.SSP
		cpu.PSR &= ^StatusUser
	}

	return err.interrupt.Handle(cpu)
}

// RTI: Return from trap or interrupt
//
// | 1000 | 0000 0000 0000 |
// |------+----------------|
// |15  12|11             0|
type rti struct{ mo }

var (
	_ executable = &rti{}
)

func (op *rti) Decode(vm *LC3) {
	op.vm = vm
}

func (op *rti) Execute() {
	if op.vm.PSR.Privilege() == PrivilegeUser {
		op.err = &pmv{
			interrupt{
				table: ExceptionServiceRoutines,
				vec:   ExceptionPMV,
				pc:    op.vm.PC,
				psr:   op.vm.PSR,
			},
		}

		return
	}

	// Restore program counter and status register. Popping might fail if the stack is empty.
	if err := op.vm.PopStack(); err != nil {
		op.Fail(err)
		return
	}

	op.vm.PC = ProgramCounter(op.vm.Mem.MDR)

	if err := op.vm.PopStack(); err != nil {
		op.Fail(err)
		return
	}

	op.vm.PSR = ProcessorStatus(op.vm.Mem.MDR)

	if op.vm.PSR.Privilege() == PrivilegeUser {
		// When dropping privileges, swap system and user stacks.
		op.vm.SSP = op.vm.REG[SP]
		op.vm.REG[SP] = op.vm.USP
	}
}

type pmv struct {
	interrupt
}

func (pmv *pmv) Error() string {
	return fmt.Sprintf("INT: PMV (%s:%s)", pmv.table, pmv.vec)
}

func (pmv *pmv) Handle(cpu *LC3) error {
	// PMV only occurs with user privileges so switch to system before
	// handling the interrupt.
	cpu.USP = cpu.REG[SP]
	cpu.REG[SP] = cpu.SSP
	cpu.PSR ^= StatusUser

	return pmv.interrupt.Handle(cpu)
}

// RESV: Reserved operator
//
// | 1101 | 0000 0000 0000 |
// |------+----------------|
// |15  12|11             0|
type resv struct{ mo }

var _ executable = &resv{}

func (op *resv) Decode(vm *LC3) {
	op.vm = vm
}

func (op *resv) Execute() {
	op.err = &xop{
		interrupt{
			table: ExceptionServiceRoutines,
			vec:   ExceptionXOP,
			pc:    op.vm.PC,
			psr:   op.vm.PSR,
		},
	}
}

type xop struct {
	interrupt
}

func (xop *xop) Error() string {
	return fmt.Sprintf("INT: XOP (%s:%s)", xop.table, xop.vec)
}

func (xop *xop) Handle(cpu *LC3) error {
	// Switch from the user to the system stack and system privilege level
	// if it is a user trap.
	if cpu.PSR.Privilege() == PrivilegeUser {
		cpu.USP = cpu.REG[SP]
		cpu.REG[SP] = cpu.SSP
		cpu.PSR ^= StatusUser
	}

	return xop.interrupt.Handle(cpu)
}
