package cmd

import (
	"fmt"
	"net/http"
	"time"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

const timeout = 10

type URLHandler struct {
	url      string
	username string
	password string
	token    string
}

func (h *URLHandler) CRDs() ([]*v1.CustomResourceDefinition, error) {
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

	crd := &v1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	return []*v1.CustomResourceDefinition{crd}, nil
}
