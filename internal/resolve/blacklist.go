package resolve

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt

type BlacklistResolverOpts struct {
	AutoReloadInterval time.Duration
	BlacklistURL       string
	Next               HResolver
	Lg                 *slog.Logger
}

type BlacklistResolver struct {
	autoReloadInterval time.Duration
	next               HResolver

	blacklistURL string

	mu        sync.Mutex
	blacklist map[string]struct{}

	lg *slog.Logger
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

	return b.next.Resolve(ctx, in)
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
		blacklistURL:       opts.BlacklistURL,

		mu:        sync.Mutex{},
		blacklist: blacklist,

		next: opts.Next,
		lg:   opts.Lg,
	}
}

func (b *BlacklistResolver) reload() error {
	// download file
	// parse file
	// create index
	resp, err := http.Get(b.blacklistURL)
	if err != nil {
		return fmt.Errorf("failed to download blacklist: %v", err)
	}
	defer resp.Body.Close()

	// File format:
	// # Expires: 1 days
	// # Hosts contributed by Anudeep <anudeep@protonmail.com>
	// # Domain count: 42,531
	// # ===================================================================
	// 0.0.0.0 0001-cab8-4c8c-43de.reporo.net
	// 0.0.0.0 002-slq-470.mktoresp.com
	// 0.0.0.0 004-btr-463.mktoresp.com
	// 0.0.0.0 005.free-counters.co.uk

	blacklist := map[string]struct{}{}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		var ip, domain string
		n, err := fmt.Sscanf(line, "%s %s", &ip, &domain)
		if err != nil || n != 2 {
			return fmt.Errorf("failed to parse line: %q", line)
		}

		blacklist[domain] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan blacklist: %v", err)
	}

	b.lg.Info("blacklist reloaded", "count", len(blacklist))

	b.mu.Lock()
	defer b.mu.Unlock()

	b.blacklist = blacklist

	return nil
}
