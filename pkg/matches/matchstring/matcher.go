package matchstring

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

type Matcher struct{}

type Config struct {
	IgnoreErrors []string `yaml:"ignoreErrors,omitempty"`
}

func (m *Matcher) Match(_ context.Context, crdLocation string, payload []byte) error {
	c := &Config{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	crdContent, err := os.ReadFile(crdLocation)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", crdLocation, err)
	}

	return matches.Validate(crdContent, payload, c.IgnoreErrors)
}

func init() {
	tests.Register(&Matcher{}, "matchString")
}
