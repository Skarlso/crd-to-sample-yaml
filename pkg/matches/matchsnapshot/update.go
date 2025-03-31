package matchsnapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

const (
	perm = 0o644
)

type Updater interface {
	Update(sourceTemplateLocation string, targetSnapshot string, minimal bool) error
}

type Update struct{}

// Update any given files in the snapshots.
func (u *Update) Update(sourceTemplateLocation string, targetSnapshotLocation string, minimal bool) error {
	sourceTemplate, err := os.ReadFile(filepath.Clean(sourceTemplateLocation))
	if err != nil {
		return err
	}
	baseName := strings.Trim(filepath.Base(sourceTemplateLocation), filepath.Ext(sourceTemplateLocation))

	crd := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(sourceTemplate, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}
	schemaType, err := pkg.ExtractSchemaType(crd)
	if err != nil {
		return fmt.Errorf("failed to extract schema type: %w", err)
	}

	for _, version := range schemaType.Versions {
		name := baseName + "-" + version.Name + ".yaml"
		if minimal {
			name = baseName + "-" + version.Name + ".min.yaml"
		}
		file, err := os.OpenFile(filepath.Clean(filepath.Join(targetSnapshotLocation, name)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filepath.Clean(filepath.Join(targetSnapshotLocation, name)), err)
		}

		parser := pkg.NewParser(schemaType.Group, schemaType.Kind, false, minimal, true)
		if err := parser.ParseProperties(version.Name, file, version.Schema.Properties, pkg.RootRequiredFields); err != nil {
			_ = file.Close()

			return fmt.Errorf("failed to parse properties: %w", err)
		}

		_ = file.Close()
	}

	if len(schemaType.Versions) == 0 && schemaType.Validation != nil {
		name := baseName + "-" + schemaType.Validation.Name + ".yaml"
		if minimal {
			name = baseName + "-" + schemaType.Validation.Name + ".min.yaml"
		}
		file, err := os.OpenFile(filepath.Clean(filepath.Join(targetSnapshotLocation, name)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filepath.Clean(filepath.Join(targetSnapshotLocation, name)), err)
		}

		defer func() {
			_ = file.Close()
		}()

		schemaType.Validation.Schema.Properties["kind"] = v1beta1.JSONSchemaProps{}
		schemaType.Validation.Schema.Properties["apiVersion"] = v1beta1.JSONSchemaProps{}
		parser := pkg.NewParser(schemaType.Group, schemaType.Kind, false, minimal, false)
		if err := parser.ParseProperties(schemaType.Validation.Name, file, schemaType.Validation.Schema.Properties, pkg.RootRequiredFields); err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
		}
	}

	return nil
}
