package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
	"github.com/google/go-github/v68/github"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type FetchRepoContent struct {
	URL      string
	Tag      string
	Username string
	Password string
	Token    string
}

// Fetch fetches the code archive of a GitHub repository.
// main archive https://github.com/github/codeql/archive/refs/heads/main.tar.gz
// tagged archive https://github.com/github/codeql/archive/refs/tags/codeql-cli/v2.12.0.zip
func (f *FetchRepoContent) Fetch() ([]*pkg.SchemaType, error) {
	//fetchURL := fmt.Sprintf("%s/archive/refs/heads/main.zip", f.URL)
	//if f.Tag != "" {
	//	fetchURL = fmt.Sprintf("%s/archive/refs/tags/%s.zip", f.URL, f.Tag)
	//}
	//
	//fetch := fetcher.NewFetcher(http.DefaultClient, f.Username, f.Password, f.Token)
	//
	//zipContent, err := fetch.Fetch(fetchURL)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to fetch content from %s: %w", fetchURL, err)
	//}

	client := github.NewClient(http.DefaultClient)
	if f.Token != "" {
		client = client.WithAuthToken(f.Token)
	}

	//client.Repositories.ListReleaseAssets()
	//content, _, err := client.Repositories.DownloadContents(context.Background(), "", "", , &github.RepositoryContentGetOptions{})
	release, _, err := client.Repositories.GetArchiveLink(context.Background(), "Skarlso", "crd-bootstrap", github.Zipball, &github.RepositoryContentGetOptions{
		Ref: f.Tag,
	}, 1)
	if err != nil {
		return nil, err
	}

	//for _, asset := range release.Assets {
	//	fmt.Printf("Downloading asset %s\n", asset.GetName())
	//}
	fetch := fetcher.NewFetcher(http.DefaultClient, f.Username, f.Password, f.Token)
	zipContent, err := fetch.Fetch(release.String())
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return nil, fmt.Errorf("failed to open content from %s: %w", fetch, err)
	}

	var crds []*pkg.SchemaType
	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open content from %s: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("failed to read content from %s: %w", file.Name, err)
		}

		_ = rc.Close()
		sanitized, err := sanitize.Sanitize(content)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize content: %w", err)
		}

		crd := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(sanitized, crd); err != nil {
			return nil, nil //nolint:nilerr // intentional
		}

		schemaType, err := pkg.ExtractSchemaType(crd)
		if err != nil {
			return nil, nil //nolint:nilerr // intentional
		}

		if schemaType != nil {
			crds = append(crds, schemaType)
		}
	}

	return crds, nil
}
