package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"net/netip"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rokkerruslan/dnska/internal/endpoints"
	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/internal/resolvers/stub"
	"github.com/rokkerruslan/dnska/internal/udp"
	"github.com/rokkerruslan/dnska/pkg/debug"
)

type Opts struct {
	EndpointsFilePath     string
	DumpDir               string
	BlacklistURL          string
	MetricsHttpServerPort int

	L *slog.Logger
}

type App struct {
	endpoints []endpoints.Endpoint

	dumpDir      string
	blacklistURL string

	metricsHttpServerPort int

	l *slog.Logger
}

func New(opts Opts) (*App, error) {
	a := &App{
		dumpDir:               opts.DumpDir,
		blacklistURL:          opts.BlacklistURL,
		metricsHttpServerPort: opts.MetricsHttpServerPort,
		l:                     opts.L,
	}

	var config endpointsFileConfigurationV0
	if _, err := toml.DecodeFile(opts.EndpointsFilePath, &config); err != nil {
		return nil, err
	}

	endpointsList, err := a.instantiateEndpoints(config)
	if err != nil {
		return nil, err
	}

	a.endpoints = endpointsList

	return a, nil
}

func (a *App) Run(_ context.Context) error {
	if err := a.bootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(a.endpoints))

	for i := range a.endpoints {
		endpoint := a.endpoints[i]

		go func() {
			endpoint.Start(wg.Done)
		}()
	}

	wg.Wait()

	return nil
}

func (a *App) Shutdown() {
	for _, endpoint := range a.endpoints {
		if err := endpoint.Stop(); err != nil {
			a.l.Warn("failed to stop endpoint", "endpoint", endpoint.Name(), "error", err)
		}
	}
}

func (a *App) bootstrap() error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		addr := fmt.Sprintf(":%d", a.metricsHttpServerPort)

		a.l.Info("metrics server starts", "address", addr)

		err := http.ListenAndServe(addr, mux)
		if err != nil {
			a.l.Info("failed to stop server", "error", err)
		}
	}()

	return nil
}

type endpointsFileConfigurationV0 struct {
	LocalAddress string `toml:"local-address"`
}

func (a *App) instantiateEndpoints(efc endpointsFileConfigurationV0) ([]endpoints.Endpoint, error) {
	var resolver resolve.HResolver

	forwardAddrPort := netip.MustParseAddrPort("1.1.1.1:53")

	udpConn, err := udp.NewConn(net.UDPAddrFromAddrPort(forwardAddrPort))
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %v", err)
	}

	mpd := debug.NewFileMalformedPacketDumper(debug.FileMalformedPacketDumperOpts{
		DumpDir: a.dumpDir,
		L:       a.l,
	})

	resolver = resolve.NewCacheResolver(
		resolve.NewBlacklistResolver(resolve.BlacklistResolverOpts{
			AutoReloadInterval: time.Hour,
			BlacklistURL:       a.blacklistURL,
			Next: resolve.NewChainResolver(
				a.l,
				// static.NewStaticResolver(static.Opts{L: a.l}),
				// resolve.NewIterativeResolver(a.l),
				stub.NewSimpleForwardUDPResolver(stub.SimpleForwardUDPResolverOpts{
					UdpConn:               udpConn,
					SetRecursionDesired:   true,
					MalformedPacketDumper: mpd,
				}),
			),
			Lg: a.l,
		}))

	// resolver = resolve.NewAuthorityCleaner(resolver)

	udpLocalAddr, err := netip.ParseAddrPort(efc.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local addr: %v", err)
	}
	// tcpLocalAddr := udpLocalAddr

	endpoints := []endpoints.Endpoint{
		endpoints.NewUDPEndpoint(endpoints.UDPEndpointOpts{
			Addr:                                     udpLocalAddr,
			Resolver:                                 resolver,
			L:                                        a.l,
			WriteErrorReponseToClientIfErrorOccurred: true,
			WriteTimeout:                             500 * time.Millisecond,
			ReadTimeout:                              500 * time.Millisecond,
		}),
		// endpoints.NewTCPEndpoint(tcpLocalAddr, resolver, a.l),
	}

	return endpoints, nil
}
