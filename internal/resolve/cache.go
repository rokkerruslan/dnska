package resolve

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/rokkerruslan/dnska/pkg/bucket"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

type cacheResolver struct {
	sub    Resolver
	bucket *bucket.Bucket
}

func (c *cacheResolver) Resolve(ctx context.Context, in proto.Message) (proto.Message, error) {
	if len(in.Question) != 1 {
		return proto.Message{}, errors.New("cache resolver currency not support multi-question requests")
	}

	q := in.Question[0]

	// todo: Calculate hash from q.Name and another fields

	entry, expired, exists := c.bucket.Get(q.Name)
	if exists && !expired {
		dec := proto.NewDecoder()

		decoded, err := dec.Decode(entry.Val)
		if err != nil {
			return proto.Message{}, err
		}

		// todo: id?
		decoded.Header.ID = in.Header.ID

		return decoded, nil
	}

	out, err := c.sub.Resolve(ctx, in)
	if err != nil {
		return proto.Message{}, err
	}

	if out.Header.RCode == proto.RCodeNoErrorCondition {
		enc := proto.NewEncoder(make([]byte, 512))

		buf, err := enc.Encode(out)
		if err == nil {

			entry := bucket.Entry{
				Val: buf,
				Tag: "",
			}

			c.bucket.Set(q.Name, entry, 10*time.Second)
		}
	}

	return out, nil
}

func NewCacheResolver(sub Resolver) Resolver {
	return &cacheResolver{
		sub: sub,
		bucket: bucket.New(bucket.Opts{
			Path:    "/tmp/resolve-cache",
			Verbose: false,
			L:       zerolog.New(os.Stdout),
		}),
	}
}
