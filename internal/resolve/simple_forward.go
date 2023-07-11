package resolve

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

type SimpleForwardUDPResolverOpts struct {
	ForwardAddr          netip.AddrPort
	DumpMalformedPackets bool
	L                    *slog.Logger
}

func NewSimpleForwardUDPResolver(opts SimpleForwardUDPResolverOpts) *SimpleForwardUDPResolver {
	return &SimpleForwardUDPResolver{
		addr:                 opts.ForwardAddr,
		dumpMalformedPackets: opts.DumpMalformedPackets,
		l:                    opts.L,
	}
}

// SimpleForwardUDPResolver allocate local port every resolve.
type SimpleForwardUDPResolver struct {
	addr                 netip.AddrPort
	dumpMalformedPackets bool

	l *slog.Logger
}

func (sfr *SimpleForwardUDPResolver) Resolve(
	ctx context.Context,
	in *proto.InternalMessage,
) (*proto.InternalMessage, error) {

	// Currently, we dial up a new UDP socket for every lookup
	// operation. It's not efficient, but it appropriate (and simple)
	// solution for non-concurrent lookup operations. We will change
	// it later.

	conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(sfr.addr))
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			sfr.l.Error("failed to close udp conn", "error", err)
		}
	}()

	enc := proto.NewEncoder(make([]byte, limits.DefaultUDPPayloadSizeLimit))

	outBuf, err := enc.Encode(in.ToProtoMessage())
	if err != nil {
		return nil, fmt.Errorf("failed to encode: %v", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set write deadline: %v", err)
		}
	}

	if _, err := conn.Write(outBuf); err != nil {
		return nil, fmt.Errorf("failed to send packet: %v", err)
	}

	out := make([]byte, limits.DefaultUDPPayloadSizeLimit)

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	n, err := conn.Read(out)
	if err != nil {
		return nil, fmt.Errorf("failed to recieve packet: %v", err)
	}
	out = out[:n]

	dec := proto.NewDecoder()

	outMsg, err := dec.Decode(out)
	if err != nil {
		if sfr.dumpMalformedPackets {
			// todo: dump and query too.
			debug.DumpMalformedPacket(out)
		}

		return nil, fmt.Errorf("failed to decode packet: %v", err)
	}

	if outMsg.Header.ID != 0 {
		if sfr.dumpMalformedPackets {
			// todo: dump and query too.
			debug.DumpMalformedPacket(out)
		}

		return nil, fmt.Errorf("id is not equal :: in=0 out=%d", outMsg.Header.ID)
	}

	return proto.FromProtoMessage(outMsg), nil
}
