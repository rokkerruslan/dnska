package endpoints

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewTCPEndpoint(addr netip.AddrPort, resolver resolve.Resolver, l *slog.Logger) *TCPEndpoint {
	return &TCPEndpoint{
		addr:     addr,
		resolver: resolver,
		l:        l,

		exit: make(chan struct{}),
	}
}

type TCPEndpoint struct {
	addr     netip.AddrPort
	resolver resolve.Resolver
	l        *slog.Logger

	exit   chan struct{}
	onStop func()
}

func (t *TCPEndpoint) Name() string {
	return "tcp"
}

func (t *TCPEndpoint) Start(onStop func()) {
	t.onStop = onStop

	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(t.addr))
	if err != nil {
		t.l.Error("tcp :: failed to listen :: error=%v", err)
		return
	}

	defer func() {
		closeErr := listener.Close()
		if closeErr != nil {
			t.l.Error("tcp :: failed to close connection :: error=%v", closeErr)
		}
	}()

	t.l.Error("tcp starts", "addr", listener.Addr())

	for {
		select {
		case <-t.exit:
			return
		default:
			// nop
		}

		conn, err := listener.AcceptTCP()
		if err != nil {
			t.l.Error("tcp :: failed to establish connection :: error=%v", err)
			continue
		}

		t.step(conn)
		conn.Close()
	}
}

func (t *TCPEndpoint) step(conn *net.TCPConn) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.l.Error("tcp :: failed to set deadline :: error=%v", err)
		return
	}

	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		packetReadErrorsTotal.Inc()
		t.l.Error("tcp :: failed to read length field :: error=%v", err)
		return
	}

	buf := make([]byte, length)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}

		packetReadErrorsTotal.Inc()
		t.l.Error("tcp :: failed to read data :: error=%v", err)
		return
	}

	startTs := time.Now()

	dec := proto.NewDecoder()

	inMsg, err := dec.Decode(buf)
	if err != nil {
		packetDecodeErrorsTotal.Inc()
		t.l.Error("tcp :: failed to decode message :: error=%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inMsg2 := proto.FromProtoMessage(inMsg)

	outMsg, err := t.resolver.Resolve(ctx, inMsg2)
	if err != nil {
		t.l.Error("tcp :: failed to lookup :: error=%v", err)
		return
	}

	outMsg2 := outMsg.ToProtoMessage()

	outMsg2.Header.ID = inMsg.Header.ID

	// todo: move to different abstraction layer
	outMsg2.Header.Response = true
	outMsg2.Header.RecursionAvailable = true
	outMsg2.Header.RecursionDesired = inMsg.Header.RecursionDesired

	enc := proto.NewEncoder(make([]byte, 512))

	dataBuf, err := enc.Encode(outMsg.ToProtoMessage())
	if err != nil {
		packetEncodeErrorsTotal.Inc()
		t.l.Error("tcp :: failed to encode message :: error=%v", err)
		return
	}

	outBuf := make([]byte, 2+len(dataBuf))
	binary.BigEndian.PutUint16(outBuf, uint16(len(dataBuf)))
	copy(outBuf[2:], dataBuf)

	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.l.Error("tcp :: failed to set write deadline :: error=%v", err)
		return
	}
	if _, err := conn.Write(outBuf); err != nil {
		t.l.Error("tcp :: failed to write :: error=%v", err)
		packetWriteErrorsTotal.Inc()
		return
	}

	t.l.Info(
		"done",
		slog.Duration("total", time.Since(startTs)),
		slog.String("name", inMsg2.Question.Name),
	)

	successesProcessedOpsTotal.Inc()
}

func (t *TCPEndpoint) Stop() error {
	t.onStop()

	close(t.exit)

	return nil
}
