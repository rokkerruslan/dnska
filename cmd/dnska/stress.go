package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/resolvers/stub"
	"github.com/rokkerruslan/dnska/internal/udp"
	"github.com/rokkerruslan/dnska/pkg/debug"
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
			// return stress(opts)
			return foo(opts.AttackedAddr, 1)
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

	mpd := debug.NewFakeMalformedPacketDumper()

	for n := 0; n < concurrency; n++ {
		nn := n
		go func() {
			defer wg.Done()
			timer := time.NewTimer(opts.Duration)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			totalRequests := 0
			totalErrors := 0

			udpConn, err := udp.NewConn(net.UDPAddrFromAddrPort(addr))
			if err != nil {
				logger.Error("failed to create UDP connection", "error", err)
				return
			}

			resolver := stub.NewSimpleForwardUDPResolver(stub.SimpleForwardUDPResolverOpts{
				UdpConn:               udpConn,
				SetRecursionDesired:   false,
				MalformedPacketDumper: mpd,
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

				_, err := resolver.Resolve(context.Background(), proto.FromProtoMessage(&in))
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

var dnsTemplate = []byte{
	0x00, 0x00, // [0-1] Transaction ID (заполняется в рантайме)
	0x01, 0x00, // [2-3] Flags: Standard query
	0x00, 0x01, // [4-5] Questions: 1
	0x00, 0x00, // [6-7] Answer RRs: 0
	0x00, 0x00, // [8-9] Authority RRs: 0
	0x00, 0x00, // [10-11] Additional RRs: 0
	0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
	0x03, 'c', 'o', 'm',
	0x00,       // Name terminator
	0x00, 0x01, // Type: A
	0x00, 0x01, // Class: IN
}

func foo(target string, workers int) error {
	raddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		fmt.Printf("Resolution error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting stress test against %s using %d workers...\n", target, workers)

	var totalSent uint64
	stopChan := make(chan struct{})

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			conn, err := net.DialUDP("udp", nil, raddr)
			if err != nil {
				fmt.Printf("Worker %d init error: %v\n", workerID, err)
				return
			}
			defer conn.Close()

			packet := make([]byte, len(dnsTemplate))
			copy(packet, dnsTemplate)

			txID := make([]byte, 2)

			for {
				select {
				case <-stopChan:
					return
				default:
					_, _ = rand.Read(txID)
					packet[0] = txID[0]
					packet[1] = txID[1]

					_, err := conn.Write(packet)
					if err == nil {
						atomic.AddUint64(&totalSent, 1)
					}
				}
			}
		}(i)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	lastSent := uint64(0)

	for {
		select {
		case <-ticker.C:
			currentTotal := atomic.LoadUint64(&totalSent)
			pps := currentTotal - lastSent
			lastSent = currentTotal
			fmt.Printf("Total Packets Sent: %d | Speed: %d pps\n", currentTotal, pps)
		case <-sigChan:
			fmt.Println("\nStopping benchmarks...")
			close(stopChan)
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("Final count: %d packets sent.\n", atomic.LoadUint64(&totalSent))
			return nil
		}
	}
}
