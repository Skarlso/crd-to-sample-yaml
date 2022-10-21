package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	output       string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.PersistentFlags()
	f.StringVarP(&fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		return fmt.Errorf("file under '%s' does not exist", fileLocation)
	}

	content, err := os.ReadFile(fileLocation)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if output == "" {
		output = filepath.Dir(fileLocation)
	}

	return pkg.Generate(content, output)
}
