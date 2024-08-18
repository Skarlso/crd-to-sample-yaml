package tests

import (
	"fmt"
	"os"

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
func (u *Update) Update(sourceTemplateLocation string, targetSnapshot string, minimal bool) error {
	content, err := os.ReadFile(sourceTemplateLocation)
	if err != nil {
		return fmt.Errorf("could not read source template: %w", err)
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	file, err := os.OpenFile(targetSnapshot, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", targetSnapshot, err)
	}

	defer file.Close()

	if err := pkg.Generate(crd, file, false, minimal); err != nil {
		return err
	}

	return nil
}
