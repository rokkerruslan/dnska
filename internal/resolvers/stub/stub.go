package stub

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type StubResolver struct {
}

type StubResolverOpts struct {
}

func NewStubResolver(opts StubResolverOpts) *StubResolver {
	return &StubResolver{}
}

func (r *StubResolver) Resolve(ctx context.Context, in *proto.InternalMessage) (*proto.InternalMessage, error) {
	return in, nil
}
