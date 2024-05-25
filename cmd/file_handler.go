package cmd

import (
	"fmt"
	"os"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type FileHandler struct {
	location string
}

func (h *FileHandler) CRDs() ([]*v1beta1.CustomResourceDefinition, error) {
	if _, err := os.Stat(h.location); os.IsNotExist(err) {
		return nil, fmt.Errorf("file under '%s' does not exist", h.location)
	}
	content, err := os.ReadFile(h.location)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	return []*v1beta1.CustomResourceDefinition{crd}, nil
}
