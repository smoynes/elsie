// Code generated by "stringer -type AddressingMode -output strings_gen.go"; DO NOT EDIT.

package asm

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ImmediateMode-0]
	_ = x[RegisterMode-1]
	_ = x[IndirectMode-2]
}

const _AddressingMode_name = "ImmediateModeRegisterModeIndirectMode"

var _AddressingMode_index = [...]uint8{0, 13, 25, 37}

func (i AddressingMode) String() string {
	if i >= AddressingMode(len(_AddressingMode_index)-1) {
		return "AddressingMode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AddressingMode_name[_AddressingMode_index[i]:_AddressingMode_index[i+1]]
}