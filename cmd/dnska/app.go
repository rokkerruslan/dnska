package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/app"
)

func NewAppCommand(l zerolog.Logger) *cobra.Command {
	var opts struct {
		EndpointsFilePath string
	}

	cmd := cobra.Command{
		Use:   "app",
		Short: "Run DNS server application",
		RunE: func(cmd *cobra.Command, args []string) error {
			signals := make(chan os.Signal, 1)
			defer close(signals)
			defer signal.Stop(signals)

			signal.Notify(signals, os.Interrupt)
			defer signal.Reset(os.Interrupt)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			application, err := app.New(app.Opts{
				EndpointsFilePath: opts.EndpointsFilePath,
				L:                 l,
			})
			if err != nil {
				return err
			}

			go func() {
				defer cancel()
				defer application.Shutdown()

				<-signals
			}()

			if err := application.Run(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.EndpointsFilePath, "endpoints-file-path", "./configs/endpoints.example.toml",
		"path to endpoints configuration on OS filesystem")

	return &cmd
}
