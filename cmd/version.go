package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "v0.0.0-dev"

// versionCmd is root for various `generate ...` commands.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Return the current version.",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	_, err := fmt.Fprintln(os.Stdout, Version)

	return err
}
