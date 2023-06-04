package bv

import (
	"fmt"
)

// ByteView ...
//
// Use big-endian (BE) bytes order of digital data.
type ByteView struct {
	data []byte
	pos  uint
}

func NewByteView(data []byte) *ByteView {
	return &ByteView{
		data: data,
		pos:  0,
	}
}

func (nb *ByteView) Pos() uint {
	return nb.pos
}

func (nb *ByteView) Advance(v uint) {
	nb.pos += v
}

func (nb *ByteView) Seek(pos uint) {
	nb.pos = pos
}

func (nb *ByteView) Take() (byte, error) {
	return nb.take()
}

func (nb *ByteView) Index(pos uint) (byte, error) {
	if pos >= uint(len(nb.data)) {
		return 0, &ErrBuf{op: "get", pos: nb.pos}
	}

	return nb.data[pos], nil
}

func (nb *ByteView) TakeRange(start, length uint) ([]byte, error) {
	if start+length > uint(len(nb.data)) {
		return nil, &ErrBuf{op: "range", pos: start + length}
	}

	return nb.data[start : start+length], nil
}

func (nb *ByteView) TakeUint16() (uint16, error) {
	a, err := nb.take()
	if err != nil {
		return 0, err
	}
	b, err := nb.take()
	if err != nil {
		return 0, err
	}

	return uint16(a)<<8 | uint16(b), nil
}

func (nb *ByteView) TakeUint32() (uint32, error) {
	a, err := nb.take()
	if err != nil {
		return 0, err
	}
	b, err := nb.take()
	if err != nil {
		return 0, err
	}
	c, err := nb.take()
	if err != nil {
		return 0, err
	}
	d, err := nb.take()
	if err != nil {
		return 0, err
	}

	return uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d), nil
}

func (nb *ByteView) take() (uint8, error) {
	if nb.pos >= uint(len(nb.data)) {
		return 0, &ErrBuf{op: "take", pos: nb.pos}
	}

	b := nb.data[nb.pos]

	nb.pos++

	return b, nil
}

func (nb *ByteView) put(v uint8) error {
	if nb.pos >= uint(len(nb.data)) {
		return &ErrBuf{
			op:  "put",
			pos: nb.pos,
		}
	}

	nb.data[nb.pos] = v
	nb.pos++

	return nil
}

func (nb *ByteView) PutUint8(b uint8) error {
	if err := nb.put(b); err != nil {
		return err
	}

	return nil
}

func (nb *ByteView) PutUint16(b uint16) error {
	if err := nb.put(uint8(b >> 8)); err != nil {
		return err
	}

	if err := nb.put(uint8(b & 0xff)); err != nil {
		return err
	}

	return nil
}

func (nb *ByteView) PutUint32(b uint32) error {
	if err := nb.put(uint8(b >> 24 & 0xff)); err != nil {
		return err
	}
	if err := nb.put(uint8(b >> 16 & 0xff)); err != nil {
		return err
	}
	if err := nb.put(uint8(b >> 8 & 0xff)); err != nil {
		return err
	}
	if err := nb.put(uint8(b & 0xff)); err != nil {
		return err
	}

	return nil
}

func (nb *ByteView) Bytes() []byte {
	return nb.data[:nb.pos]
}

func (nb *ByteView) Len() int {
	return len(nb.data)
}

type ErrBuf struct {
	op  string
	pos uint
}

func (e *ErrBuf) Error() string {
	return fmt.Sprintf("buf error, op=%s pos=%d", e.op, e.pos)
}
