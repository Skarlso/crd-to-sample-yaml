package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type discoverArgs struct {
	url      string
	username string
	password string
	token    string
	provider string
	tag      string
}

var (
	// discoverCmd is root for various `generate ...` commands.
	discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "Point to a git repository to discover and display all CRD contents.",
		RunE:  runDiscovery,
	}

	discoverArg = &discoverArgs{}
)

func init() {
	rootCmd.AddCommand(discoverCmd)
	// using persistent flags so all flags will be available for all sub commands.
	f := discoverCmd.PersistentFlags()
	f.StringVarP(&discoverArg.url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVar(&discoverArg.username, "username", "", "Optional username to authenticate a URL.")
	f.StringVar(&discoverArg.password, "password", "", "Optional password to authenticate a URL.")
	f.StringVar(&discoverArg.token, "token", "", "A bearer token to authenticate a URL.")
	f.StringVar(&discoverArg.tag, "tag", "", "The ref to check out.")
	//f.StringVar(&discoverArg.provider, "provider", "github", "What provider to use, GitHub is default.")
}

func runDiscovery(_ *cobra.Command, _ []string) error {
	// TODO: Check if we can write to disk, if yes, use that.
	// git.PlainClone...
	// Tags should be settable
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:   discoverArg.url,
		Depth: 1,
		Auth: &http.BasicAuth{
			Username: discoverArg.username,
			Password: discoverArg.password,
		},
	})
	if err != nil {
		return err
	}

	var ref *plumbing.Reference
	if discoverArg.tag != "" {
		ref, err = r.Tag(discoverArg.tag)
	} else {
		ref, err = r.Head()
	}
	if err != nil {
		return err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return err
	}

	var crds []*pkg.SchemaType

	if err := commitTree.Files().ForEach(func(f *object.File) error {
		if ext := filepath.Ext(f.Name); ext != ".yaml" {
			return nil
		}

		content, err := f.Contents()
		if err != nil {
			return err
		}

		sanitized, err := sanitize.Sanitize([]byte(content))
		if err != nil {
			return fmt.Errorf("failed to sanitize content: %w", err)
		}

		crd := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(sanitized, crd); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "skipping none CRD file: "+f.Name)

			return nil //nolint:nilerr // intentional
		}
		schemaType, err := pkg.ExtractSchemaType(crd)
		if err != nil {
			// skip a failing crd...
			_, _ = fmt.Fprintln(os.Stderr, "skipping none CRD file: "+crd.GetName())
			return nil
		}

		if schemaType != nil {
			crds = append(crds, schemaType)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
