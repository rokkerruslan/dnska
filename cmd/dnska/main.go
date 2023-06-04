package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	opts := slog.HandlerOptions{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &opts))

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

	if err := defaultCmd.Execute(); err != nil {
		logger.Error("failed to execute command", "error", err)
		os.Exit(1)
	}
}
