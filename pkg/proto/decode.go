package proto

import (
	"errors"
	"fmt"

	"github.com/rokkerruslan/dnska/pkg/bv"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (dec *Decoder) Decode(in []byte) (Message, error) {
	buf := bv.NewByteView(in)

	var err error
	var header Header

	header.ID, err = buf.TakeUint16()
	if err != nil {
		return Message{}, err
	}

	flagsH, err := buf.Take()
	if err != nil {
		return Message{}, err
	}

	flagsL, err := buf.Take()
	if err != nil {
		return Message{}, err
	}

	header.Response = (flagsH & 0b10000000) != 0
	header.Opcode = Opcode((flagsH & 0b01111000) >> 3)
	header.AuthoritativeAnswer = (flagsH & 0b00000100) != 0
	header.TruncateCation = (flagsH & 0b00000010) != 0
	header.RecursionDesired = (flagsH & 0b00000001) != 0
	header.RecursionAvailable = (flagsL & 0b10000000) != 0
	header.RCode = RCode(flagsL & 0b00001111)

	header.QDCount, err = buf.TakeUint16()
	if err != nil {
		return Message{}, err
	}

	header.ANCount, err = buf.TakeUint16()
	if err != nil {
		return Message{}, err
	}

	header.NSCount, err = buf.TakeUint16()
	if err != nil {
		return Message{}, err
	}

	header.ARCount, err = buf.TakeUint16()
	if err != nil {
		return Message{}, err
	}

	questions, err := parseQuestion(buf, int(header.QDCount))
	if err != nil {
		return Message{}, err
	}

	answers, err := decodeResourceRecords(buf, int(header.ANCount))
	if err != nil {
		return Message{}, err
	}

	authorities, err := decodeResourceRecords(buf, int(header.NSCount))
	if err != nil {
		return Message{}, err
	}

	additional, err := decodeResourceRecords(buf, int(header.ARCount))
	if err != nil {
		return Message{}, err
	}

	return Message{
		Header:     header,
		Question:   questions,
		Answer:     answers,
		Authority:  authorities,
		Additional: additional,
	}, nil
}

func parseQuestion(nb *bv.ByteView, n int) ([]Question, error) {
	out := make([]Question, 0, n)

	for i := 0; i < n; i++ {
		qName, err := decodeName(nb)
		if err != nil {
			return nil, err
		}

		qType, err := nb.TakeUint16()
		if err != nil {
			return nil, err
		}

		qClass, err := nb.TakeUint16()
		if err != nil {
			return nil, err
		}

		out = append(out, Question{
			Name:  qName,
			Type:  QType(qType),
			Class: QClass(qClass),
		})
	}

	return out, nil
}

func decodeResourceRecords(nb *bv.ByteView, n int) ([]ResourceRecord, error) {
	out := make([]ResourceRecord, 0, n)

	for i := 0; i < n; i++ {
		record, err := decodeResourceRecord(nb)
		if err != nil {
			return out, err
		}

		out = append(out, record)
	}

	return out, nil
}

func decodeResourceRecord(nb *bv.ByteView) (ResourceRecord, error) {
	name, err := decodeName(nb)
	if err != nil {
		return ResourceRecord{}, err
	}

	queryType, err := nb.TakeUint16()
	if err != nil {
		return ResourceRecord{}, err
	}

	class, err := nb.TakeUint16()
	if err != nil {
		return ResourceRecord{}, err
	}

	ttl, err := nb.TakeUint32()
	if err != nil {
		return ResourceRecord{}, err
	}

	rdLength, err := nb.TakeUint16()
	if err != nil {
		return ResourceRecord{}, err
	}

	rData, err := decodeResourceData(nb, QType(queryType), rdLength)
	if err != nil {
		return ResourceRecord{}, err
	}

	return ResourceRecord{
		Name:     name,
		Type:     QType(queryType),
		Class:    QClass(class),
		TTL:      ttl,
		RDLength: rdLength,
		RData:    rData,
	}, nil
}

func decodeResourceData(nb *bv.ByteView, queryType QType, length uint16) (string, error) {
	switch queryType {
	case QTypeA:
		addr, err := nb.TakeUint32()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(
			"%d.%d.%d.%d",
			uint8(addr>>24),
			uint8(addr>>16),
			uint8(addr>>8),
			uint8(addr>>0),
		), nil

	case QTypeNS:
		// A domain name which specifies a host which should be
		// authoritative for the specified class and domain.

		// NS records cause both the usual additional section processing to locate
		// a type A record, and, when used in a referral, a special search of the
		// zone in which they reside for glue information.

		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}

		return name, nil

	case QTypeMD:
		// A <domain-name> which specifies a host which has a mail
		// agent for the domain which should be able to deliver
		// mail for the domain.

		// MD records cause additional section processing which looks
		// up an A type record corresponding to MADNAME.

		// MD is obsolete.

		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}

		return name, nil

	case QTypeMF:
		// A <domain-name> which specifies a host which has a mail
		// agent for the domain which will accept mail for
		// forwarding to the domain.

		// MF is obsolete.

		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}

		return name, nil

	case QTypeCName:
		// A domain name which specifies the canonical or primary
		// name for the owner. The owner name is an alias.
		cname, err := decodeName(nb)
		if err != nil {
			return "", err
		}
		return cname, nil

	case QTypeSOA:
		buf, err := nb.TakeRange(nb.Pos(), uint(length))
		if err != nil {
			return "", err
		}

		return string(buf), nil

	case QTypeMB:
		// A <domain-name> which specifies a host which has the
		// specified mailbox.

		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}
		return name, nil

	case QTypeMG:
		// A <domain-name> which specifies a mailbox which is a
		// member of the mail group specified by the domain name.
		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}
		return name, nil

	case QTypeMR:
		// A <domain-name> which specifies a mailbox which is the
		// proper rename of the specified mailbox.
		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}
		return name, nil

	case QTypeNULL:
		// Anything at all may be in the RDATA field so long as it
		// is 65535 octets or less.

		// NULL records cause no additional section processing.  NULL RRs are not
		// allowed in master files.  NULLs are used as placeholders in some
		// experimental extensions of the DNS.

		bts, err := nb.TakeRange(nb.Pos(), uint(length))
		if err != nil {
			return "", err
		}

		return string(bts), err

	case QTypeWKS:

	case QTypePTR:
		// A domain name which points to some location in the
		// domain name space.
		name, err := decodeName(nb)
		if err != nil {
			return "", err
		}

		return name, nil

	case QTypeHINFO:
		// CPU A <character-string> which specifies the CPU type.
		// OS A <character-string> which specifies the operating
		// system type.

		// Standard values for CPU and OS can be found in [RFC-1010].
		//
		// HINFO records are used to acquire general information
		// about a host. The main use is for protocols such as FTP
		// that can use special procedures when talking between
		// machines or operating systems of the same type.

		cpu, err := decodeCharacterString(nb)
		if err != nil {
			return "", err
		}

		os, err := decodeCharacterString(nb)
		if err != nil {
			return "", err
		}

		// todo: must ensure that "|" does not occur at the cpu and the os parts.
		return cpu + "|" + os, nil

	case QTypeMINFO:
	case QTypeMX:
	case QTypeTXT:

	case QTypeAAAA:
		// 128 bit IPv6 address is encoded in the data portion of an AAAA
		// resource record in network byte order (high-order byte first).

		bts, err := nb.TakeRange(nb.Pos(), 16)
		if err != nil {
			return "", err
		}

		hx := ""
		for i, b := range bts {
			if i%2 == 0 && i != 0 {
				hx += ":"
			}
			hx += fmt.Sprintf("%02x", b)
		}
		nb.Advance(16)

		return hx, nil

	case QTypeAXFR:
	case QTypeMAILB:
	case QTypeMAILA:
	case QTypeALL:
		// todo: how to handle a request for all records?
	}

	buf, err := nb.TakeRange(nb.Pos(), uint(length))
	if err != nil {
		return "", err
	}
	nb.Advance(uint(length))

	return string(buf), nil
}

// decodeName ...
//
// <domain-name> is a domain name represented as a series of labels, and
// terminated by a label with zero length.
func decodeName(nb *bv.ByteView) (string, error) {
	pos := nb.Pos()

	jumped := false
	maxJumps := 5
	jumpsPerformed := 0

	delim := ""
	out := ""

	for {
		if jumpsPerformed > maxJumps {
			return "", errors.New("jumps limit reached")
		}

		length, err := nb.Index(pos)
		if err != nil {
			return "", err
		}

		if length&0xc0 == 0xc0 {
			if !jumped {
				nb.Seek(pos + 2)
			}

			b2, err := nb.Index(pos + 1)
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
			part, err := nb.TakeRange(pos, uint(length))
			if err != nil {
				return "", err
			}

			out += string(part)

			delim = "."
			pos += uint(length)
		}
	}

	if !jumped {
		nb.Seek(pos)
	}

	return out, nil
}

// decodeCharacterString ...
//
// <character-string> is a single length octet followed by
// that number of characters.  <character-string> is treated
// as binary information, and can be up to 256 characters in
// length (including the length octet).
func decodeCharacterString(nb *bv.ByteView) (string, error) {
	length, err := nb.Take()
	if err != nil {
		return "", err
	}

	buf, err := nb.TakeRange(nb.Pos(), uint(length))
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
