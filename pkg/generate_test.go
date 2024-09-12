package pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestGenerate(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	version := crd.Spec.Versions[0]
	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithTemplateDelimiter(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	version := crd.Spec.Versions[0]
	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithComments(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, true, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_comments_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateMinimal(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, true, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateMinimalWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, true, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_with_example_for_field_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithAdditionalProperties(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties.yaml"))
	require.NoError(t, err)

	crd := &v1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties, RootRequiredFields))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}
