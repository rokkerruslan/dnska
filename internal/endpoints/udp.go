package endpoints

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewUDPEndpoint(addr netip.AddrPort, resolver resolve.Resolver, l zerolog.Logger) *UDPEndpoint {
	return &UDPEndpoint{
		addr:     addr,
		resolver: resolver,
		l:        l,

		exit: make(chan struct{}),
	}
}

type UDPEndpoint struct {
	addr      netip.AddrPort
	resolver  resolve.Resolver
	resolver2 resolve.ResolverV2
	l         zerolog.Logger

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
		ep.l.Printf("failed to listen :: error=%v", err)
	}

	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			ep.l.Printf("failed to close connection :: error=%v", closeErr)
		}
	}()

	ep.l.Printf("starts on %v | %v", conn.LocalAddr(), ep.addr)

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
		ep.l.Printf("failed to set deadline :: error=%v", err)
		return
	}
	_, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}

		packetReadErrorsTotal.Inc()
		ep.l.Printf("failed to read from udp :: error=%v", err)
		return
	}

	startTs := time.Now()

	dec := proto.NewDecoder()

	inMsg, err := dec.Decode(buf)
	if err != nil {
		packetDecodeErrorsTotal.Inc()
		ep.l.Printf("failed to decode message :: error=%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	outMsg, err := ep.resolver.Resolve(ctx, inMsg)
	if err != nil {
		ep.l.Printf("failed to lookup :: error=%v", err)
		return
	}

	enc := proto.NewEncoder(buf)

	buf, err = enc.Encode(outMsg)
	if err != nil {
		packetEncodeErrorsTotal.Inc()
		ep.l.Printf("failed to encode message :: error=%v", err)
		return
	}

	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		ep.l.Printf("failed to set write deadline :: error=%v", err)
	}
	if _, err := conn.WriteToUDP(buf, remoteAddr); err != nil {
		ep.l.Printf("failed to write to udp :: error=%v", err)
		packetWriteErrorsTotal.Inc()
		return
	}

	ep.l.Printf("trace :: total time is %v :: q=%ep", time.Since(startTs), inMsg.Question[0].Name)

	successesProcessedOpsTotal.Inc()
}

func (ep *UDPEndpoint) Stop() error {
	ep.onStop()

	close(ep.exit)

	return nil
}
