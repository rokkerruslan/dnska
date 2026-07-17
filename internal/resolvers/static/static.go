package static

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type Opts struct {
	L *slog.Logger
}

func NewStaticResolver(opts Opts) *StaticResolver {
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
		return nil, errors.New("no answer for question")
	}

	out := proto.InternalMessage{
		Question: question,
		Answer:   ans.records,
	}

	return &out, nil
}

type answer struct {
	records []proto.ResourceRecord
}
