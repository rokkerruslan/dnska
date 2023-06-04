package bucket

import (
	"encoding/gob"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// todo: Add cache for availability to cache
// todo: responses into a proxy name server.

// Interesting implementation.
// https://dgraph.io/blog/post/introducing-ristretto-high-perf-go-cache/

// A Bucket is a rudimentary in-memory cache.
type Bucket struct {
	path    string
	verbose bool
	l       *slog.Logger

	mu      sync.Mutex
	entries map[string]holder
}

type Opts struct {
	Path    string
	Verbose bool
	L       *slog.Logger
}

func New(opts Opts) *Bucket {
	b := Bucket{
		path:    opts.Path,
		verbose: opts.Verbose,
		entries: map[string]holder{},
		l:       opts.L,
	}

	func() {
		f, err := os.Open(opts.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				b.trace("info, cache file %s not found", opts.Path)
			} else {
				b.trace("info, cache opening failed: %v", err)
			}

			return
		}

		if err := gob.NewDecoder(f).Decode(&b.entries); err != nil {
			b.trace("error, decoding cache failed: %v", err)
		}
	}()

	return &b
}

func (b *Bucket) Set(key string, ent Entry, ttl time.Duration) {
	h := holder{
		Ent: ent,
		Exp: b.now().Add(ttl),
	}

	b.trace("trace, bucket set key[%s] tag[%s] exp[%s]\n", key, ent.Tag, h.Exp.Format(time.RFC3339))
	setKeyTotal.Inc()

	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries[key] = h
}

// Get returns entry if entry exists, then expired status
// then existing status.
func (b *Bucket) Get(key string) (Entry, bool, bool) {
	b.mu.Lock()
	h, ok := b.entries[key]
	b.mu.Unlock()

	getKeyTotal.Inc()

	if !ok {
		b.trace("trace, bucket cache miss / key[%s]\n", key)
		cacheMissTotal.Inc()
		return Entry{}, false, false
	}

	if b.now().After(h.Exp) {
		b.trace("trace, bucket cache hit expired key / key[%s] tag[%s]\n", key, h.Ent.Tag)
		cacheHitExpiredTotal.Inc()
		return h.Ent, true, true
	}

	b.trace("trace, bucket cache hit key / key[%s] tag[%s]\n", key, h.Ent.Tag)
	cacheHitTotalNonExpiredTotal.Inc()

	return h.Ent, false, true
}

// Dump dump all existed entries cache to destination file
// path.
func (b *Bucket) Dump() error {
	f, err := os.Create(b.path)
	if err != nil {
		return err
	}

	encode := func() error {
		enc := gob.NewEncoder(f)

		b.mu.Lock()
		defer b.mu.Unlock()

		return enc.Encode(b.entries)
	}

	if err := encode(); err != nil {
		return err
	}

	return f.Close()
}

func (b *Bucket) trace(format string, a ...interface{}) {
	if b.verbose {
		b.l.Debug(format, a...)
	}
}

func (b *Bucket) now() time.Time {
	return time.Now().UTC() // interface?
}

type Entry struct {
	Val []byte
	Tag string
}

type holder struct {
	Ent Entry
	Exp time.Time
}

var (
	setKeyTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_bucket_set_total",
		Help: "The total number of set calls",
	})
	getKeyTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_bucket_get_total",
		Help: "The total number of get calls",
	})
	cacheMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_bucket_cache_miss_total",
		Help: "The total number of cache miss by keys",
	})
	cacheHitExpiredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_bucket_cache_hit_expired_total",
		Help: "The total number of cache hit with expired entries",
	})
	cacheHitTotalNonExpiredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_bucket_cache_hit_non_expired_total",
		Help: "The total number of cahce hit with non expired entries",
	})
)
