package resolve

import (
	"context"
	"log/slog"
	"math"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type answer struct {
	records []proto.ResourceRecord
}

func NewStaticResolver(_ *slog.Logger) *StaticResolver {
	return &StaticResolver{
		map[proto.Question]answer{
			{
				Name:  "ya.ru",
				Type:  proto.QTypeA,
				Class: proto.ClassIN,
			}: {records: []proto.ResourceRecord{{
				Name:     "lolkek",
				Type:     proto.QTypeA,
				Class:    proto.ClassIN,
				TTL:      math.MaxUint32,
				RDLength: 4,
				RData:    "127.0.0.1",
			}}},
		},
	}
}

type StaticResolver struct {
	m map[proto.Question]answer
}

func (s *StaticResolver) Resolve(_ context.Context, in *proto.InternalMessage) (*proto.InternalMessage, error) {
	question := in.Question

	ans, ok := s.m[question]
	if !ok {
		return nil, errNoReport
	}

	out := proto.InternalMessage{
		Question: question,
		Answer:   ans.records,
	}

	return &out, nil
}
