package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "crd-to-sample",
	Short: "CRDs to a Sample file.",
	Run:   ShowUsage,
}

// ShowUsage shows usage of the given command on stdout.
func ShowUsage(cmd *cobra.Command, _ []string) {
	_ = cmd.Usage()
}

// Execute runs the main krok command.
func Execute() error {
	return rootCmd.Execute()
}
