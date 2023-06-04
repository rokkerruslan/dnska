package resolve

import (
	"context"
	"errors"
	"log/slog"
	"reflect"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type chainResolverMode int

const (
	sequenceChainResolverMode chainResolverMode = iota
)

var errNoReport = errors.New("no report")

func NewChainResolver(l *slog.Logger, list ...Resolver) Resolver {
	return &ChainResolver{
		chain: list,
		mode:  sequenceChainResolverMode,
		l:     l,
	}
}

type ChainResolver struct {
	chain []Resolver
	mode  chainResolverMode

	l *slog.Logger
}

func (c *ChainResolver) Resolve(ctx context.Context, in *proto.InternalMessage) (*proto.InternalMessage, error) {
	if len(c.chain) == 0 {
		return nil, errors.New("chain resolver has zero sub resolvers")
	}

	for _, el := range c.chain {
		out, err := el.Resolve(ctx, in)
		if err != nil {
			if !errors.Is(err, errNoReport) {
				c.l.Error("failed attempt", "resolver", reflect.ValueOf(el).Type(), "error", err)
			}
			continue
		}

		return out, nil
	}

	return nil, errors.New("all resolvers return error")
}
