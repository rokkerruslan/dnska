package resolve

import (
	"context"
	"fmt"
	"math/rand"
	"net/netip"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type LookupOpts struct {
	Name              string
	Type              proto.QType
	Class             proto.QClass
	ID                uint16
	DumpUnknownPacket bool
	L                 zerolog.Logger
}

func cycle(ctx context.Context, opts LookupOpts) (proto.Message, error) {
	forward := namedRootIndex[rand.Intn(len(namedRootIndex))]

	n := 0
	for {
		n++

		addr, err := netip.ParseAddr(forward)
		if err != nil {
			return proto.Message{}, fmt.Errorf("failed to parse next addr :: addr=%v, error=%v", forward, err)
		}

		addrPort := netip.AddrPortFrom(addr, 53)

		in := proto.Message{
			Header: proto.Header{
				ID:               opts.ID,
				RecursionDesired: false,
				QDCount:          1,
			},
			Question: []proto.Question{
				{
					Name:  opts.Name,
					Type:  opts.Type,
					Class: opts.Class,
				},
			},
		}

		res := NewSimpleForwardUDPResolver(SimpleForwardUDPResolverOpts{
			ForwardAddr:          addrPort,
			DumpMalformedPackets: opts.DumpUnknownPacket,
			L:                    opts.L,
		})

		out, err := res.Resolve(ctx, in)
		if err != nil {
			return proto.Message{}, err
		}

		if len(out.Answer) != 0 && out.Header.RCode != proto.RCodeServerFailure {

			if _, ok := FirstRecord(out.Answer, opts.Type); ok {
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
				out2, err := cycle(ctx, LookupOpts{
					Name:              record.RData,
					Type:              proto.QTypeA,
					Class:             proto.ClassIN,
					ID:                1,
					DumpUnknownPacket: true,
					L:                 opts.L,
				})
				if err != nil {
					return proto.Message{}, err
				}

				out.Header.ANCount += out2.Header.ANCount
				out.Answer = append(out.Answer, out2.Answer...)

				return out, nil
			}

			return out, nil
		}

		// NS records cause both the usual additional section processing to locate
		// a type A record, and, when used in a referral, a special search of the
		// zone in which they reside for glue information.
		if record, ok := FirstRecord(out.Additional, proto.QTypeA); ok {
			//opts.L.Printf("found additional part :: name=%v, data=%v", record.Name, record.RData)

			forward = record.RData
			continue
		}

		record, ok := FirstRecord(out.Authority, proto.QTypeNS)
		if !ok {
			return out, nil
		}

		nsOut, err := cycle(ctx, LookupOpts{
			Name:              record.RData,
			Type:              proto.QTypeA,
			Class:             proto.ClassIN,
			ID:                1,
			DumpUnknownPacket: opts.DumpUnknownPacket,
			L:                 opts.L,
		})
		if err != nil {
			return proto.Message{}, err
		}

		if record, ok := findRecord(nsOut.Answer, proto.QTypeA); ok {
			forward = record.RData
			continue
		}

		// RFC1034 says that CNAME RRs cause special action in DNS software. When
		// a name server fails to find a desired RR in the resource set associated with the
		// domain name, it checks to see if the resource set consists of a CNAME
		// record with a matching class.  If so, the name server includes the CNAME
		// record in the response and restarts the query at the domain name
		// specified in the data field of the CNAME record. The one exception to
		// this rule is that queries which match the CNAME type are not restarted.

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

func FirstRecord(list []proto.ResourceRecord, t proto.QType) (proto.ResourceRecord, bool) {
	return findRecord(list, t)
}
