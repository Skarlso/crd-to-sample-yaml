package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
)

const (
	FormatHTML = "html"
	FormatYAML = "yaml"
)

type rootArgs struct {
	fileLocation   string
	folderLocation string
	url            string
	username       string
	password       string
	token          string
	output         string
	format         string
	stdOut         bool
	comments       bool
	minimal        bool
	skipRandom     bool
}

var (
	// generateCmd is root for various `generate ...` commands.
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Simply generate a CRD output.",
		RunE:  runGenerate,
	}

	args = &rootArgs{}
)

type Handler interface {
	CRDs() ([]*v1.CustomResourceDefinition, error)
}

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.PersistentFlags()
	f.StringVarP(&args.fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&args.folderLocation, "folder", "r", "", "A folder from which to parse a series of CRDs.")
	f.StringVarP(&args.url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVar(&args.username, "username", "", "Optional username to authenticate a URL.")
	f.StringVar(&args.password, "password", "", "Optional password to authenticate a URL.")
	f.StringVar(&args.token, "token", "", "A bearer token to authenticate a URL.")
	f.StringVarP(&args.output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
	f.StringVarP(&args.format, "format", "f", FormatYAML, "The format in which to output. Default is YAML. Options are: yaml, html.")
	f.BoolVarP(&args.stdOut, "stdout", "s", false, "If set, it will output the generated content to stdout.")
	f.BoolVarP(&args.comments, "comments", "m", false, "If set, it will add descriptions as comments to each line where available.")
	f.BoolVarP(&args.minimal, "minimal", "l", false, "If set, only the minimal required example yaml is generated.")
	f.BoolVar(&args.skipRandom, "no-random", false, "Skip generating random values that satisfy the property patterns.")
}

func runGenerate(_ *cobra.Command, _ []string) error {
	crdHandler, err := constructHandler(args)
	if err != nil {
		return err
	}

	if args.format == FormatHTML {
		if err := pkg.LoadTemplates(); err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}
	}

	// determine location of output
	if args.output == "" {
		loc, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable location: %w", err)
		}

		args.output = filepath.Dir(loc)
	}

	crds, err := crdHandler.CRDs()
	if err != nil {
		return fmt.Errorf("failed to load CRDs: %w", err)
	}

	var w io.WriteCloser

	var errs []error //nolint:prealloc // nope
	for _, crd := range crds {
		if args.stdOut {
			w = os.Stdout
		} else {
			outputLocation := filepath.Join(args.output, crd.Name+"_sample."+args.format)
			// closed later during render
			outputFile, err := os.Create(outputLocation)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create file at: '%s': %w", outputLocation, err))

				continue
			}

			w = outputFile
		}

		if args.format == FormatHTML {
			errs = append(errs, pkg.RenderContent(w, crd, args.comments, args.minimal))

			continue
		}

		errs = append(errs, pkg.Generate(crd, w, args.comments, args.minimal, args.skipRandom))
	}

	return errors.Join(errs...)
}

func constructHandler(args *rootArgs) (Handler, error) {
	var crdHandler Handler

	switch {
	case args.fileLocation != "":
		crdHandler = &FileHandler{location: args.fileLocation}
	case args.folderLocation != "":
		crdHandler = &FolderHandler{location: args.folderLocation}
	case args.url != "":
		crdHandler = &URLHandler{
			url:      args.url,
			username: args.username,
			password: args.password,
			token:    args.token,
		}
	}

	if crdHandler == nil {
		return nil, errors.New("one of the flags (file, folder, url) must be set")
	}

	return crdHandler, nil
}
