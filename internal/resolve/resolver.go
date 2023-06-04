package resolve

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// Resolver is an interface for resolving DNS queries.
type Resolver interface {
	Resolve(context.Context, *proto.InternalMessage) (*proto.InternalMessage, error)
}
