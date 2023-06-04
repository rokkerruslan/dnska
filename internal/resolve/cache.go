package resolve

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

type RecordsCache struct {
	mu    sync.Mutex
	cache map[proto.Question]cacheline
}

func (rc *RecordsCache) Get(q proto.Question) ([]proto.ResourceRecord, bool, bool) {
	rc.mu.Lock()
	line, ok := rc.cache[q]
	rc.mu.Unlock()

	if !ok {
		return nil, false, false
	}

	return line.list, line.ttl.Before(time.Now().UTC()), true
}

func (rc *RecordsCache) Put(q proto.Question, list []proto.ResourceRecord) {
	if len(list) == 0 {
		return
	}

	mi := uint32(math.MaxUint32)
	for i := range list {
		if list[i].TTL < mi {
			mi = list[i].TTL
		}
	}

	ttl := time.Now().UTC().Add(time.Second * time.Duration(mi))

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache[q] = cacheline{ttl: ttl, list: list}
}

type cacheline struct {
	ttl  time.Time
	list []proto.ResourceRecord
}

type CacheResolver struct {
	sub   Resolver
	cache RecordsCache
}

func NewCacheResolver(sub Resolver) *CacheResolver {
	return &CacheResolver{
		sub: sub,
		cache: RecordsCache{
			mu:    sync.Mutex{},
			cache: map[proto.Question]cacheline{},
		},
	}
}

func (cr *CacheResolver) Resolve(
	ctx context.Context, in *proto.InternalMessage,
) (
	*proto.InternalMessage,
	error,
) {
	records, expired, ok := cr.cache.Get(in.Question)
	if !ok || expired {
		out, err := cr.sub.Resolve(ctx, in)
		if err != nil {
			return nil, err
		}

		// todo: cache authority and additional info too

		cr.cache.Put(in.Question, out.Answer)
		in.Answer = out.Answer
	} else {
		in.Answer = records
	}

	return in, nil
}
