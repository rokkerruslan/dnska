package bv_test

import (
	"errors"
	"math"
	"testing"

	tt "github.com/rokkerruslan/dnska/testing"

	. "github.com/rokkerruslan/dnska/pkg/bv"
)

func TestByteView(t *testing.T) {
	// Test basic API functionality of the BV type.

	t.Run("TestNew", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03}
		nb := New(data)

		if nb.Len() != len(data) {
			t.Fatalf("Len() = %v; want %v", nb.Len(), len(data))
		}
	})

	t.Run("TestPos", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03}
		nb := New(data)

		if nb.Pos() != 0 {
			t.Fatalf("Pos() = %v; want 0", nb.Pos())
		}

		_, err := nb.Uint8()
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}

		if nb.Pos() != 1 {
			t.Fatalf("Pos() = %v; want 1", nb.Pos())
		}
	})

	t.Run("TestTake", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03}
		nb := New(data)

		b, err := nb.Uint8()
		if err != nil {
			t.Fatalf("Take() error: %v", err)
		}
		if b != 0x01 {
			t.Fatalf("Take() = %v; want 0x01", b)
		}

		b, err = nb.Uint8()
		if err != nil {
			t.Fatalf("Take() error: %v", err)
		}
		if b != 0x02 {
			t.Fatalf("Take() = %v; want 0x02", b)
		}

		b, err = nb.Uint8()
		if err != nil {
			t.Fatalf("Take() error: %v", err)
		}
		if b != 0x03 {
			t.Fatalf("Take() = %v; want 0x03", b)
		}

		_, err = nb.Uint8()
		if err == nil {
			t.Fatalf("Take() expected error, got nil")
		}
	})

	t.Run("TestAdvance", func(t *testing.T) {
		nb := New([]byte{0x01, 0x02, 0x03})

		if ok := nb.Advance(2); !ok {
			t.Fatalf("Advance(2) = false; want true")
		}

		if nb.Pos() != 2 {
			t.Fatalf("Pos() = %v; want 2", nb.Pos())
		}

		if ok := nb.Advance(2); ok {
			t.Fatalf("Advance(2) = true; want false")
		}

		if nb.Pos() != 2 {
			t.Fatalf("Pos() after failed Advance = %v; want 2", nb.Pos())
		}
	})

	t.Run("TestSeek", func(t *testing.T) {
		nb := New([]byte{0x01, 0x02, 0x03})

		if ok := nb.Seek(1); !ok {
			t.Fatalf("Seek(1) = false; want true")
		}

		if nb.Pos() != 1 {
			t.Fatalf("Pos() = %v; want 1", nb.Pos())
		}

		b, err := nb.Uint8()
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if b != 0x02 {
			t.Fatalf("Get() = %v; want 0x02", b)
		}

		if ok := nb.Seek(9); ok {
			t.Fatalf("Seek(9) = true; want false")
		}

		if nb.Pos() != 2 {
			t.Fatalf("Pos() after failed Seek = %v; want 2", nb.Pos())
		}
	})

	t.Run("TestUnreadBytes", func(t *testing.T) {
		nb := New([]byte{0x01, 0x02, 0x03})

		if ok := nb.Advance(1); !ok {
			t.Fatalf("Advance(1) = false; want true")
		}

		got := nb.UnreadBytes()
		want := []byte{0x02, 0x03}
		if len(got) != len(want) {
			t.Fatalf("len(UnreadBytes()) = %d; want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("UnreadBytes()[%d] = %#x; want %#x", i, got[i], want[i])
			}
		}
	})

	t.Run("TestIndex", func(t *testing.T) {
		nb := New([]byte{0x01, 0x02, 0x03})

		b, err := nb.Index(2)
		if err != nil {
			t.Fatalf("Index() error: %v", err)
		}
		if b != 0x03 {
			t.Fatalf("Index() = %v; want 0x03", b)
		}

		if nb.Pos() != 0 {
			t.Fatalf("Pos() = %v; want 0", nb.Pos())
		}

		_, err = nb.Index(3)
		if err == nil {
			t.Fatalf("Index() expected error, got nil")
		}

		var bufErr *ErrBuf
		if !errors.As(err, &bufErr) {
			t.Fatalf("Index() error = %T; want *ErrBuf", err)
		}
	})

	t.Run("TestRange", func(t *testing.T) {
		nb := New([]byte{0x01, 0x02, 0x03, 0x04})

		got, err := nb.Range(1, 2)
		if err != nil {
			t.Fatalf("Range() error: %v", err)
		}
		want := []byte{0x02, 0x03}
		if len(got) != len(want) {
			t.Fatalf("len(Range()) = %v; want %v", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Range()[%d] = %v; want %v", i, got[i], want[i])
			}
		}

		if nb.Pos() != 0 {
			t.Fatalf("Pos() = %v; want 0", nb.Pos())
		}

		_, err = nb.Range(3, 2)
		if err == nil {
			t.Fatalf("Range() expected error, got nil")
		}
	})

	t.Run("TestTakeUint16", func(t *testing.T) {
		nb := New([]byte{0x12, 0x34, 0x56})

		v, err := nb.Uint16()
		if err != nil {
			t.Fatalf("TakeUint16() error: %v", err)
		}
		if v != 0x1234 {
			t.Fatalf("TakeUint16() = %#x; want %#x", v, 0x1234)
		}
		if nb.Pos() != 2 {
			t.Fatalf("Pos() = %v; want 2", nb.Pos())
		}

		_, err = nb.Uint16()
		if err == nil {
			t.Fatalf("TakeUint16() expected error, got nil")
		}
	})

	t.Run("TestTakeUint32", func(t *testing.T) {
		nb := New([]byte{0x12, 0x34, 0x56, 0x78})

		v, err := nb.Uint32()
		if err != nil {
			t.Fatalf("TakeUint32() error: %v", err)
		}
		if v != 0x12345678 {
			t.Fatalf("TakeUint32() = %#x; want %#x", v, 0x12345678)
		}

		_, err = nb.Uint32()
		if err == nil {
			t.Fatalf("TakeUint32() expected error, got nil")
		}
	})

	t.Run("TestTakeUint64", func(t *testing.T) {
		nb := New([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef})

		v, err := nb.Uint64()
		if err != nil {
			t.Fatalf("Uint64() error: %v", err)
		}
		if v != 0x0123456789abcdef {
			t.Fatalf("Uint64() = %#x; want %#x", v, uint64(0x0123456789abcdef))
		}

		_, err = nb.Uint64()
		if err == nil {
			t.Fatalf("Uint64() expected error, got nil")
		}
	})

	t.Run("TestPutUint8", func(t *testing.T) {
		buf := make([]byte, 1)
		nb := New(buf)

		if err := nb.PutUint8(0xab); err != nil {
			t.Fatalf("PutUint8() error: %v", err)
		}
		if buf[0] != 0xab {
			t.Fatalf("buf[0] = %#x; want %#x", buf[0], 0xab)
		}
		if nb.Pos() != 1 {
			t.Fatalf("Pos() = %v; want 1", nb.Pos())
		}

		if err := nb.PutUint8(0xcd); err == nil {
			t.Fatalf("PutUint8() expected error, got nil")
		}
	})

	t.Run("TestPutUint16", func(t *testing.T) {
		buf := make([]byte, 2)
		nb := New(buf)

		if err := nb.PutUint16(0xabcd); err != nil {
			t.Fatalf("PutUint16() error: %v", err)
		}
		want := []byte{0xab, 0xcd}
		for i := range want {
			if buf[i] != want[i] {
				t.Fatalf("buf[%d] = %#x; want %#x", i, buf[i], want[i])
			}
		}

		if err := nb.PutUint16(0x0102); err == nil {
			t.Fatalf("PutUint16() expected error, got nil")
		}
	})

	t.Run("TestPutUint32", func(t *testing.T) {
		buf := make([]byte, 4)
		nb := New(buf)

		if err := nb.PutUint32(0x12345678); err != nil {
			t.Fatalf("PutUint32() error: %v", err)
		}
		want := []byte{0x12, 0x34, 0x56, 0x78}
		for i := range want {
			if buf[i] != want[i] {
				t.Fatalf("buf[%d] = %#x; want %#x", i, buf[i], want[i])
			}
		}

		if err := nb.PutUint32(0x01020304); err == nil {
			t.Fatalf("PutUint32() expected error, got nil")
		}
	})

	t.Run("TestPutUint64", func(t *testing.T) {
		buf := make([]byte, 8)
		nb := New(buf)

		if err := nb.PutUint64(0x0123456789abcdef); err != nil {
			t.Fatalf("PutUint64() error: %v", err)
		}
		want := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
		for i := range want {
			if buf[i] != want[i] {
				t.Fatalf("buf[%d] = %#x; want %#x", i, buf[i], want[i])
			}
		}

		if err := nb.PutUint64(0x0102030405060708); err == nil {
			t.Fatalf("PutUint64() expected error, got nil")
		}
	})

	t.Run("TestBytes", func(t *testing.T) {
		buf := make([]byte, 4)
		nb := New(buf)

		if err := nb.PutUint16(0xabcd); err != nil {
			t.Fatalf("PutUint16() error: %v", err)
		}

		got := nb.Bytes()
		want := []byte{0xab, 0xcd}
		if len(got) != len(want) {
			t.Fatalf("len(Bytes()) = %v; want %v", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Bytes()[%d] = %#x; want %#x", i, got[i], want[i])
			}
		}
	})

	t.Run("TestErrBufError", func(t *testing.T) {
		nb := New([]byte{0x01})
		_, _ = nb.Uint8()
		_, err := nb.Uint8()
		if err == nil {
			t.Fatalf("Get() expected error, got nil")
		}

		if err.Error() != "buf error, op=get pos=1" {
			t.Fatalf("Error() = %q; want %q", err.Error(), "buf error, op=get pos=1")
		}
	})
}

func TestRange(t *testing.T) {
	data := []byte{1, 2, 3}
	view := New(data)

	var out []byte
	var err error

	out, err = view.Range(0, 0)
	tt.Check(t, err, nil)
	tt.Check(t, len(out), 0)

	out, err = view.Range(0, 1)
	tt.Check(t, err, nil)
	tt.Check(t, len(out), 1)
	tt.Check(t, out[0], uint8(1))

	out, err = view.Range(0, 2)
	tt.Check(t, err, nil)
	tt.Check(t, len(out), 2)
	tt.Check(t, out[0], uint8(1))
	tt.Check(t, out[1], uint8(2))

	out, err = view.Range(0, 3)
	tt.Check(t, err, nil)
	tt.Check(t, len(out), 3)
	tt.Check(t, out[0], uint8(1))
	tt.Check(t, out[1], uint8(2))
	tt.Check(t, out[2], uint8(3))

	out, err = view.Range(0, 4)
	if err == nil {
		t.Error("an error was expected due to overflow, but got nil")
	}
	tt.Check(t, len(out), 0)

	badStart := uint(math.MaxUint64 - 2)
	var length uint = 4

	_, err = view.Range(badStart, length)
	if err == nil {
		t.Error("an error was expected due to overflow, but got nil")
	}

	badStartInt := uint(math.MaxInt64 - 2)
	_, err = view.Range(badStartInt, length)
	if err == nil {
		t.Error("an error was expected due to exceeding MaxInt, but got nil")
	}
}
