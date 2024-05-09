package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
)

const (
	FormatHTML = "html"
	FormatYAML = "yaml"
)

var (
	// generateCmd is root for various `generate ...` commands.
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Simply generate a CRD output.",
		RunE:  runGenerate,
	}

	fileLocation string
	url          string
	output       string
	format       string
	stdOut       bool
	comments     bool
)

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.PersistentFlags()
	f.StringVarP(&fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVarP(&output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
	f.StringVarP(&format, "format", "f", FormatYAML, "The format in which to output. Default is YAML. Options are: yaml, html.")
	f.BoolVarP(&stdOut, "stdout", "s", false, "If set, it will output the generated content to stdout")
	f.BoolVarP(&comments, "comments", "m", false, "If set, it will add descriptions as comments to each line where available")
}

func runGenerate(_ *cobra.Command, _ []string) error {
	var (
		content []byte
		err     error
		w       io.WriteCloser
	)
	if url != "" {
		f := fetcher.NewFetcher(http.DefaultClient)
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
		return errors.New("failed to unmarshal into custom resource definition")
	}

	if stdOut {
		w = os.Stdout
	} else {
		if output == "" {
			output = filepath.Dir(fileLocation)
		}
		outputLocation := filepath.Join(output, crd.Name+"_sample."+format)
		outputFile, err := os.Create(outputLocation)
		if err != nil {
			return fmt.Errorf("failed to create file at: '%s': %w", outputLocation, err)
		}
		w = outputFile
	}

	if format == FormatHTML {
		if err := pkg.LoadTemplates(); err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}

		return pkg.RenderContent(w, content, comments)
	}

	return pkg.Generate(crd, w, comments)
}
