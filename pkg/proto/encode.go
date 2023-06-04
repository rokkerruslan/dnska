package proto

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/bv"
)

type Encoder struct {
	bv     *bv.ByteView
	index  *labelsIndex // deprecated
	index2 map[string]uint16
}

func NewEncoder(to []byte) *Encoder {
	return &Encoder{
		bv:    bv.NewByteView(to),
		index: &labelsIndex{nameIndex: map[string]uint{}},
	}
}

// Encode encodes message "m".
func (enc *Encoder) Encode(m Message) ([]byte, error) {
	if err := encodeHeader(enc.bv, m.Header); err != nil {
		return enc.bv.Bytes(), err
	}

	for _, question := range m.Question {
		if err := encodeQuestion(enc.bv, enc.index, question); err != nil {
			return enc.bv.Bytes(), err
		}
	}

	for _, record := range m.Answer {
		if err := encodeRecord(enc.bv, enc.index, record); err != nil {
			return enc.bv.Bytes(), err
		}
	}

	for _, record := range m.Authority {
		if err := encodeRecord(enc.bv, enc.index, record); err != nil {
			return enc.bv.Bytes(), err
		}
	}

	for _, record := range m.Additional {
		if err := encodeRecord(enc.bv, enc.index, record); err != nil {
			return enc.bv.Bytes(), err
		}
	}

	return enc.bv.Bytes(), nil
}

func encodeHeader(buf *bv.ByteView, h Header) error {
	if err := buf.PutUint16(h.ID); err != nil {
		return err
	}

	flags := uint16(0)
	if h.Response {
		flags |= 0x8000
	}
	flags |= uint16(h.Opcode) << 11
	if h.AuthoritativeAnswer {
		flags |= 0x400
	}
	if h.TruncateCation {
		flags |= 0x200
	}
	if h.RecursionDesired {
		flags |= 0x100
	}
	if h.RecursionAvailable {
		flags |= 0x80
	}
	flags |= uint16(h.RCode)

	if err := buf.PutUint8(uint8(flags >> 8)); err != nil {
		return err
	}

	if err := buf.PutUint8(uint8(flags & 0xff)); err != nil {
		return err
	}

	if err := buf.PutUint16(h.QDCount); err != nil {
		return err
	}

	if err := buf.PutUint16(h.ANCount); err != nil {
		return err
	}

	if err := buf.PutUint16(h.NSCount); err != nil {
		return err
	}

	if err := buf.PutUint16(h.ARCount); err != nil {
		return err
	}

	return nil
}

func encodeQuestion(b *bv.ByteView, index *labelsIndex, q Question) error {
	if err := index.EncodeName(b, q.Name); err != nil {
		return err
	}

	if err := b.PutUint16(uint16(q.Type)); err != nil {
		return err
	}

	if err := b.PutUint16(uint16(q.Class)); err != nil {
		return err
	}

	return nil
}

func encodeRecord(b *bv.ByteView, index *labelsIndex, r ResourceRecord) error {
	if err := index.EncodeName(b, r.Name); err != nil {
		return err
	}

	if err := b.PutUint16(uint16(r.Type)); err != nil {
		return err
	}

	if err := b.PutUint16(uint16(r.Class)); err != nil {
		return err
	}

	if err := b.PutUint32(r.TTL); err != nil {
		return err
	}

	b.Seek(b.Pos() + 2)
	//if err := b.PutUint16(r.RDLength); err != nil {
	//	return err
	//}

	start := b.Pos()

	if err := encodeResourceData(b, index, r); err != nil {
		return err
	}

	end := b.Pos()

	// todo: Workaround. We can't put a value of RDLength field because it
	//       can produce malformed message packet. Length must be calculated
	//       based on encodeResourceData function call below.

	b.Seek(start - 2) // seek to the RDLength position.

	if err := b.PutUint16(uint16(end - start)); err != nil {
		return err
	}

	b.Seek(end)

	return nil
}

func encodeResourceData(nb *bv.ByteView, index *labelsIndex, r ResourceRecord) error {
	switch r.Type {
	case QTypeA:
		for _, part := range strings.Split(r.RData, ".") {
			i, _ := strconv.Atoi(part)
			if err := nb.PutUint8(uint8(i)); err != nil {
				return err
			}
		}

		return nil

	case QTypeNS:
		return index.EncodeName(nb, r.RData)

	case QTypeMD:
		return index.EncodeName(nb, r.RData)

	case QTypeMF:
		return index.EncodeName(nb, r.RData)

	case QTypeCName:
		return index.EncodeName(nb, r.RData)

	case QTypeSOA:

	case QTypeMB:
		return index.EncodeName(nb, r.RData)

	case QTypeMG:
		return index.EncodeName(nb, r.RData)

	case QTypeMR:
		return index.EncodeName(nb, r.RData)

	case QTypeNULL:
		for _, b := range []byte(r.RData) {
			if err := nb.PutUint8(b); err != nil {
				return err
			}
		}

		return nil

	case QTypeWKS:
	case QTypePTR:
		return index.EncodeName(nb, r.RData)

	case QTypeHINFO:
		parts := strings.Split(r.RData, "|")

		if err := encodeCharacterString(nb, parts[0]); err != nil {
			return err
		}

		if err := encodeCharacterString(nb, parts[1]); err != nil {
			return err
		}

		return nil

	case QTypeMINFO:
	case QTypeMX:
	case QTypeTXT:

	case QTypeAAAA:
		for _, part := range strings.Split(r.RData, ":") {
			a, b := part[:2], part[2:]
			ai, _ := strconv.ParseUint(a, 16, 8)
			bi, _ := strconv.ParseUint(b, 16, 8)

			if err := nb.PutUint8(uint8(ai)); err != nil {
				return err
			}

			if err := nb.PutUint8(uint8(bi)); err != nil {
				return err
			}
		}

		return nil

	case QTypeAXFR:
	case QTypeMAILB:
	case QTypeMAILA:
	case QTypeALL:
	}

	// Default encoding.

	for _, b := range []byte(r.RData) {
		if err := nb.PutUint8(b); err != nil {
			return err
		}
	}

	return nil
}

func encodeCharacterString(nb *bv.ByteView, s string) error {
	if len(s) > limits.MaxNameSize {
		return fmt.Errorf("the length should be %d or less, got %d", limits.MaxNameSize, len(s))
	}

	// todo: A compression too? The RFC 1035 says - no, the
	//       compression is only for a domain name.

	if err := nb.PutUint8(uint8(len(s))); err != nil {
		return err
	}

	for _, b := range []byte(s) {
		if err := nb.PutUint8(b); err != nil {
			return err
		}
	}

	return nil
}

type Part struct {
	Label  string
	Offset uint
}
