// Code generated by "stringer -type=QType"; DO NOT EDIT.

package proto

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[QTypeUnknown-0]
	_ = x[QTypeA-1]
	_ = x[QTypeNS-2]
	_ = x[QTypeMD-3]
	_ = x[QTypeMF-4]
	_ = x[QTypeCName-5]
	_ = x[QTypeSOA-6]
	_ = x[QTypeMB-7]
	_ = x[QTypeMG-8]
	_ = x[QTypeMR-9]
	_ = x[QTypeNULL-10]
	_ = x[QTypeWKS-11]
	_ = x[QTypePTR-12]
	_ = x[QTypeHINFO-13]
	_ = x[QTypeMINFO-14]
	_ = x[QTypeMX-15]
	_ = x[QTypeTXT-16]
	_ = x[QTypeAAAA-28]
	_ = x[QTypeAXFR-252]
	_ = x[QTypeMAILB-253]
	_ = x[QTypeMAILA-254]
	_ = x[QTypeALL-255]
}

const (
	_QType_name_0 = "QTypeUnknownQTypeAQTypeNSQTypeMDQTypeMFQTypeCNameQTypeSOAQTypeMBQTypeMGQTypeMRQTypeNULLQTypeWKSQTypePTRQTypeHINFOQTypeMINFOQTypeMXQTypeTXT"
	_QType_name_1 = "QTypeAAAA"
	_QType_name_2 = "QTypeAXFRQTypeMAILBQTypeMAILAQTypeALL"
)

var (
	_QType_index_0 = [...]uint8{0, 12, 18, 25, 32, 39, 49, 57, 64, 71, 78, 87, 95, 103, 113, 123, 130, 138}
	_QType_index_2 = [...]uint8{0, 9, 19, 29, 37}
)

func (i QType) String() string {
	switch {
	case i <= 16:
		return _QType_name_0[_QType_index_0[i]:_QType_index_0[i+1]]
	case i == 28:
		return _QType_name_1
	case 252 <= i && i <= 255:
		i -= 252
		return _QType_name_2[_QType_index_2[i]:_QType_index_2[i+1]]
	default:
		return "QType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}