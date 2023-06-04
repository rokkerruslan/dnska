package resolve

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

type SimpleForwardUDPResolverOpts struct {
	ForwardAddr          netip.AddrPort
	DumpMalformedPackets bool
	L                    zerolog.Logger
}

func NewSimpleForwardUDPResolver(opts SimpleForwardUDPResolverOpts) *SimpleForwardUDPResolver {
	return &SimpleForwardUDPResolver{
		addr:                 opts.ForwardAddr,
		dumpMalformedPackets: opts.DumpMalformedPackets,
		l:                    opts.L,
	}
}

type SimpleForwardUDPResolver struct {
	addr                 netip.AddrPort
	dumpMalformedPackets bool

	l zerolog.Logger
}

func (sfr *SimpleForwardUDPResolver) Resolve(ctx context.Context, in proto.Message) (proto.Message, error) {

	// Currently, we dial up a new UDP socket for every lookup
	// operation. It's not efficient, but it appropriate (and simple)
	// solution for non-concurrent lookup operations. We will change
	// it later.

	conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(sfr.addr))
	if err != nil {
		return proto.Message{}, fmt.Errorf("failed to dial: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			sfr.l.Printf("failed to close udp conn: %v", err)
		}
	}()

	enc := proto.NewEncoder(make([]byte, limits.UDPPayloadSizeLimit))

	outBuf, err := enc.Encode(in)
	if err != nil {
		return proto.Message{}, fmt.Errorf("failed to encode: %v", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return proto.Message{}, fmt.Errorf("failed to set write deadline: %v", err)
		}
	}

	if _, err := conn.Write(outBuf); err != nil {
		return proto.Message{}, fmt.Errorf("failed to send packet: %v", err)
	}

	out := make([]byte, limits.UDPPayloadSizeLimit)

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return proto.Message{}, fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	n, err := conn.Read(out)
	if err != nil {
		return proto.Message{}, fmt.Errorf("failed to recieve packet: %v", err)
	}
	out = out[:n]

	dec := proto.NewDecoder()

	outMsg, err := dec.Decode(out)
	if err != nil {
		if sfr.dumpMalformedPackets {
			// todo: dump and query too.
			debug.DumpMalformedPacket(out)
		}

		return proto.Message{}, fmt.Errorf("failed to decode packet: %v", err)
	}

	if in.Header.ID != outMsg.Header.ID {
		if sfr.dumpMalformedPackets {
			// todo: dump and query too.
			debug.DumpMalformedPacket(out)
		}

		return proto.Message{}, fmt.Errorf("id is not equal :: in=%d out=%d", in.Header.ID, outMsg.Header.ID)
	}

	return outMsg, nil
}
