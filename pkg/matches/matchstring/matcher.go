package matchstring

import (
	"context"
	"fmt"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

type Matcher struct{}

func (m *Matcher) Match(_ context.Context, crdLocation string, payload []byte) error {
	c := &apiextensionsv1.JSON{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	crdContent, err := os.ReadFile(crdLocation)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", crdLocation, err)
	}

	return matches.Validate(crdContent, payload)
}

func init() {
	tests.Register(&Matcher{}, "matchString")
}
