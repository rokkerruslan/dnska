package resolve

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/netip"
	"sync"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewIterativeResolver(l *slog.Logger) *IterativeResolver {
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

	l *slog.Logger
}

func (fr *IterativeResolver) Resolve(
	ctx context.Context,
	in *proto.InternalMessage,
) (
	*proto.InternalMessage,
	error,
) {
	question := in.Question

	opts := cycleOpts{
		Name:              question.Name,
		Type:              question.Type,
		Class:             question.Class,
		DumpUnknownPacket: fr.dumpMalformedPackets,
		L:                 fr.l,
	}

	// The top level algorithm has four steps:
	//
	//   1. See if the answer is in local information, and if so return
	//      it to the client.
	//

	// iterative resolver does not have any local information

	found, err := fr.cycle(ctx, opts)
	if err == nil {
		out := proto.InternalMessage{
			Question:   question,
			Answer:     found.Answer,
			Authority:  found.Authority,
			Additional: found.Additional,
		}

		return &out, nil
	}

	return nil, fmt.Errorf("failed to iterative lookup:%v", err)
}

type cycleOpts struct {
	Name              string
	Type              proto.QType
	Class             proto.QClass
	DumpUnknownPacket bool
	L                 *slog.Logger
}

func (fr *IterativeResolver) cycle(ctx context.Context, opts cycleOpts) (*proto.InternalMessage, error) {
	forward := namedRootIndex[rand.Intn(len(namedRootIndex))]

	n := 0
	for {
		n++

		addr, err := netip.ParseAddr(forward)
		if err != nil {
			return nil, fmt.Errorf("failed to parse next addr :: addr=%v, error=%v", forward, err)
		}

		addrPort := netip.AddrPortFrom(addr, 53)

		// Instantiave CacheResolver/AdvancedForwardUDPResolver for addrPort
		// and store it for perfomance improving.

		res := NewSimpleForwardUDPResolver(SimpleForwardUDPResolverOpts{
			ForwardAddr:          addrPort,
			DumpMalformedPackets: opts.DumpUnknownPacket,
			L:                    opts.L,
		})

		in := proto.InternalMessage{
			Question: proto.Question{
				Name:  opts.Name,
				Type:  opts.Type,
				Class: opts.Class,
			},
		}

		out, err := res.Resolve(ctx, &in)
		if err != nil {
			return nil, err
		}

		if len(out.Answer) != 0 {
			if _, ok := findRecord(out.Answer, opts.Type); ok {
				return out, nil
			}

			// RFC1034 says that CNAME RRs cause special action in DNS software. When
			// a name server fails to find a desired RR in the resource set associated with the
			// domain name, it checks to see if the resource set consists of a CNAME
			// record with a matching class. If so, the name server includes the CNAME
			// record in the response and restarts the query at the domain name
			// specified in the data field of the CNAME record. The one exception to
			// this rule is that queries which match the CNAME type are not restarted.

			if record, ok := findRecord(out.Answer, proto.QTypeCName); ok {
				out2, err := fr.cycle(ctx, cycleOpts{
					Name:              record.RData,
					Type:              proto.QTypeA,
					Class:             proto.ClassIN,
					DumpUnknownPacket: true,
					L:                 opts.L,
				})
				if err != nil {
					return nil, err
				}

				out.Answer = append(out.Answer, out2.Answer...)

				return out, nil
			}

			return out, nil
		}

		// NS records cause both the usual additional section processing to locate
		// a type A record, and, when used in a referral, a special search of the
		// zone in which they reside for glue information.
		if record, ok := findRecord(out.Additional, proto.QTypeA); ok {
			//opts.L.Printf("found additional part :: name=%v, data=%v", record.Name, record.RData)

			forward = record.RData
			continue
		}

		record, ok := findRecord(out.Authority, proto.QTypeNS)
		if !ok {
			return out, nil
		}

		nsOut, err := fr.cycle(ctx, cycleOpts{
			Name:              record.RData,
			Type:              proto.QTypeA,
			Class:             proto.ClassIN,
			DumpUnknownPacket: opts.DumpUnknownPacket,
			L:                 opts.L,
		})
		if err != nil {
			return nil, err
		}

		if record, ok := findRecord(nsOut.Answer, proto.QTypeA); ok {
			forward = record.RData
			continue
		}

		return out, nil
	}
}

func findRecord(records []proto.ResourceRecord, t proto.QType) (proto.ResourceRecord, bool) {
	for _, el := range records {
		if el.Type == t {
			return el, true
		}
	}

	return proto.ResourceRecord{}, false
}
