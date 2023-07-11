package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"net/netip"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rokkerruslan/dnska/internal/endpoints"
	"github.com/rokkerruslan/dnska/internal/resolve"
)

type Opts struct {
	EndpointsFilePath string

	L *slog.Logger
}

type App struct {
	endpoints []endpoints.Endpoint

	l *slog.Logger
}

func New(opts Opts) (*App, error) {
	endpointsList, err := setup(opts.L, opts.EndpointsFilePath)
	if err != nil {
		return nil, err
	}

	return &App{
		endpoints: endpointsList,
		l:         opts.L,
	}, nil
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
		err := http.ListenAndServe(":8888", mux)
		if err != nil {
			a.l.Info("failed to stop server", "error", err)
		}
	}()

	return nil
}

type endpointsFileConfigurationV0 struct {
	LocalAddress string `toml:"local-address"`
}

func (efc endpointsFileConfigurationV0) InstantiateEndpoints(l *slog.Logger) ([]endpoints.Endpoint, error) {

	var resolver resolve.Resolver

	resolver = resolve.NewCacheResolver(
		resolve.NewBlacklistResolver(resolve.BlacklistResolverOpts{
			AutoReloadInterval: time.Hour,
			BlacklistURL:       "http://github.com/black",
			Pass: resolve.NewChainResolver(
				l,
				resolve.NewStaticResolver(l),
				resolve.NewIterativeResolver(l),
				resolve.NewAdvancedForwardUDPResolver(resolve.AdvancedForwardUDPResolverOpts{
					UpstreamAddrPort:     endpoints.DefaultCloudflareAddrPort,
					DumpMalformedPackets: true,
					L:                    l,
				}),
			),
		}))

	resolver = resolve.NewAuthorityCleaner(resolver)

	udpLocalAddr, err := netip.ParseAddrPort(efc.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local addr: %v", err)
	}
	tcpLocalAddr := udpLocalAddr

	endpoints := []endpoints.Endpoint{
		endpoints.NewUDPEndpoint(udpLocalAddr, resolver, l),
		endpoints.NewTCPEndpoint(tcpLocalAddr, resolver, l),
	}

	return endpoints, nil
}

func setup(l *slog.Logger, endpointsFilePath string) ([]endpoints.Endpoint, error) {
	var config endpointsFileConfigurationV0
	if _, err := toml.DecodeFile(endpointsFilePath, &config); err != nil {
		return nil, err
	}

	return config.InstantiateEndpoints(l)
}
