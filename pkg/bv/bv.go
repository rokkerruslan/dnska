package bv

import (
	"encoding/binary"
	"fmt"
)

// BV mutable view of a byte slice. It allows reading and writing
// bytes in a sequential manner, maintaining a position pointer to track
// the current read/write location.
//
// Use big-endian (BE) bytes order of digital data.
type BV struct {
	b []byte
	p uint
}

// New creates a new BV from an existing byte slice. Input slice
// is not copied, so any changes to the original slice will be reflected
// in the BV.
func New(b []byte) *BV {
	return &BV{
		b: b,
		p: 0,
	}
}

// Pos returns the current position of the BV pointer.
func (bv *BV) Pos() uint {
	return bv.p
}

// Advance moves the BV pointer forward by the specified number of bytes.
func (bv *BV) Advance(n uint) bool {
	if n > uint(len(bv.b)) || bv.p > uint(len(bv.b))-n {
		return false
	}
	bv.p += n
	return true
}

// Seek sets the BV pointer to the specified position.
func (bv *BV) Seek(p uint) bool {
	if p > uint(len(bv.b)) {
		return false
	}
	bv.p = p
	return true
}

// Bytes returns the underlying byte slice of the BV up to the current position.
func (bv *BV) Bytes() []byte {
	return bv.b[:bv.p]
}

// UnreadBytes returns the remaining slice of bytes from current position.
func (bv *BV) UnreadBytes() []byte {
	return bv.b[bv.p:]
}

// Len returns the length of the underlying byte slice of the BV.
func (bv *BV) Len() int {
	return len(bv.b)
}

// Index returns the byte at the specified position without changing the BV pointer.
func (bv *BV) Index(p uint) (byte, error) {
	if p >= uint(len(bv.b)) {
		return 0, &ErrBuf{Op: "index", Pos: p}
	}

	return bv.b[p], nil
}

// Range returns a slice of bytes from the specified start position with the given
// length. Without changing the BV pointer.
func (bv *BV) Range(start, length uint) ([]byte, error) {
	if start > uint(len(bv.b)) || length > uint(len(bv.b))-start {
		return nil, &ErrBuf{Op: "range", Pos: start}
	}

	return bv.b[start : start+length], nil
}

func readN[T uint8 | uint16 | uint32 | uint64](bv *BV, n uint, decode func([]byte) T) (T, error) {
	if bv.p+n > uint(len(bv.b)) {
		return 0, &ErrBuf{Op: "read", Pos: bv.p}
	}
	bv.p += n
	return decode(bv.b[bv.p-n : bv.p]), nil
}

// Uint8 returns the uint8 representation of the next byte
// from the BV and advances the pointer by one. Returns an
// error if there are not enough bytes left in the BV.
func (bv *BV) Uint8() (uint8, error) { return readN(bv, 1, func(b []byte) uint8 { return b[0] }) }

// Uint16 returns the uint16 representation of the next two
// bytes from the BV and advances the pointer by two. Returns
// an error if there are not enough bytes left in the BV.
func (bv *BV) Uint16() (uint16, error) { return readN(bv, 2, binary.BigEndian.Uint16) }

// Uint32 returns the uint32 representation of the next four
// bytes from the BV and advances the pointer by four. Returns
// an error if there are not enough bytes left in the BV.
func (bv *BV) Uint32() (uint32, error) { return readN(bv, 4, binary.BigEndian.Uint32) }

// Uint64 returns the uint64 representation of the next eight
// bytes from the BV and advances the pointer by eight. Returns
// an error if there are not enough bytes left in the BV.
func (bv *BV) Uint64() (uint64, error) { return readN(bv, 8, binary.BigEndian.Uint64) }

func putN[T uint8 | uint16 | uint32 | uint64](bv *BV, n uint, encode func(dst []byte, v T), v T) error {
	if bv.p+n > uint(len(bv.b)) {
		return &ErrBuf{Op: "put", Pos: bv.p}
	}
	bv.p += n
	encode(bv.b[bv.p-n:bv.p], v)
	return nil
}

// PutUint8 writes a single byte to the BV at the current
// position and advances the pointer by one.
func (bv *BV) PutUint8(v uint8) error {
	return putN(bv, 1, func(dst []byte, v uint8) { dst[0] = v }, v)
}

// PutUint16 writes a uint16 to the BV in big-endian order
// at the current position and advances the pointer by two.
func (bv *BV) PutUint16(v uint16) error {
	return putN(bv, 2, func(dst []byte, v uint16) { binary.BigEndian.PutUint16(dst, v) }, v)
}

// PutUint32 writes a uint32 to the BV in big-endian order
// at the current position and advances the pointer by four.
func (bv *BV) PutUint32(v uint32) error {
	return putN(bv, 4, func(dst []byte, v uint32) { binary.BigEndian.PutUint32(dst, v) }, v)
}

// PutUint64 writes a uint64 to the BV in big-endian order
// at the current position and advances the pointer by eight.
func (bv *BV) PutUint64(v uint64) error {
	return putN(bv, 8, func(dst []byte, v uint64) { binary.BigEndian.PutUint64(dst, v) }, v)
}

type ErrBuf struct {
	Op  string
	Pos uint
}

func (e *ErrBuf) Error() string {
	return fmt.Sprintf("buf error, op=%s pos=%d", e.Op, e.Pos)
}
