package resolve

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// AuthorityCleaner remove the Authority section from proto.Message and set
// the NSCount field to 0 if any answer is present.
type AuthorityCleaner struct {
	sub Resolver
}

func NewAuthorityCleaner(sub Resolver) *AuthorityCleaner {
	return &AuthorityCleaner{
		sub: sub,
	}
}

func (ac *AuthorityCleaner) Resolve(ctx context.Context, in *proto.InternalMessage) (*proto.InternalMessage, error) {
	out, err := ac.sub.Resolve(ctx, in)
	if err != nil {
		return nil, err
	}

	// todo: clarify when resolver cleanup authority section
	if len(out.Answer) != 0 {
		out.Authority = nil
	}

	return out, nil
}
