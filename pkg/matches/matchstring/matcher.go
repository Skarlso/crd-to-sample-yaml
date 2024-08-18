package matchstring

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

type Matcher struct{}

func (m *Matcher) Match(sourceTemplateLocation string, payload []byte) error {
	c := &apiextensionsv1.JSON{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	return nil
}

func init() {
	tests.Register(&Matcher{}, "matchString")
}
