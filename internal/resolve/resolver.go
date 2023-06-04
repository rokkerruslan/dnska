package resolve

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type Resolver interface {
	Resolve(context.Context, proto.Message) (proto.Message, error)
}

type ResolverV2 interface {
	ResolveV2(context.Context, proto.Question) (proto.Message, error)
}
