package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/resolve"
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

			// todo: do not ignore opts.SetRecursionDesiredFlag
			if opts.SetRecursionDesiredFlag {
				in.Header.RecursionDesired = true
			}

			resolver := resolve.NewSimpleForwardUDPResolver(resolve.SimpleForwardUDPResolverOpts{
				ForwardAddr:          addr,
				DumpMalformedPackets: opts.DumpMalformedPackets,
				L:                    l,
			})

			in2 := proto.FromProtoMessage(in)

			message2, err := resolver.Resolve(ctx, in2)
			if err != nil {
				return err
			}

			message := message2.ToProtoMessage()

			if err != nil {
				return fmt.Errorf("failed to lookup :: name=%s error=%v", name, err)
			}

			if opts.OnlyAnswer {
				spew.Dump(message.Answer)
			} else {
				spew.Dump(message)
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
