package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

// FolderHandler scans folders and returns schemas found in that folder.
type FolderHandler struct {
	location string
	group    string
}

// CRDs goes through schemas in folders.
func (h *FolderHandler) CRDs() ([]*pkg.SchemaType, error) {
	if _, err := os.Stat(h.location); os.IsNotExist(err) {
		return nil, fmt.Errorf("file under '%s' does not exist", h.location)
	}

	var crds []*pkg.SchemaType

	err := filepath.Walk(h.location, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			_, _ = fmt.Fprintln(os.Stderr, "skipping file "+path)

			return nil
		}

		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		content, err = sanitize.Sanitize(content)
		if err != nil {
			return fmt.Errorf("failed to sanitize content: %w", err)
		}

		crd := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(content, crd); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "skipping none CRD file: "+path)

			return nil //nolint:nilerr // intentional none nil
		}
		schemaType, err := pkg.ExtractSchemaType(crd)
		if err != nil {
			return fmt.Errorf("failed to extract schema type: %w", err)
		}

		if schemaType != nil {
			if h.group != "" {
				schemaType.Rendering = pkg.Rendering{Group: h.group}
			}

			crds = append(crds, schemaType)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk the selected folder: %w", err)
	}

	return crds, nil
}
