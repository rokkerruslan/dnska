// Code generated by "stringer -type Opcode"; DO NOT EDIT.

package proto

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OpcodeQuery-0]
	_ = x[OpcodeIQuery-1]
	_ = x[OpcodeStatus-2]
}

const _Opcode_name = "OpcodeQueryOpcodeIQueryOpcodeStatus"

var _Opcode_index = [...]uint8{0, 11, 23, 35}

func (i Opcode) String() string {
	if i >= Opcode(len(_Opcode_index)-1) {
		return "Opcode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Opcode_name[_Opcode_index[i]:_Opcode_index[i+1]]
}