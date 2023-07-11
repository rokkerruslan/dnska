package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/netip"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/resolve"
	"github.com/rokkerruslan/dnska/pkg/proto"
)

type stressOpts struct {
	Duration     time.Duration
	AttackedAddr string
	Concurrency  uint8
}

func NewStressCommand() *cobra.Command {
	var opts stressOpts

	cmd := cobra.Command{
		Use:   "stress",
		Short: "Run stress test for name server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return stress(opts)
		},
	}

	cmd.Flags().DurationVarP(
		&opts.Duration,
		"duration",
		"d",
		10*time.Second,
		"duration of test, if it is 0 the test has not limit of time",
	)

	cmd.Flags().StringVarP(
		&opts.AttackedAddr,
		"attacked-addr",
		"a",
		"",
		"address of a name server",
	)

	cmd.Flags().Uint8VarP(&opts.Concurrency, "concurrency", "c", 1, "total number of simulated requests to server")

	MustMarkFlagRequired(cmd.Flags(), "attacked-addr")

	return &cmd
}

func stress(opts stressOpts) error {
	// Concurrency level.
	// Setup workers based on concurrency level.
	// Prepare list and types of queries.

	concurrency := int(opts.Concurrency)
	if concurrency == 0 {
		concurrency = 1
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)

	addr, err := netip.ParseAddrPort(opts.AttackedAddr)
	if err != nil {
		return err
	}

	for n := 0; n < concurrency; n++ {
		nn := n
		go func() {
			defer wg.Done()
			timer := time.NewTimer(opts.Duration)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			totalRequests := 0
			totalErrors := 0

			resolver := resolve.NewSimpleForwardUDPResolver(resolve.SimpleForwardUDPResolverOpts{
				ForwardAddr:          addr,
				DumpMalformedPackets: true,
				L:                    logger,
			})

		loop:
			for {
				select {
				case <-timer.C:
					timer.Stop()
					break loop
				default:
				}

				totalRequests++

				in := proto.Message{
					Header: proto.Header{
						ID:               1,
						Response:         false,
						RecursionDesired: false,
						QDCount:          1,
					},
					Question: []proto.Question{
						{
							Name:  "lolkek",
							Type:  proto.QTypeA,
							Class: proto.ClassIN,
						},
					},
				}

				_, err := resolver.Resolve(context.Background(), proto.FromProtoMessage(in))
				if err != nil {
					totalErrors++
				}
			}

			fmt.Printf("done :: n=%d total=%d errors=%d\n", nn, totalRequests, totalErrors)
		}()
	}

	wg.Wait()

	return nil
}
