package matchstring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

// Matcher defines a match for String based matching.
type Matcher struct{}

// Config contains errors that could be ignored by this matcher.
type Config struct {
	IgnoreErrors []string `yaml:"ignoreErrors,omitempty"`
}

// Match does the actual Match job.
func (m *Matcher) Match(_ context.Context, crdLocation string, payload []byte) error {
	c := &Config{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	crdContent, err := os.ReadFile(filepath.Clean(crdLocation))
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", crdLocation, err)
	}

	return matches.Validate(crdContent, payload, c.IgnoreErrors)
}

func init() {
	tests.Register(&Matcher{}, "matchString")
}
