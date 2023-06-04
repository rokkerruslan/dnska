package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/netip"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	endpoints2 "github.com/rokkerruslan/dnska/internal/endpoints"
	resolve2 "github.com/rokkerruslan/dnska/internal/resolve"
)

type Opts struct {
	EndpointsFilePath string

	L zerolog.Logger
}

type App struct {
	endpoints []endpoints2.Endpoint

	l zerolog.Logger
}

func New(opts Opts) (*App, error) {
	logger := opts.L

	endpointsList, err := setup(logger, opts.EndpointsFilePath)
	if err != nil {
		return nil, err
	}

	return &App{
		endpoints: endpointsList,
		l:         logger,
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
			a.l.Printf("failed to stop endpoint %s :: error=%s", endpoint.Name(), err)
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
			a.l.Printf("listen and serve error: %v", err)
		}
	}()

	return nil
}

type endpointsFileConfigurationV0 struct {
	LocalAddress string `toml:"local-address"`
}

func (efc endpointsFileConfigurationV0) InstantiateEndpoints(l zerolog.Logger) ([]endpoints2.Endpoint, error) {
	resolver := resolve2.NewCacheResolver(
		resolve2.NewBlacklistResolver(resolve2.BlacklistResolverOpts{
			AutoReloadInterval: time.Hour,
			BlacklistURL:       "http://github.com/black",
			Pass: resolve2.NewChainResolver(
				l,
				resolve2.NewStaticResolver(l),
				resolve2.NewIterativeResolver(l)),
		}))

	udpLocalAddr, err := netip.ParseAddrPort(efc.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local addr: %v", err)
	}

	return []endpoints2.Endpoint{endpoints2.NewUDPEndpoint(udpLocalAddr, resolver, l)}, nil
}

func setup(l zerolog.Logger, endpointsFilePath string) ([]endpoints2.Endpoint, error) {
	var config endpointsFileConfigurationV0
	if _, err := toml.DecodeFile(endpointsFilePath, &config); err != nil {
		return nil, err
	}

	return config.InstantiateEndpoints(l)
}
