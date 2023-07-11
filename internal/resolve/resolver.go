package resolve

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type Resolver interface {
	Resolve(context.Context, *proto.InternalMessage) (*proto.InternalMessage, error)
}
