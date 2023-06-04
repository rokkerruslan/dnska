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

// Contains multiple connections to one addr
// addr -> conn1 COUNTER:100
//      -> conn2 COUNTER:200
// Resolve process
//  Select one of connection.
//  Request from connection.

type ForwardUDPResolverOpts struct {
	ForwardAddr          []netip.AddrPort
	DumpMalformedPackets bool
	L                    zerolog.Logger
}

func NewForwardUDPResolver(opts SimpleForwardUDPResolverOpts) (*ForwardUDPResolver, error) {
	conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(opts.ForwardAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &ForwardUDPResolver{
		addr:                 opts.ForwardAddr,
		dumpMalformedPackets: opts.DumpMalformedPackets,
		conn:                 conn,
		l:                    opts.L,
	}, nil
}

type ForwardUDPResolver struct {
	addr                 netip.AddrPort
	dumpMalformedPackets bool

	conn *net.UDPConn

	l zerolog.Logger
}

func (fur *ForwardUDPResolver) Resolve(ctx context.Context, in proto.Message) (proto.Message, error) {
	enc := proto.NewEncoder(make([]byte, limits.UDPPayloadSizeLimit))

	outBuf, err := enc.Encode(in)
	if err != nil {
		return proto.Message{}, fmt.Errorf("failed to encode: %v", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := fur.conn.SetWriteDeadline(deadline); err != nil {
			return proto.Message{}, fmt.Errorf("failed to set write deadline: %v", err)
		}
	}

	if _, err := fur.conn.Write(outBuf); err != nil {
		return proto.Message{}, fmt.Errorf("failed to send packet: %v", err)
	}

	out := make([]byte, limits.UDPPayloadSizeLimit)

	if deadline, ok := ctx.Deadline(); ok {
		if err := fur.conn.SetReadDeadline(deadline); err != nil {
			return proto.Message{}, fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	n, err := fur.conn.Read(out)
	if err != nil {
		return proto.Message{}, fmt.Errorf("failed to recieve packet: %v", err)
	}
	out = out[:n]

	dec := proto.NewDecoder()

	outMsg, err := dec.Decode(out)
	if err != nil {
		if fur.dumpMalformedPackets {
			// todo: dump and query too.
			debug.DumpMalformedPacket(out)
		}

		return proto.Message{}, fmt.Errorf("failed to decode packet: %v", err)
	}

	if in.Header.ID != outMsg.Header.ID {
		// todo: dump and query too.
		debug.DumpMalformedPacket(out)

		return proto.Message{}, fmt.Errorf("id is not equal :: in=%d out=%d", in.Header.ID, outMsg.Header.ID)
	}

	return outMsg, nil
}

func (fur *ForwardUDPResolver) Close() error {
	if err := fur.conn.Close(); err != nil {
		return err
	}

	return nil
}
