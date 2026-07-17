package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/rokkerruslan/dnska/internal/diagnostics"
)

func NewVersionCommand() *cobra.Command {
	var opts struct {
		JSON    bool
		Verbose bool
	}

	cmd := cobra.Command{
		Use:   "version",
		Short: "Print program version and build info",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := diagnostics.CollectInfo()

			if opts.JSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")

				return enc.Encode(info)
			}

			printInfo(os.Stdout, info)

			// The full build info (dependency list and every build setting)
			// is verbose, so it is only shown on demand.
			if opts.Verbose {
				if bi, ok := debug.ReadBuildInfo(); ok {
					fmt.Fprintf(os.Stdout, "\n%s", bi)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "print build info as JSON")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false,
		"also print the full build info (dependencies and settings)")

	return &cmd
}

func printInfo(w io.Writer, i diagnostics.Info) {
	version := i.Version
	if i.Modified {
		version += " (dirty)"
	}

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	fmt.Fprintf(tw, "Version\t%s\n", version)
	fmt.Fprintf(tw, "Go\t%s\n", i.GoVersion)
	fmt.Fprintf(tw, "Revision\t%s\n", i.Revision)
	fmt.Fprintf(tw, "Committed\t%s\n", i.Committed)
	fmt.Fprintf(tw, "Platform\t%s/%s\n", i.OS, i.Arch)
	fmt.Fprintf(tw, "Built by\t%s@%s\n", i.BuildUser, i.BuildHost)
	fmt.Fprintf(tw, "Machine ID\t%s\n", i.BuildMachineID)
	fmt.Fprintf(tw, "Uptime\t%s\n", i.Uptime)

	_ = tw.Flush()
}
