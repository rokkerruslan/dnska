package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func NewEncodeCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "encode",
		Short: "Construct and encodes DNS message",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("not implemented")
		},
	}

	return &cmd
}
