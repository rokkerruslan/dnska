package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

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
