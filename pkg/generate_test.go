package pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestGenerate(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	version := crd.Spec.Versions[0]
	require.NoError(t, ParseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, buffer, 0, false, false))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	version := crd.Spec.Versions[0]
	require.NoError(t, ParseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, buffer, 0, false, false))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}
