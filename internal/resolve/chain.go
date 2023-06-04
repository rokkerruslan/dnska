package resolve

import (
	"context"
	"errors"
	"reflect"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

func check(in proto.Message) error {
	if len(in.Question) == 0 {
		return errors.New("question section is empty")
	}

	return nil
}

type chainResolverMode int

const (
	sequenceChainResolverMode chainResolverMode = iota
)

func NewChainResolver(logger zerolog.Logger, list ...Resolver) Resolver {
	return &ChainResolver{
		chain: list,
		mode:  sequenceChainResolverMode,
		l:     logger,
	}
}

type ChainResolver struct {
	chain []Resolver
	mode  chainResolverMode

	l zerolog.Logger
}

func (c *ChainResolver) Resolve(ctx context.Context, in proto.Message) (proto.Message, error) {
	if len(c.chain) == 0 {
		return proto.Message{}, errors.New("chain resolver has zero sub resolvers")
	}

	for _, el := range c.chain {
		out, err := el.Resolve(ctx, in)
		if err != nil {
			c.l.Printf("resolver=%s returns error=%s", reflect.ValueOf(el).Type(), err)
			continue
		}

		return out, nil
	}

	return proto.Message{}, errors.New("all resolvers return error")
}
