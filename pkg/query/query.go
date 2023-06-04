package query

import (
	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewTemplate() proto.Message {
	return proto.Message{
		Header: proto.Header{
			ID:                  0,
			Response:            false,
			Opcode:              0,
			AuthoritativeAnswer: false,
			TruncateCation:      false,
			RecursionDesired:    false,
			RecursionAvailable:  false,
			Z:                   0,
			RCode:               0,
			QDCount:             0,
			ANCount:             0,
			NSCount:             0,
			ARCount:             0,
		},
		Question:   nil,
		Answer:     nil,
		Authority:  nil,
		Additional: nil,
	}
}

func AddQuestion(in proto.Message, name string, qtype proto.QType, qclass proto.QClass) proto.Message {
	in.Question = append(in.Question, proto.Question{
		Name:  name,
		Type:  qtype,
		Class: qclass,
	})

	in.Header.QDCount++

	return in
}

func NewQueryA(name string) proto.Message {
	return proto.Message{
		Header: proto.Header{
			ID:                  1,
			Response:            false,
			Opcode:              0,
			AuthoritativeAnswer: false,
			TruncateCation:      false,
			RecursionDesired:    false,
			RecursionAvailable:  false,
			Z:                   0,
			RCode:               0,
			QDCount:             211,
			ANCount:             0,
			NSCount:             0,
			ARCount:             0,
		},
		Question: []proto.Question{
			{
				Name:  name,
				Type:  proto.QTypeA,
				Class: proto.ClassIN,
			},
		},
		Answer:     nil,
		Authority:  nil,
		Additional: nil,
	}
}
