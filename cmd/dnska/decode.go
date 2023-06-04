package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

func NewDecodeCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "decode [PATH]",
		Short: "Try to decode DNS packet from file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			buf, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			dec := proto.NewDecoder()

			msg, err := dec.Decode(buf)
			if err != nil {
				return fmt.Errorf("failed to decode msg :: error=%v", err)
			}

			spew.Dump(msg)

			return nil
		},
	}

	return &cmd
}
