package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	opts := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))

	defaultCmd := cobra.Command{
		Use:   "dnska",
		Short: "Toy DNS implementation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	defaultCmd.AddCommand(
		NewDecodeCommand(),
		NewEncodeCommand(),
		NewLookupCommand(logger),
		NewAppCommand(logger),
		NewStressCommand(),
		NewVersionCommand(),
	)

	_ = defaultCmd.Execute()
}
