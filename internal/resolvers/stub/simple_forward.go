package stub

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/internal/udp"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

var (
	dnskaSfurRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_sfur_requests_total",
		Help: "The total number of requests sent to the simple forward resolver",
	})

	dnskaSfurErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_sfur_errors_total",
		Help: "The total number of errors encountered by the simple forward resolver",
	})
)

// SimpleForwardUDPResolverOpts defines the options for creating a SimpleForwardUDPResolver.
//
// Options:
//
//   - UdpConn is the (connected) UDP socket to the upstream DNS server.
//
//   - SetRecursionDesired, if true, sets the RD flag in queries so the upstream
//     server performs recursive resolution.
//
//   - MalformedPacketDumper receives packets that fail validation.
type SimpleForwardUDPResolverOpts struct {
	UdpConn               udp.Conn
	SetRecursionDesired   bool
	MalformedPacketDumper debug.MalformedPacketDumper
}

// SimpleForwardUDPResolver forwards each query to a single upstream over UDP.
//
// It is NOT safe for concurrent use: it owns one UDP socket and mutates a
// per-request ID counter without synchronization, so callers must serialize
// calls to Resolve (single flight).
type SimpleForwardUDPResolver struct {
	udpConn               udp.Conn
	malformedPacketDumper debug.MalformedPacketDumper
	setRecursionDesired   bool
	currentHeaderID       uint16
}

const defaultRequestResponseTimeout = time.Second

func NewSimpleForwardUDPResolver(opts SimpleForwardUDPResolverOpts) *SimpleForwardUDPResolver {
	return &SimpleForwardUDPResolver{
		udpConn:               opts.UdpConn,
		malformedPacketDumper: opts.MalformedPacketDumper,
		setRecursionDesired:   opts.SetRecursionDesired,
		currentHeaderID:       0,
	}
}

func (sfr *SimpleForwardUDPResolver) Resolve(
	ctx context.Context,
	in *proto.InternalMessage,
) (*proto.InternalMessage, error) {
	dnskaSfurRequestsTotal.Inc()

	headerID := sfr.currentHeaderID
	sfr.currentHeaderID++

	enc := proto.NewEncoder(make([]byte, limits.DefaultUDPPayloadSizeLimit))

	msg := in.ToProtoMessage()
	if sfr.setRecursionDesired {
		msg.Header.RecursionDesired = true
	}
	msg.Header.ID = headerID

	outBuf, err := enc.Encode(msg)
	if err != nil {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to encode: %v", err)
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultRequestResponseTimeout)
	}

	if err := sfr.udpConn.SetWriteDeadline(deadline); err != nil {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to set write deadline: %v", err)
	}

	n, err := sfr.udpConn.Write(outBuf)
	if err != nil {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to send packet: %v", err)
	}

	if n != len(outBuf) {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to send packet: sent %d bytes, expected %d", n, len(outBuf))
	}

	out := make([]byte, limits.DefaultUDPPayloadSizeLimit)

	if err := sfr.udpConn.SetReadDeadline(deadline); err != nil {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to set read deadline: %v", err)
	}

	n, err = sfr.udpConn.Read(out)
	if err != nil {
		dnskaSfurErrorsTotal.Inc()
		return nil, fmt.Errorf("failed to receive packet: %v", err)
	}
	out = out[:n]

	dec := proto.NewDecoder()

	outMsg, err := dec.Decode(out)
	if err != nil {
		dnskaSfurErrorsTotal.Inc()
		sfr.malformedPacketDumper.Dump(out)

		return nil, fmt.Errorf("failed to decode packet: %v", err)
	}

	if !outMsg.Header.Response {
		dnskaSfurErrorsTotal.Inc()
		sfr.malformedPacketDumper.Dump(out)

		return nil, fmt.Errorf("response flag is not set")
	}

	if outMsg.Header.ID != headerID {
		dnskaSfurErrorsTotal.Inc()
		sfr.malformedPacketDumper.Dump(out)

		return nil, fmt.Errorf("id is not equal :: in=%d out=%d", headerID, outMsg.Header.ID)
	}

	if outMsg.Header.Truncated {
		dnskaSfurErrorsTotal.Inc()
		sfr.malformedPacketDumper.Dump(out)

		return nil, fmt.Errorf("truncated flag is set")
	}

	// spew.Dump(outMsg)

	return proto.FromProtoMessage(outMsg), nil
}
