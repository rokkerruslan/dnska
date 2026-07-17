package stub_test

import (
	"context"
	"testing"
	"time"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/internal/resolvers/stub"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

// loopbackConn is a fake udp.Conn that turns the written query into a matching
// response, so SimpleForwardUDPResolver can be exercised without a network.
type loopbackConn struct {
	query []byte
}

func (c *loopbackConn) Write(b []byte) (int, error) {
	c.query = append([]byte(nil), b...)
	return len(b), nil
}

func (c *loopbackConn) Read(b []byte) (int, error) {
	msg, err := proto.NewDecoder().Decode(c.query)
	if err != nil {
		return 0, err
	}

	msg.Header.Response = true // flip the echoed query into a response

	resp, err := proto.NewEncoder(make([]byte, limits.DefaultUDPPayloadSizeLimit)).Encode(msg)
	if err != nil {
		return 0, err
	}

	return copy(b, resp), nil
}

func (c *loopbackConn) SetReadDeadline(time.Time) error  { return nil }
func (c *loopbackConn) SetWriteDeadline(time.Time) error { return nil }
func (c *loopbackConn) Close() error                     { return nil }

func TestSimpleForwardUDPResolver_Resolve(t *testing.T) {
	resolver := stub.NewSimpleForwardUDPResolver(stub.SimpleForwardUDPResolverOpts{
		UdpConn:               &loopbackConn{},
		SetRecursionDesired:   false,
		MalformedPacketDumper: debug.NewFakeMalformedPacketDumper(),
	})

	in := proto.FromProtoMessage(&proto.Message{
		Header: proto.Header{QDCount: 1},
		Question: []proto.Question{
			{Name: "example.com", Type: proto.QTypeA, Class: proto.ClassIN},
		},
	})

	out, err := resolver.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if out.Question.Name != "example.com" {
		t.Errorf("Question.Name = %q, want %q", out.Question.Name, "example.com")
	}
}
