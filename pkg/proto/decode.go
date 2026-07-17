package proto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rokkerruslan/dnska/pkg/bv"
)

// Decoder is a struct that provides functionality to decode DNS messages from byte slices.
type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (dec *Decoder) Decode(in []byte) (*Message, error) {
	buf := bv.New(in)

	var err error
	var header Header

	header.ID, err = buf.Uint16()
	if err != nil {
		return nil, err
	}

	flagsH, err := buf.Uint8()
	if err != nil {
		return nil, err
	}

	flagsL, err := buf.Uint8()
	if err != nil {
		return nil, err
	}

	header.Response = (flagsH & 0b10000000) != 0
	header.Opcode = Opcode((flagsH & 0b01111000) >> 3)
	header.AuthoritativeAnswer = (flagsH & 0b00000100) != 0
	header.Truncated = (flagsH & 0b00000010) != 0
	header.RecursionDesired = (flagsH & 0b00000001) != 0
	header.RecursionAvailable = (flagsL & 0b10000000) != 0
	header.RCode = RCode(flagsL & 0b00001111)

	header.QDCount, err = buf.Uint16()
	if err != nil {
		return nil, err
	}

	header.ANCount, err = buf.Uint16()
	if err != nil {
		return nil, err
	}

	header.NSCount, err = buf.Uint16()
	if err != nil {
		return nil, err
	}

	header.ARCount, err = buf.Uint16()
	if err != nil {
		return nil, err
	}

	questions, err := parseQuestion(buf, int(header.QDCount))
	if err != nil {
		return nil, err
	}

	answers, err := decodeResourceRecords(buf, int(header.ANCount))
	if err != nil {
		return nil, err
	}

	authorities, err := decodeResourceRecords(buf, int(header.NSCount))
	if err != nil {
		return nil, err
	}

	additional, err := decodeResourceRecords(buf, int(header.ARCount))
	if err != nil {
		return nil, err
	}

	return &Message{
		Header:     header,
		Question:   questions,
		Answer:     answers,
		Authority:  authorities,
		Additional: additional,
	}, nil
}

func parseQuestion(nb *bv.BV, n int) ([]Question, error) {
	out := make([]Question, 0, n)

	for i := 0; i < n; i++ {
		qName, err := decodeName(nb)
		if err != nil {
			return nil, err
		}

		qType, err := nb.Uint16()
		if err != nil {
			return nil, err
		}

		qClass, err := nb.Uint16()
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

func decodeResourceRecords(nb *bv.BV, n int) ([]ResourceRecord, error) {
	out := make([]ResourceRecord, 0, n)

	for range n {
		record, err := decodeResourceRecord(nb)
		if err != nil {
			return out, err
		}

		out = append(out, record)
	}

	return out, nil
}

func decodeResourceRecord(nb *bv.BV) (ResourceRecord, error) {
	name, err := decodeName(nb)
	if err != nil {
		return ResourceRecord{}, err
	}

	queryType, err := nb.Uint16()
	if err != nil {
		return ResourceRecord{}, err
	}

	class, err := nb.Uint16()
	if err != nil {
		return ResourceRecord{}, err
	}

	ttl, err := nb.Uint32()
	if err != nil {
		return ResourceRecord{}, err
	}

	rdLength, err := nb.Uint16()
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

func decodeResourceData(nb *bv.BV, queryType QType, length uint16) (string, error) {
	startPos := nb.Pos()
	var result string
	var err error

	switch queryType {
	case QTypeA:
		var addr uint32
		addr, err = nb.Uint32()
		if err == nil {
			result = fmt.Sprintf(
				"%d.%d.%d.%d",
				uint8(addr>>24),
				uint8(addr>>16),
				uint8(addr>>8),
				uint8(addr>>0),
			)
		}

	case QTypeNS, QTypeMD, QTypeMF, QTypeCName, QTypeMB, QTypeMG, QTypeMR, QTypePTR:
		result, err = decodeName(nb)

	case QTypeSOA:
		var mname, rname string
		mname, err = decodeName(nb)
		if err == nil {
			rname, err = decodeName(nb)
		}

		var serial, refresh, retry, expire, minimum uint32
		if err == nil {
			serial, err = nb.Uint32()
		}
		if err == nil {
			refresh, err = nb.Uint32()
		}
		if err == nil {
			retry, err = nb.Uint32()
		}
		if err == nil {
			expire, err = nb.Uint32()
		}
		if err == nil {
			minimum, err = nb.Uint32()
		}

		if err == nil {
			result = fmt.Sprintf("%s %s %d %d %d %d %d", mname, rname, serial, refresh, retry, expire, minimum)
		}

	case QTypeNULL:
		var bts []byte
		bts, err = nb.Range(nb.Pos(), uint(length))
		if err == nil {
			result = string(bts)
		}

	case QTypeHINFO:
		var cpu, os string
		cpu, err = decodeStr(nb)
		if err == nil {
			os, err = decodeStr(nb)
		}
		if err == nil {
			result = cpu + "|" + os
		}

	case QTypeAAAA:
		var bts []byte
		bts, err = nb.Range(nb.Pos(), 16)
		if err == nil {
			var hx strings.Builder
			for i, b := range bts {
				if i%2 == 0 && i != 0 {
					hx.WriteString(":")
				}
				hx.WriteString(fmt.Sprintf("%02x", b))
			}
			result = hx.String()
		}

	default:
		var bts []byte
		bts, err = nb.Range(nb.Pos(), uint(length))
		if err == nil {
			result = string(bts)
		}
	}

	if err != nil {
		return "", err
	}

	nb.Seek(startPos + uint(length))

	return result, nil
}

// decodeName ...
//
// <domain-name> is a domain name represented as a series of labels, and
// terminated by a label with zero length.
func decodeName(nb *bv.BV) (string, error) {

	const maxJumps = 5
	const mask = 0b11000000
	const delim = "."

	pos := nb.Pos()
	jumped := false
	jumpsPerformed := 0

	var out strings.Builder

	for {
		length, err := nb.Index(pos)
		if err != nil {
			return "", err
		}

		// If the two most significant bits of the length byte are set, it indicates
		// a pointer to another location in the message.
		if length&mask == mask {
			b2, err := nb.Index(pos + 1)
			if err != nil {
				return "", err
			}

			if !jumped {
				nb.Seek(pos + 2)
				jumped = true
			}

			offset := (uint16(length&^mask) << 8) | uint16(b2)

			// offset must be at least 12 bytes into the message to avoid
			// pointing into the header. It must also be less than the
			// total length of the message to avoid pointing outside the message.
			if offset < 12 || offset >= uint16(nb.Pos()) {
				return "", errors.New("invalid offset in pointer")
			}

			pos = uint(offset)
			jumpsPerformed++

			// Detect and prevent infinite loops in case of malformed messages with circular pointers.
			if jumpsPerformed > maxJumps {
				return "", errors.New("jumps limit reached")
			}

			continue
		}

		if length > 63 {
			return "", errors.New("label length exceeds 63 bytes")
		}

		// Skip the length byte
		pos++

		// If the length is zero, it indicates the end of the domain name.
		if length == 0 {
			if !jumped {
				nb.Seek(pos)
			}
			break
		}

		out.WriteString(delim)
		part, err := nb.Range(pos, uint(length))
		if err != nil {
			return "", err
		}

		out.WriteString(string(part))
		pos += uint(length)
	}

	return out.String(), nil
}

func decodeStr(nb *bv.BV) (string, error) {
	length, err := nb.Uint8()
	if err != nil {
		return "", err
	}

	buf, err := nb.Range(nb.Pos(), uint(length))
	if err != nil {
		return "", err
	}

	nb.Advance(uint(length))

	return string(buf), nil
}
