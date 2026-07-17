package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/resolvers/stub"
	"github.com/rokkerruslan/dnska/internal/udp"
	"github.com/rokkerruslan/dnska/pkg/debug"
	"github.com/rokkerruslan/dnska/pkg/format"
	"github.com/rokkerruslan/dnska/pkg/proto"
	"github.com/rokkerruslan/dnska/pkg/query"
)

func NewLookupCommand(l *slog.Logger) *cobra.Command {
	var opts struct {
		Type                    string
		Class                   uint16
		Addr                    string
		Stub                    bool
		OnlyAnswer              bool
		SetRecursionDesiredFlag bool
		DumpMalformedPackets    bool
	}

	cmd := cobra.Command{
		Use:          "lookup [NAME]",
		Short:        "Use resolver",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			addr, err := netip.ParseAddrPort(opts.Addr)
			if err != nil {
				return err
			}

			// todo: A configurable deadline?
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			in := query.AddQuestion(query.NewTemplate(), name, proto.ParseQType(opts.Type), proto.QClass(opts.Class))

			udpConn, err := udp.NewConn(net.UDPAddrFromAddrPort(addr))
			if err != nil {
				return err
			}

			mpd := debug.NewFileMalformedPacketDumper(debug.FileMalformedPacketDumperOpts{
				DumpDir: "./dumps",
				L:       l,
			})

			resolver := stub.NewSimpleForwardUDPResolver(stub.SimpleForwardUDPResolverOpts{
				UdpConn:               udpConn,
				MalformedPacketDumper: mpd,
				SetRecursionDesired:   opts.SetRecursionDesiredFlag,
			})

			in2 := proto.FromProtoMessage(&in)

			message2, err := resolver.Resolve(ctx, in2)
			if err != nil {
				return err
			}

			message := message2.ToProtoMessage()

			if opts.OnlyAnswer {
				fmt.Print(format.FormatDNSAnswer(message.Answer))
			} else {
				fmt.Print(format.FormatDNSMessage(message, name))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", proto.QTypeA.String(), "record type")
	cmd.Flags().Uint16VarP(&opts.Class, "class", "c", uint16(proto.ClassIN), "record class")

	cmd.Flags().StringVarP(&opts.Addr, "addr", "a", "1.1.1.1:53",
		"address of a name server that will be used")

	cmd.Flags().BoolVar(&opts.OnlyAnswer, "only-answer", false, "display only answer part of response")

	cmd.Flags().BoolVarP(&opts.SetRecursionDesiredFlag, "recursion-desired", "r", false,
		"set to 1 the recursion desired bit flag in request message")

	cmd.Flags().BoolVarP(&opts.Stub, "stub", "s", true, "stub mode")

	return &cmd
}
