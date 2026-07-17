package proto

import (
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

// EncodeName encodes domain name in buffer using DNS compression (RFC 1035).
func (li *labelsIndex) EncodeName(b *bv.BV, s string) error {
	if len(s) > limits.MaxNameSize {
		return fmt.Errorf("the name length should be %d or less, got %d", limits.MaxNameSize, len(s))
	}

	s = strings.Trim(s, ".")

	// root zone
	if s == "" {
		return b.PutUint8(0)
	}

	if labels, offset, exist := li.getName(s); exist {
		startPos := b.Pos()

		for _, label := range labels {
			if len(label) > limits.MaxLabelSize {
				return fmt.Errorf("label %q exceeds limit %d", label, limits.MaxLabelSize)
			}
			if err := b.PutUint8(uint8(len(label))); err != nil {
				return err
			}
			for i := 0; i < len(label); i++ {
				if err := b.PutUint8(label[i]); err != nil {
					return err
				}
			}
		}

		if offset > 0x3fff {
			return fmt.Errorf("compression offset %d exceeds 14-bit limit", offset)
		}

		ptr := uint16(offset) | 0xc000
		if err := b.PutUint8(uint8(ptr >> 8)); err != nil {
			return err
		}
		if err := b.PutUint8(uint8(ptr)); err != nil {
			return err
		}

		li.putName(s, startPos)
		return nil
	}

	rawLabels := strings.Split(s, ".")
	type Part struct {
		Label  string
		Offset uint
	}
	parts := make([]Part, 0, len(rawLabels))

	for _, label := range rawLabels {
		if len(label) == 0 {
			return fmt.Errorf("invalid domain name %q: empty label", s)
		}
		if len(label) > limits.MaxLabelSize {
			return fmt.Errorf("label %q too big", label)
		}

		parts = append(parts, Part{
			Label:  label,
			Offset: b.Pos(),
		})

		if err := b.PutUint8(uint8(len(label))); err != nil {
			return err
		}
		for i := 0; i < len(label); i++ {
			if err := b.PutUint8(label[i]); err != nil {
				return err
			}
		}
	}

	if err := b.PutUint8(0); err != nil {
		return err
	}

	suffix := ""
	for i := len(parts) - 1; i >= 0; i-- {
		if suffix == "" {
			suffix = parts[i].Label
		} else {
			suffix = parts[i].Label + "." + suffix
		}
		li.putName(suffix, parts[i].Offset)
	}

	return nil
}

func (li *labelsIndex) DecodeName(b *bv.BV) (string, error) {
	return "", nil
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
