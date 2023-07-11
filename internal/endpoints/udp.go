package endpoints

import (
	"context"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewUDPEndpoint(addr netip.AddrPort, resolver resolve.Resolver, l *slog.Logger) *UDPEndpoint {
	return &UDPEndpoint{
		addr:     addr,
		resolver: resolver,
		l:        l,

		exit: make(chan struct{}),
	}
}

type UDPEndpoint struct {
	addr     netip.AddrPort
	resolver resolve.Resolver
	l        *slog.Logger

	exit   chan struct{}
	onStop func()
}

func (ep *UDPEndpoint) Name() string {
	return "udp"
}

func (ep *UDPEndpoint) Start(onStop func()) {
	ep.onStop = onStop

	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(ep.addr))
	if err != nil {
		ep.l.Error("failed to listen", "error", err)
		return
	}

	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			ep.l.Error("failed to close connection", closeErr)
		}
	}()

	ep.l.Info("udp starts", "addr", conn.LocalAddr())

	for {
		select {
		case <-ep.exit:
			return
		default:
			// nop
		}

		ep.step(conn)
	}
}

func (ep *UDPEndpoint) step(conn *net.UDPConn) {
	buf := make([]byte, 512)

	if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		ep.l.Error("failed to set deadline", "error", err)
		return
	}
	_, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}

		packetReadErrorsTotal.Inc()
		ep.l.Error("failed to read from udp :: error=%v", err)
		return
	}

	startTs := time.Now()

	dec := proto.NewDecoder()

	inMsg, err := dec.Decode(buf)
	if err != nil {
		packetDecodeErrorsTotal.Inc()
		ep.l.Error("failed to decode message", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inMsg2 := proto.FromProtoMessage(inMsg)

	outMsg, err := ep.resolver.Resolve(ctx, inMsg2)
	if err != nil {
		ep.l.Error("failed to lookup :: error=%v", err)
		return
	}

	enc := proto.NewEncoder(buf)

	outMsg2 := outMsg.ToProtoMessage()

	// todo: ID of message can be different and we need to copy from input
	outMsg2.Header.ID = inMsg.Header.ID

	// todo: move to different abstraction layer
	outMsg2.Header.Response = true
	outMsg2.Header.RecursionAvailable = true
	outMsg2.Header.RecursionDesired = inMsg.Header.RecursionDesired

	buf, err = enc.Encode(outMsg2)
	if err != nil {
		packetEncodeErrorsTotal.Inc()
		ep.l.Error("failed to encode message :: error=%v", err)
		return
	}

	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		ep.l.Error("failed to set write deadline :: error=%v", err)
	}
	if _, err := conn.WriteToUDP(buf, remoteAddr); err != nil {
		ep.l.Error("failed to write to udp", "error", err)
		packetWriteErrorsTotal.Inc()
		return
	}

	ep.l.Info(
		"done",
		slog.Duration("total", time.Since(startTs)),
		slog.String("name", inMsg2.Question.Name),
	)

	successesProcessedOpsTotal.Inc()
}

func (ep *UDPEndpoint) Stop() error {
	ep.onStop()

	close(ep.exit)

	return nil
}
