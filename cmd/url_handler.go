package cmd

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

const timeout = 10

type URLHandler struct {
	url      string
	username string
	password string
	token    string
	group    string
}

func (h *URLHandler) CRDs() ([]*pkg.SchemaType, error) {
	client := http.DefaultClient
	client.Timeout = timeout * time.Second

	f := fetcher.NewFetcher(client, h.username, h.password, h.token)

	content, err := f.Fetch(h.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	content, err = sanitize.Sanitize(content)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize content: %w", err)
	}

	crd := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	schemaType, err := pkg.ExtractSchemaType(crd)
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema type: %w", err)
	}

	if schemaType == nil {
		return nil, nil
	}

	if h.group != "" {
		schemaType.Rendering = pkg.Rendering{Group: h.group}
	}

	return []*pkg.SchemaType{schemaType}, nil
}
