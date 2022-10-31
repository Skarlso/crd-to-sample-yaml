package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
)

var (
	// generateCmd is root for various `generate ...` commands
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Simply generate a CRD output.",
		RunE:  runGenerate,
	}
	fileLocation string
	url          string
	output       string
	stdOut       bool
)

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.PersistentFlags()
	f.StringVarP(&fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVarP(&output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
	f.BoolVarP(&stdOut, "stdout", "s", false, "If set, it will output the generated content to stdout")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	var (
		content []byte
		err     error
		w       io.WriteCloser
	)
	if url != "" {
		f := NewFetcher(http.DefaultClient)
		content, err = f.Fetch(url)
		if err != nil {
			return fmt.Errorf("failed to fetch content: %w", err)
		}
	} else {
		if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
			return fmt.Errorf("file under '%s' does not exist", fileLocation)
		}
		content, err = os.ReadFile(fileLocation)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition")
	}
	if stdOut {
		w = os.Stdout
	} else {
		if output == "" {
			output = filepath.Dir(fileLocation)
		}
		outputLocation := filepath.Join(output, fmt.Sprintf("%s_sample.yaml", crd.Name))
		outputFile, err := os.Create(outputLocation)
		if err != nil {
			return fmt.Errorf("failed to create file at: '%s': %w", outputLocation, err)
		}
		w = outputFile
	}

	return pkg.Generate(crd, w)
}
