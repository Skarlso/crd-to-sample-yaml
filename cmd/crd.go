package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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

type crdGenArgs struct {
	comments   bool
	minimal    bool
	skipRandom bool
	output     string
	format     string
	stdOut     bool
}

var crdArgs = &crdGenArgs{}

type Handler interface {
	CRDs() ([]*pkg.SchemaType, error)
}

func init() {
	generateCmd.AddCommand(crdCmd)
	f := crdCmd.PersistentFlags()
	f.BoolVarP(&crdArgs.comments, "comments", "m", false, "If set, it will add descriptions as comments to each line where available.")
	f.BoolVarP(&crdArgs.minimal, "minimal", "l", false, "If set, only the minimal required example yaml is generated.")
	f.BoolVar(&crdArgs.skipRandom, "no-random", false, "Skip generating random values that satisfy the property patterns.")
	f.StringVarP(&crdArgs.output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
	f.StringVarP(&crdArgs.format, "format", "f", FormatYAML, "The format in which to output. Default is YAML. Options are: yaml, html.")
	f.BoolVarP(&crdArgs.stdOut, "stdout", "s", false, "If set, it will output the generated content to stdout.")
}

func runGenerate(_ *cobra.Command, _ []string) error {
	crdHandler, err := constructHandler(args)
	if err != nil {
		return err
	}

	if crdArgs.format == FormatHTML {
		if crdArgs.output == "" {
			return errors.New("output must be set to a filename if format is HTML")
		}

		if err := pkg.LoadTemplates(); err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}
	}

	// determine location of output
	if crdArgs.output == "" {
		loc, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable location: %w", err)
		}

		crdArgs.output = filepath.Dir(loc)
	}

	crds, err := crdHandler.CRDs()
	if err != nil {
		return fmt.Errorf("failed to load CRDs: %w", err)
	}

	var w io.WriteCloser
	if crdArgs.format == FormatHTML {
		if crdArgs.stdOut {
			w = os.Stdout
		} else {
			w, err = os.Create(crdArgs.output)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
		}

		opts := pkg.RenderOpts{
			Comments: crdArgs.comments,
			Minimal:  crdArgs.minimal,
			Random:   crdArgs.skipRandom,
		}

		return pkg.RenderContent(w, crds, opts)
	}

	var errs []error //nolint:prealloc // nope
	for _, crd := range crds {
		if crdArgs.stdOut {
			w = os.Stdout
		} else {
			outputLocation := filepath.Join(crdArgs.output, crd.Kind+"_sample."+crdArgs.format)
			// closed later during render
			outputFile, err := os.Create(outputLocation)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create file at: '%s': %w", outputLocation, err))

				continue
			}

			w = outputFile
		}

		errs = append(errs, pkg.Generate(crd, w, crdArgs.comments, crdArgs.minimal, crdArgs.skipRandom))
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
	case args.configFileLocation != "":
		crdHandler = &ConfigHandler{configFileLocation: args.configFileLocation}
	case args.gitAccess:
		crdHandler = &GitHandler{
			URL:         args.url,
			Username:    args.username,
			Password:    args.password,
			Token:       args.token,
			Tag:         args.tag,
			caBundle:    args.caBundle,
			privSSHKey:  args.privSSHKey,
			useSSHAgent: args.useSSHAgent,
		}
	case args.url != "":
		crdHandler = &URLHandler{
			url:      args.url,
			username: args.username,
			password: args.password,
			token:    args.token,
		}
	}

	if crdHandler == nil {
		return nil, errors.New("one of the flags (file, folder, url, configFile) must be set")
	}

	return crdHandler, nil
}
