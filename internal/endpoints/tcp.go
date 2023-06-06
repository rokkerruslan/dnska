package endpoints

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/netip"
	"time"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewTCPEndpoint(addr netip.AddrPort, resolver resolve.Resolver, l zerolog.Logger) *TCPEndpoint {
	return &TCPEndpoint{
		addr:     addr,
		resolver: resolver,
		l:        l,

		exit: make(chan struct{}),
	}
}

type TCPEndpoint struct {
	addr      netip.AddrPort
	resolver  resolve.Resolver
	resolver2 resolve.ResolverV2
	l         zerolog.Logger

	exit   chan struct{}
	onStop func()
}

func (t *TCPEndpoint) Name() string {
	return "tcp"
}

func (t *TCPEndpoint) Start(onStop func()) error {
	t.onStop = onStop

	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(t.addr))
	if err != nil {
		return fmt.Errorf("tcp :: failed to listen :: error=%v", err)
	}

	defer func() {
		closeErr := listener.Close()
		if closeErr != nil {
			t.l.Printf("tcp :: failed to close connection :: error=%v", closeErr)
		}
	}()

	t.l.Printf("tcp :: starts on %v | %v", listener.Addr(), t.addr)

	for {
		select {
		case <-t.exit:
			return nil
		default:
			// nop
		}

		conn, err := listener.AcceptTCP()
		if err != nil {
			t.l.Printf("tcp :: failed to establish connection :: error=%v", err)
			continue
		}

		t.step(conn)
		conn.Close()
	}
}

func (t *TCPEndpoint) step(conn *net.TCPConn) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.l.Printf("tcp :: failed to set deadline :: error=%v", err)
		return
	}

	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		packetReadErrorsTotal.Inc()
		t.l.Printf("tcp :: failed to read length field :: error=%v", err)
		return
	}

	buf := make([]byte, length)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}

		packetReadErrorsTotal.Inc()
		t.l.Printf("tcp :: failed to read data :: error=%v", err)
		return
	}

	startTs := time.Now()

	dec := proto.NewDecoder()

	inMsg, err := dec.Decode(buf)
	if err != nil {
		packetDecodeErrorsTotal.Inc()
		t.l.Printf("tcp :: failed to decode message :: error=%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	outMsg, err := t.resolver.Resolve(ctx, inMsg)
	if err != nil {
		t.l.Printf("tcp :: failed to lookup :: error=%v", err)
		return
	}

	enc := proto.NewEncoder(make([]byte, 512))

	dataBuf, err := enc.Encode(outMsg)
	if err != nil {
		packetEncodeErrorsTotal.Inc()
		t.l.Printf("tcp :: failed to encode message :: error=%v", err)
		return
	}

	outBuf := make([]byte, 2+len(dataBuf))
	binary.BigEndian.PutUint16(outBuf, uint16(len(dataBuf)))
	copy(outBuf[2:], dataBuf)

	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.l.Printf("tcp :: failed to set write deadline :: error=%v", err)
		return
	}
	if _, err := conn.Write(outBuf); err != nil {
		t.l.Printf("tcp :: failed to write :: error=%v", err)
		packetWriteErrorsTotal.Inc()
		return
	}

	t.l.Printf("trace :: tcp :: total time is %v :: q=%s", time.Since(startTs), inMsg.Question[0].Name)

	successesProcessedOpsTotal.Inc()
}

func (t *TCPEndpoint) Stop() error {
	t.onStop()

	close(t.exit)

	return nil
}
