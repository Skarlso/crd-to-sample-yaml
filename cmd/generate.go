package cmd

import (
	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/spf13/cobra"
)

var (
	// generateCmd is root for various `generate ...` commands
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Simply generate a CRD output.",
		RunE:  runGenerate,
	}
	fileLocation string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.PersistentFlags()
	f.StringVarP(&fileLocation, "location", "f", "", "The CRD file to generate a yaml from.")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	return pkg.Generate(fileLocation)
}
