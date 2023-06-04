package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MustMarkFlagRequired call the cobra.MarkFlagRequired function and
// panics if error is not equal nil.
func MustMarkFlagRequired(flags *pflag.FlagSet, name string) {
	if err := cobra.MarkFlagRequired(flags, name); err != nil {
		panic(err)
	}
}
