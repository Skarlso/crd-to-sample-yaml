package matchsnapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
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
	sourceTemplate, err := os.ReadFile(sourceTemplateLocation)
	if err != nil {
		return err
	}
	baseName := strings.Trim(filepath.Base(sourceTemplateLocation), filepath.Ext(sourceTemplateLocation))

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(sourceTemplate, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	for _, version := range crd.Spec.Versions {
		name := baseName + "-" + version.Name + ".yaml"
		if minimal {
			name = baseName + "-" + version.Name + ".min.yaml"
		}
		file, err := os.OpenFile(filepath.Join(targetSnapshotLocation, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filepath.Join(targetSnapshotLocation, name), err)
		}

		defer file.Close()

		parser := pkg.NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, minimal)
		if err := parser.ParseProperties(version.Name, file, version.Schema.OpenAPIV3Schema.Properties, pkg.RootRequiredFields); err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
		}
	}

	return nil
}
