package resolve

import (
	"context"
	"math"
	"time"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt

type BlacklistResolverOpts struct {
	AutoReloadInterval time.Duration
	BlacklistURL       string
	Pass               Resolver
}

type BlacklistResolver struct {
	autoReloadInterval time.Duration
	pass               Resolver

	blacklist map[string]struct{}
}

var answerFuckOff = proto.ResourceRecord{
	Name:     "",
	Type:     proto.QTypeA,
	Class:    proto.ClassIN,
	TTL:      math.MaxUint32,
	RDLength: 9,
	RData:    "127.0.0.1",
}

func (b *BlacklistResolver) Resolve(ctx context.Context, in *proto.InternalMessage) (*proto.InternalMessage, error) {
	q := in.Question

	if _, ok := b.blacklist[q.Name]; ok {
		out := proto.InternalMessage{
			Question:   q,
			Answer:     []proto.ResourceRecord{answerFuckOff},
			Authority:  nil,
			Additional: nil,
		}

		return &out, nil
	}

	return b.pass.Resolve(ctx, in)
}

func NewBlacklistResolver(opts BlacklistResolverOpts) *BlacklistResolver {
	if opts.AutoReloadInterval <= time.Second {
		panic("auto-reload interval too small")
	}

	// start downloader
	//  download file
	//  parse file
	//  create index
	//

	blacklist := map[string]struct{}{
		"www.yahoo.com": {},
	}

	return &BlacklistResolver{
		autoReloadInterval: opts.AutoReloadInterval,
		blacklist:          blacklist,
		pass:               opts.Pass,
	}
}
