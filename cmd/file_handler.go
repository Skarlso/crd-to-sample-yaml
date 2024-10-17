package cmd

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

type FileHandler struct {
	location string
}

func (h *FileHandler) CRDs() ([]*pkg.SchemaType, error) {
	if _, err := os.Stat(h.location); os.IsNotExist(err) {
		return nil, fmt.Errorf("file under '%s' does not exist", h.location)
	}
	content, err := os.ReadFile(h.location)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
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

	return []*pkg.SchemaType{schemaType}, nil
}
