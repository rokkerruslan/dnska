package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/diagnostics"
)

func NewVersionCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Print program version and build info",
		Run: func(_ *cobra.Command, _ []string) {
			spew.Dump(diagnostics.CollectInfo())
		},
	}

	return &cmd
}
