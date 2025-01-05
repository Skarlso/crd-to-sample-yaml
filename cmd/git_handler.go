package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

type GitHandler struct {
	URL      string
	Username string
	Password string
	Token    string
	Tag      string

	caBundle    string
	privSSHKey  string
	useSSHAgent bool
	group       string // this is used by the configfile.
}

func (g *GitHandler) CRDs() ([]*pkg.SchemaType, error) {
	opts, err := g.constructGitOptions()
	if err != nil {
		return nil, err
	}

	r, err := git.Clone(memory.NewStorage(), nil, opts)
	if err != nil {
		return nil, fmt.Errorf("error cloning git repository: %w", err)
	}

	var ref *plumbing.Reference
	if g.Tag != "" {
		ref, err = r.Tag(g.Tag)
	} else {
		ref, err = r.Head()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to construct reference: %w", err)
	}

	crds, err := g.gatherSchemaTypesForRef(r, ref)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintln(os.Stderr, "Discovered number of CRDs: ", len(crds))

	return crds, nil
}

func (g *GitHandler) gatherSchemaTypesForRef(r *git.Repository, ref *plumbing.Reference) ([]*pkg.SchemaType, error) {
	// Need to resolve the ref first to the right hash otherwise it's not found.
	hash, err := r.ResolveRevision(plumbing.Revision(ref.Hash().String()))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve revision: %w", err)
	}

	commit, err := r.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("error getting commit object: %w", err)
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var crds []*pkg.SchemaType
	// Tried to make this concurrent, but there was very little gain. It just takes this long to
	// clone a large repository. It's not the processing OR the rendering that takes long.
	if err := commitTree.Files().ForEach(func(f *object.File) error {
		crd, err := g.processEntry(f)
		if err != nil {
			return err
		}

		if crd != nil {
			crds = append(crds, crd)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return crds, nil
}

func (g *GitHandler) processEntry(f *object.File) (*pkg.SchemaType, error) {
	for _, path := range strings.Split(f.Name, string(filepath.Separator)) {
		if path == "test" {
			return nil, nil
		}
	}

	if ext := filepath.Ext(f.Name); ext != ".yaml" {
		return nil, nil
	}

	content, err := f.Contents()
	if err != nil {
		return nil, err
	}

	sanitized, err := sanitize.Sanitize([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize content: %w", err)
	}

	crd := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(sanitized, crd); err != nil {
		return nil, nil //nolint:nilerr // intentional
	}

	schemaType, err := pkg.ExtractSchemaType(crd)
	if err != nil || schemaType == nil {
		return nil, nil //nolint:nilerr // intentional
	}

	if g.group != "" {
		schemaType.Rendering = pkg.Rendering{Group: g.group}
	}

	return schemaType, nil
}

func (g *GitHandler) constructGitOptions() (*git.CloneOptions, error) {
	opts := &git.CloneOptions{
		URL:   g.URL,
		Depth: 1,
	}

	// trickle down. if ssh key is set, this will be overwritten.
	if g.Username != "" && g.Password != "" {
		opts.Auth = &http.BasicAuth{
			Username: g.Username,
			Password: g.Password,
		}
	}
	if g.Token != "" {
		opts.Auth = &http.TokenAuth{
			Token: g.Token,
		}
	}
	if g.caBundle != "" {
		opts.CABundle = []byte(g.caBundle)
	}
	if g.privSSHKey != "" {
		if !strings.Contains(g.URL, "@") {
			return nil, fmt.Errorf("git URL does not contain an ssh address: %s", g.URL)
		}

		keys, err := ssh.NewPublicKeysFromFile("git", g.privSSHKey, g.Password)
		if err != nil {
			return nil, err
		}

		opts.Auth = keys
	}
	if g.useSSHAgent {
		if !strings.Contains(g.URL, "@") {
			return nil, fmt.Errorf("git URL does not contain an ssh address: %s", g.URL)
		}

		authMethod, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, err
		}
		opts.Auth = authMethod
	}

	return opts, nil
}
