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
	url         string
	username    string
	password    string
	token       string
	tag         string
	caBundle    string
	privSSHKey  string
	useSSHAgent bool
}

func (g *GitHandler) CRDs() ([]*pkg.SchemaType, error) {
	opts, err := g.constructGitOptions()
	if err != nil {
		return nil, err
	}

	r, err := git.Clone(memory.NewStorage(), nil, opts)
	if err != nil {
		return nil, err
	}

	var ref *plumbing.Reference
	if g.tag != "" {
		ref, err = r.Tag(g.tag)
	} else {
		ref, err = r.Head()
	}
	if err != nil {
		return nil, err
	}

	crds, err := gatherSchemaTypesForRef(r, ref)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintln(os.Stderr, "Discovered number of CRDs: ", len(crds))

	return crds, nil
}

func gatherSchemaTypesForRef(r *git.Repository, ref *plumbing.Reference) ([]*pkg.SchemaType, error) {
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, err
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
			_, _ = fmt.Fprintln(os.Stderr, "skipping none CRD file: "+crd.GetName())

			return nil //nolint:nilerr // intentional
		}

		if schemaType != nil {
			crds = append(crds, schemaType)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return crds, nil
}

func (g *GitHandler) constructGitOptions() (*git.CloneOptions, error) {
	opts := &git.CloneOptions{
		URL:   g.url,
		Depth: 1,
	}

	// trickle down. if ssh key is set, this will be overwritten.
	if g.username != "" && g.password != "" {
		opts.Auth = &http.BasicAuth{
			Username: g.username,
			Password: g.password,
		}
	}
	if g.token != "" {
		opts.Auth = &http.TokenAuth{
			Token: g.token,
		}
	}
	if g.caBundle != "" {
		opts.CABundle = []byte(g.caBundle)
	}
	if g.privSSHKey != "" {
		if !strings.Contains(g.url, "@") {
			return nil, fmt.Errorf("git URL does not contain an ssh address: %s", g.url)
		}

		keys, err := ssh.NewPublicKeysFromFile("git", g.privSSHKey, g.password)
		if err != nil {
			return nil, err
		}

		opts.Auth = keys
	}
	if g.useSSHAgent {
		if !strings.Contains(g.url, "@") {
			return nil, fmt.Errorf("git URL does not contain an ssh address: %s", g.url)
		}

		authMethod, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, err
		}
		opts.Auth = authMethod
	}

	return opts, nil
}
