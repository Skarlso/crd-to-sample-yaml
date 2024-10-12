package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
)

const (
	FormatHTML = "html"
	FormatYAML = "yaml"
)

// crdCmd is the command that generates CRD output.
var crdCmd = &cobra.Command{
	Use:   "crd",
	Short: "Simply generate a CRD output.",
	RunE:  runGenerate,
}

type Handler interface {
	CRDs() ([]*v1beta1.CustomResourceDefinition, error)
}

func init() {
	generateCmd.AddCommand(crdCmd)
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

	if args.format == FormatHTML {
		if args.stdOut {
			w = os.Stdout
		} else {
			w, err = os.Create(args.output)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}

			defer w.Close()
		}

		return pkg.RenderContent(w, crds, args.comments, args.minimal)
	}

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
