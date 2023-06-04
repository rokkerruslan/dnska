package resolve

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type answer struct {
	records []proto.ResourceRecord
}

func NewStaticResolver(_ zerolog.Logger) *StaticResolver {
	return &StaticResolver{
		map[proto.Question]answer{
			proto.Question{
				Name:  "lolkek",
				Type:  proto.QTypeA,
				Class: proto.ClassIN,
			}: {records: []proto.ResourceRecord{{
				Name:     "lolkek",
				Type:     proto.QTypeA,
				Class:    proto.ClassIN,
				TTL:      10,
				RDLength: 9,
				RData:    "127.0.0.1",
			}}},
		},
	}
}

type StaticResolver struct {
	m map[proto.Question]answer
}

func (s *StaticResolver) Resolve(_ context.Context, in proto.Message) (proto.Message, error) {
	if err := check(in); err != nil {
		return proto.Message{}, err
	}

	if len(in.Question) != 1 {
		return proto.Message{}, fmt.Errorf("static resolver can not handle more than one query, got=%d", len(in.Question))
	}

	question := in.Question[0]

	ans, ok := s.m[question]
	if !ok {
		return proto.Message{}, fmt.Errorf("static resolver do not contains answer on q=%v", question)
	}

	out := proto.Message{
		Header: proto.Header{
			ID:                  in.Header.ID,
			Response:            true,
			Opcode:              in.Header.Opcode,
			AuthoritativeAnswer: false,
			TruncateCation:      false,
			RecursionDesired:    false,
			RecursionAvailable:  true,
			Z:                   0,
			RCode:               proto.RCodeNoErrorCondition,
			QDCount:             0,
			ANCount:             1,
			NSCount:             0,
			ARCount:             0,
		},
		Answer: ans.records,
	}

	return out, nil
}
