package proto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/bv"
)

// RFC 1035 4.1.4. Message compression
//
// In order to reduce the size of messages, the domain system utilizes a
// compression scheme which eliminates the repetition of domain names in a
// message. In this scheme, an entire domain name or a list of labels at
// the end of a domain name is replaced with a pointer to a prior occurance
// of the same name.
//
// The pointer takes the form of a two octet sequence:
//
//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    | 1  1|                OFFSET                   |
//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// The first two bits are ones. This allows a pointer to be distinguished
// from a label, since the label must begin with two zero bits because
// labels are restricted to 63 octets or less. (The 10 and 01 combinations
// are reserved for future use.) The OFFSET field specifies an offset from
// the start of the message (i.e., the first octet of the ID field in the
// domain header). A zero offset specifies the first byte of the ID field,
// etc.
//
// The compression scheme allows a domain name in a message to be
// represented as either:
//
//   - a sequence of labels ending in a zero octet
//   - a pointer
//   - a sequence of labels ending with a pointer
//
// Pointers can only be used for occurances of a domain name where the
// format is not class specific. If this were not the case, a name server
// or resolver would be required to know the format of all RRs it handled.
// As yet, there are no such cases, but they may occur in future RDATA
// formats.
//
// If a domain name is contained in a part of the message subject to a
// length field (such as the RDATA section of an RR), and compression is
// used, the length of the compressed name is used in the length
// calculation, rather than the length of the expanded name.
//
// Programs are free to avoid using pointers in messages they generate,
// although this will reduce datagram capacity, and may cause truncation.
// However all programs are required to understand arriving messages that
// contain pointers.
//
// For example, a datagram might need to use the domain names F.ISI.ARPA,
// FOO.F.ISI.ARPA, ARPA, and the root. Ignoring the other fields of the
// message, these domain names might be represented as:
//
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    20 |           1           |           F           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    22 |           3           |           I           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    24 |           S           |           I           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    26 |           4           |           A           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    28 |           R           |           P           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    30 |           A           |           0           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    40 |           3           |           F           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    42 |           O           |           O           |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    44 | 1  1|                20                       |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    64 | 1  1|                26                       |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//    92 |           0           |                       |
//       +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// The domain name for F.ISI.ARPA is shown at offset 20. The domain name
// FOO.F.ISI.ARPA is shown at offset 40; this definition uses a pointer to
// concatenate a label for FOO to the previously defined F.ISI.ARPA. The
// domain name ARPA is defined at offset 64 using a pointer to the ARPA
// component of the name F.ISI.ARPA at 20; note that this pointer relies on
// ARPA being the last label in the string at 20. The root domain name is
// defined by a single octet of zeros at 92; the root domain name has no
// labels.

type labelsIndex struct {
	nameIndex map[string]uint
}

// EncodeName encodes domain name in buffer.
//
// Domain names in messages are expressed in terms of a sequence of labels.
// Each label is represented as a one octet length field followed by that
// number of octets. Since every domain name ends with the null label of
// the root, a domain name is terminated by a length byte of zero.  The
// high order two bits of every length octet must be zero, and the
// remaining six bits of the length field limit the label to 63 octets or
// less.
func (li *labelsIndex) EncodeName(b *bv.ByteView, s string) error {
	if len(s) > limits.MaxNameSize {
		return fmt.Errorf("the name length should be %d or less, got %d", limits.MaxNameSize, len(s))
	}

	s = strings.TrimSuffix(s, ".")

	if labels, offset, exist := li.getName(s); exist {
		if len(labels) != 0 { // todo: This is incorrect. Make generic algorithm for labels index.
			li.putName(s, b.Pos())
		}

		for _, label := range labels {
			if err := b.PutUint8(uint8(len(label))); err != nil {
				return err
			}

			for _, ch := range []byte(label) {
				if err := b.PutUint8(ch); err != nil {
					return err
				}
			}
		}

		offset := uint16(offset)

		if err := b.PutUint8(0xc0 | uint8(offset>>8)); err != nil {
			return err
		}

		if err := b.PutUint8(uint8(offset)); err != nil { // uint8(index & 0xff)
			return err
		}

		return nil
	}

	var parts []Part
	for _, label := range strings.Split(s, ".") {
		if len(label) == 0 {
			// todo: error because label is empty?
			break
		}

		labelLength := uint8(len(label))

		if labelLength > limits.MaxLabelSize {
			return fmt.Errorf("label %q too big", label)
		}

		parts = append(parts, Part{
			Label:  label,
			Offset: b.Pos(),
		})

		if err := b.PutUint8(labelLength); err != nil {
			return err
		}

		for _, ch := range []byte(label) {
			if err := b.PutUint8(ch); err != nil {
				return err
			}
		}
	}

	if err := b.PutUint8(0); err != nil {
		return err
	}

	// Build index.

	for i := len(parts) - 1; i >= 0; i-- {
		elements := parts[i:]

		buf := strings.Builder{}
		for _, el := range elements {
			buf.WriteString(el.Label)
			buf.WriteByte('.')
		}

		li.putName(strings.TrimSuffix(buf.String(), "."), elements[0].Offset)
	}

	return nil
}

func (li *labelsIndex) DecodeName(b *bv.ByteView) (string, error) {
	pos := b.Pos()

	jumped := false
	maxJumps := 5
	jumpsPerformed := 0

	delim := ""
	out := ""

	for {
		if jumpsPerformed > maxJumps {
			return "", errors.New("jumps limit reached")
		}

		length, err := b.Index(pos)
		if err != nil {
			return "", err
		}

		if length&0xc0 == 0xc0 {
			if !jumped {
				b.Seek(pos + 2)
			}

			b2, err := b.Index(pos + 1)
			if err != nil {
				return "", err
			}

			offset := (uint16(length^0xc0) << 8) | uint16(b2)
			pos = uint(offset)

			jumped = true
			jumpsPerformed++

			continue
		} else {
			pos++

			if length == 0 {
				break
			}

			out += delim
			part, err := b.TakeRange(pos, uint(length))
			if err != nil {
				return "", err
			}

			out += string(part)

			delim = "."
			pos += uint(length)
		}
	}

	if !jumped {
		b.Seek(pos)
	}

	return out, nil
}

func (li *labelsIndex) getName(name string) ([]string, uint, bool) {
	labels := strings.Split(name, ".")

	for i := 0; i < len(labels); i++ {
		n := strings.Join(labels[i:], ".")

		offset, ok := li.nameIndex[n]
		if ok {
			return labels[:i], offset, ok
		}
	}

	return nil, 0, false
}

func (li *labelsIndex) putName(name string, index uint) {
	if li.nameIndex == nil {
		li.nameIndex = map[string]uint{}
	}

	if _, ok := li.nameIndex[name]; ok {
		return
	}

	li.nameIndex[name] = index
}
