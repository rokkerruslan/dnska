package resolve_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/netip"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"

	tst "github.com/rokkerruslan/dnska/testing"
)

func TestAdvancedForwardUDPResolver(t *testing.T) {
	addrPort := netip.MustParseAddrPort("1.1.1.1:53")

	buf := bytes.Buffer{}

	logger := slog.New(slog.NewTextHandler(&buf, nil))

	r := resolve.NewAdvancedForwardUDPResolver(resolve.AdvancedForwardUDPResolverOpts{
		UpstreamAddrPort:     addrPort,
		DumpMalformedPackets: false,
		L:                    logger,
	})

	inM := proto.Message{
		Header: proto.Header{
			ID:               1,
			QDCount:          1,
			RecursionDesired: true,
		},
		Question: []proto.Question{
			{
				Name:  "ya.ru",
				Type:  proto.QTypeA,
				Class: proto.ClassIN,
			},
		},
	}

	in := proto.FromProtoMessage(inM)

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	time.Sleep(1 * time.Second)

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	go func() {
		out, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(out.Answer[0].RData)
		}
	}()

	time.Sleep(time.Second)

	t.Log(spew.Sdump(r))

	r.Close()

	// t.Log(buf.String())
}

func TestIndexAllocator(t *testing.T) {
	ia := resolve.NewIndexAllocator(3)

	var ok bool
	var v uint16

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(0), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(1), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(2), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, false)
	tst.Assert(t, uint16(0), v)

	ia.Free(uint16(2))

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(2), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, false)
	tst.Assert(t, uint16(0), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, false)
	tst.Assert(t, uint16(0), v)

	ia.Free(uint16(2), uint16(0), uint16(1))
	tst.Assert(t, ia.FreeLen(), 3)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(1), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(0), v)

	v, ok = ia.Reserve()
	tst.Assert(t, ok, true)
	tst.Assert(t, uint16(2), v)
}
