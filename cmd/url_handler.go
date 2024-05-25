package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type URLHandler struct {
	url string
}

func (h *URLHandler) CRDs() ([]*v1beta1.CustomResourceDefinition, error) {
	client := http.DefaultClient
	client.Timeout = 10 * time.Second

	f := fetcher.NewFetcher(client)
	content, err := f.Fetch(h.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	return []*v1beta1.CustomResourceDefinition{crd}, nil
}
