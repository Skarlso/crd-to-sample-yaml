package matchsnapshot

import (
	"fmt"
	"os"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

const MatcherName = "matchSnapshot"

type Config struct {
	Name string `yaml:"name"`
}
type Matcher struct{}

func (m *Matcher) Match(sourceTemplateLocation string, payload []byte) error {
	content, err := os.ReadFile(sourceTemplateLocation)
	if err != nil {
		return fmt.Errorf("failed to read source template: %w", err)
	}

	c := Config{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	snapshot, err := os.ReadFile(c.Name)
	if err != nil {
		return fmt.Errorf("failed to read snapshot template: %w", err)
	}

	return matches.Validate(content, snapshot)
}

func init() {
	tests.Register(&Matcher{}, MatcherName)
}
