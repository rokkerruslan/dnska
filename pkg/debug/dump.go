package debug

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"uuid"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	malformedPacketsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_malformed_packets_total",
		Help: "The total number of malformed packets received",
	})

	malformedPacketsTrottledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_malformed_packets_trottled_total",
		Help: "The total number of malformed packets received that were not dumped due to rate limiting",
	})
)

const malformedPacketsFileMode = os.FileMode(0600)

// MalformedPacketDumper defines an interface for dumping malformed packets.
type MalformedPacketDumper interface {
	Dump(buf []byte)
}

type FileMalformedPacketDumperOpts struct {
	DumpDir string
	L       *slog.Logger
}

// FileMalformedPacketDumper is an implementation of MalformedPacketDumper
// that dumps malformed packets to files.
type FileMalformedPacketDumper struct {
	dumpDir string
	limiter *Limiter
	l       *slog.Logger
}

var _ MalformedPacketDumper = (*FileMalformedPacketDumper)(nil)

func NewFileMalformedPacketDumper(opts FileMalformedPacketDumperOpts) *FileMalformedPacketDumper {
	if err := os.MkdirAll(opts.DumpDir, 0755); err != nil {
		panic(err)
	}

	limiter := NewLimiter()

	return &FileMalformedPacketDumper{
		dumpDir: opts.DumpDir,
		limiter: limiter,
		l:       opts.L,
	}
}

func (mpd *FileMalformedPacketDumper) Dump(buf []byte) {
	malformedPacketsTotal.Inc()

	if mpd.limiter.Limit() {
		malformedPacketsTrottledTotal.Inc()
		mpd.l.Info("malformed packet dump skipped due to rate limiting")
		return
	}

	id := uuid.New()

	filename := fmt.Sprintf("%s-%s.dns", time.Now().Format("20060102-150405"), id.String())

	err := os.WriteFile(path.Join(mpd.dumpDir, filename), buf, malformedPacketsFileMode)
	if err != nil {
		mpd.l.Error("failed to dump malformed packet", "error", err)
	} else {
		mpd.l.Info("dumped malformed packet", "id", id.String())
	}
}

// FakeMalformedPacketDumper is a no-op implementation of MalformedPacketDumper.
type FakeMalformedPacketDumper struct{}

func NewFakeMalformedPacketDumper() FakeMalformedPacketDumper {
	return FakeMalformedPacketDumper{}
}

var _ MalformedPacketDumper = FakeMalformedPacketDumper{}

func (mpd FakeMalformedPacketDumper) Dump(buf []byte) {
	// do nothing
}
