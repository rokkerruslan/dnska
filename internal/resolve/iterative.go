package resolve

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewIterativeResolver(l zerolog.Logger) *IterativeResolver {
	return &IterativeResolver{
		mu:                   sync.Mutex{},
		n:                    0,
		dumpMalformedPackets: true,
		l:                    l,
	}
}

type IterativeResolver struct {
	mu sync.Mutex
	n  uint16

	dumpMalformedPackets bool

	l zerolog.Logger
}

func (fr *IterativeResolver) ResolveV2(ctx context.Context, q proto.Question) (proto.Message, error) {
	opts := LookupOpts{
		Name:              q.Name,
		Type:              q.Type,
		Class:             q.Class,
		ID:                fr.nextID(),
		DumpUnknownPacket: fr.dumpMalformedPackets,
		L:                 fr.l,
	}

	return cycle(ctx, opts)
}

// deprecated
func (fr *IterativeResolver) Resolve(ctx context.Context, in proto.Message) (proto.Message, error) {
	out := proto.Message{
		Header: proto.Header{
			ID:                 in.Header.ID,
			Response:           true,
			RecursionAvailable: true,
		},
	}

	if !in.Header.RecursionDesired {
		fr.l.Print("failed to lookup non-recursion query")
		out.Header.RCode = proto.RCodeServerFailure

		return out, nil
	}

	switch len(in.Question) {
	case 1:
		break
	case 0:
		out.Header.RCode = proto.RCodeFormatError

		return out, nil

	default:
		// todo: Add support for multi-questions query?
		fr.l.Printf("failed to lookup multi-questions query :: length=%d", len(in.Question))

		out.Header.RCode = proto.RCodeServerFailure

		return out, nil
	}

	question := in.Question[0]

	opts := LookupOpts{
		Name:              question.Name,
		Type:              question.Type,
		Class:             question.Class,
		ID:                fr.nextID(),
		DumpUnknownPacket: fr.dumpMalformedPackets,
		L:                 fr.l,
	}

	if found, err := cycle(ctx, opts); err == nil {
		out.Header.QDCount = 1
		out.Header.ANCount = uint16(len(found.Answer))
		out.Header.NSCount = uint16(len(found.Authority))
		out.Header.ARCount = uint16(len(found.Additional))
		out.Header.RCode = found.Header.RCode

		out.Question = found.Question
		out.Answer = found.Answer
		out.Authority = found.Authority
		out.Additional = found.Additional

	} else {
		fr.l.Printf("failed to lookup :: name=%s, type=%v, error=%v", question.Name, question.Type, err)

		out.Header.RCode = proto.RCodeServerFailure
	}

	return out, nil
}

func (fr *IterativeResolver) nextID() uint16 {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	fr.n++

	return fr.n
}
