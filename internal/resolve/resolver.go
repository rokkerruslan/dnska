package resolve

import (
	"context"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// HResolver is an interface for resolving DNS queries.
type HResolver interface {
	Resolve(context.Context, *proto.InternalMessage) (*proto.InternalMessage, error)
}

type Response struct {
	Answer     []proto.ResourceRecord
	Authority  []proto.ResourceRecord
	Additional []proto.ResourceRecord
}

// LResolver is an high-level interface for resolving DNS queries
// with a single question.
type LResolver interface {
	Resolve(context.Context, proto.Question) (Response, error)
}
