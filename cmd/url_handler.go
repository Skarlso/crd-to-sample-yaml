package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type URLHandler struct {
	url string
}

func (h *URLHandler) CRDs() ([]*v1beta1.CustomResourceDefinition, error) {
	f := fetcher.NewFetcher(http.DefaultClient)
	content, err := f.Fetch(h.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, errors.New("failed to unmarshal into custom resource definition")
	}

	return []*v1beta1.CustomResourceDefinition{crd}, nil
}
