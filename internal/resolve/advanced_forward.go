package resolve

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/netip"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

type AdvancedForwardUDPResolverOpts struct {
	UpstreamAddrPort     netip.AddrPort
	DumpMalformedPackets bool
	L                    *slog.Logger

	Sub Resolver
}

type AdvancedForwardUDPResolver struct {
	dumpMalformedPackets bool
	upstreaminfo         upstreaminfo

	conn *net.UDPConn

	in chan req

	mu             sync.Mutex
	channels       map[uint16]chan ans
	indexAllocator IndexAllocator

	l *slog.Logger
}

type upstreaminfo struct {
	addrPort       netip.AddrPort
	payloadBufSize int
}

func NewAdvancedForwardUDPResolver(
	opts AdvancedForwardUDPResolverOpts,
) *AdvancedForwardUDPResolver {

	conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(opts.UpstreamAddrPort))
	if err != nil {
		panic(err)
		// return proto.Message{}, fmt.Errorf("failed to dial: %v", err)
	}

	r := AdvancedForwardUDPResolver{
		upstreaminfo: upstreaminfo{
			addrPort:       opts.UpstreamAddrPort,
			payloadBufSize: limits.DefaultUDPPayloadSizeLimit,
		},
		conn:                 conn,
		in:                   make(chan req),
		channels:             map[uint16]chan ans{},
		indexAllocator:       IndexAllocator{cur: 0, max: 4},
		dumpMalformedPackets: opts.DumpMalformedPackets,
		l:                    opts.L,
	}

	go r.sender()
	go r.receiver()

	return &r
}

type req struct {
	msg *proto.InternalMessage
	out chan ans
}

type ans struct {
	msg *proto.InternalMessage
	err error
}

func (afr *AdvancedForwardUDPResolver) Resolve(
	ctx context.Context,
	in *proto.InternalMessage,
) (
	*proto.InternalMessage,
	error,
) {
	out := make(chan ans)

	afr.in <- req{msg: in, out: out}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ans, ok := <-out:
		if !ok {
			return nil, errors.New("not available")
		}

		return ans.msg, ans.err
	}
}

func (afr *AdvancedForwardUDPResolver) sender() {
	in := make([]byte, afr.upstreaminfo.payloadBufSize)

	for req := range afr.in {
		identifier, ok := afr.reserve(req.out)
		if !ok {
			failedReserveTotal.Inc()
			close(req.out)
			continue
		}

		enc := proto.NewEncoder(in)

		m := req.msg.ToProtoMessage()
		m.Header.ID = identifier

		outBuf, err := enc.Encode(m)
		if err != nil {
			afr.l.Error("failed to encode", "error", err)
			continue
		}

		// if deadline, reserved := ctx.Deadline(); reserved {
		// 	if err := conn.SetWriteDeadline(deadline); err != nil {
		// 		return proto.Message{}, fmt.Errorf("failed to set write deadline: %v", err)
		// 	}
		// }

		if _, err := afr.conn.Write(outBuf); err != nil {
			afr.l.Error("failed to write", "error", err)
			continue
		}
	}
}

func (afr *AdvancedForwardUDPResolver) receiver() {
	out := make([]byte, afr.upstreaminfo.payloadBufSize)

	for {
		n, err := afr.conn.Read(out)
		if err != nil {
			afr.l.Error("failed to read: %v", err)
			break
		}

		dec := proto.NewDecoder()

		outMsg, err := dec.Decode(out[:n])
		if err != nil {
			afr.l.Error("failed to decode package", "error", err)
			if afr.dumpMalformedPackets {
				debug.DumpMalformedPacket(out)
			}

			continue
		}

		index := outMsg.Header.ID

		ch, known := afr.free(index)
		if !known {
			afr.l.Error("nobody wants this message", "id", outMsg.Header.ID)
			continue
		}

		afr.l.Error("found receiver and deallocate chan", "id", outMsg.Header.ID)

		m := proto.FromProtoMessage(outMsg)

		// todo: check errors

		select {
		case ch <- ans{msg: m}:
		default:
			clientGoneTotal.Inc()
		}

		close(ch)
	}
}

func (afr *AdvancedForwardUDPResolver) Close() {

	// todo: stop sender?
	// todo: stop receiver?

	defer func() {
		if err := afr.conn.Close(); err != nil {
			afr.l.Error("failed to close udp conn: %v", err)
		}
	}()
}

func (afr *AdvancedForwardUDPResolver) reserve(in chan ans) (uint16, bool) {
	afr.mu.Lock()
	defer afr.mu.Unlock()

	idx, reserved := afr.indexAllocator.Reserve()
	if reserved {
		afr.channels[idx] = in
	}

	return idx, reserved
}

func (afr *AdvancedForwardUDPResolver) free(index uint16) (chan ans, bool) {
	afr.mu.Lock()
	defer afr.mu.Unlock()

	ch, known := afr.channels[index]
	if known {
		afr.indexAllocator.Free(index)
		delete(afr.channels, index)
	}

	return ch, known
}

func NewIndexAllocator(max uint16) *IndexAllocator {
	return &IndexAllocator{
		max:  max,
		cur:  0,
		free: nil,
	}
}

type IndexAllocator struct {
	max  uint16
	cur  uint16
	free []uint16
}

func (ia *IndexAllocator) Reserve() (uint16, bool) {
	if len(ia.free) != 0 {
		idx := ia.free[len(ia.free)-1]
		ia.free = ia.free[:len(ia.free)-1]

		return idx, true
	}

	index := ia.cur
	if ia.cur == ia.max {
		return 0, false // todo: run cleanup procedure?
	}

	ia.cur++

	return index, true
}

func (ia *IndexAllocator) FreeLen() int {
	return len(ia.free)
}

func (ia *IndexAllocator) Free(index ...uint16) {
	ia.free = append(ia.free, index...)
}

// Diagnostic

var (
	failedReserveTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_advanced_resolver_failed_reserve_total",
		Help: "The total number of failed trying to reserve index",
	})

	clientGoneTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_anvanced_resolver_client_gone_total",
		Help: "The total number of resolve calls when client do not await response",
	})
)
