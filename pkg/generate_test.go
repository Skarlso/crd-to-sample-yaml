package pkg

import (
	"bytes"
	"io"
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
	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, golden, buffer.Bytes())
}

func TestGenerateWithTemplateDelimiter(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	version := crd.Spec.Versions[0]
	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithComments(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, true, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_comments_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimal(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, true, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimalWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, true, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_with_example_for_field_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimalWithNoRequiredFields(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_minimal_no_required_fields.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, true, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_minimal_no_required_fields_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithAdditionalProperties(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)
	version := crd.Spec.Versions[0]
	require.NoError(t, parser.ParseProperties(version.Name, buffer, version.Schema.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithValidation(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_validation.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)

	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, false, false, true)

	crd.Spec.Validation.OpenAPIV3Schema.Properties["kind"] = v1beta1.JSONSchemaProps{}
	crd.Spec.Validation.OpenAPIV3Schema.Properties["apiVersion"] = v1beta1.JSONSchemaProps{}
	require.NoError(t, parser.ParseProperties(crd.Name, buffer, crd.Spec.Validation.OpenAPIV3Schema.Properties))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_validation_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

type WriteNoOpCloser struct {
	w io.Writer
}

func (w WriteNoOpCloser) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *WriteNoOpCloser) Close() error { return nil }

func TestGenerateWithMultipleVersionsAndList(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_list_and_multiple_versions.yaml"))
	require.NoError(t, err)

	crd := &v1beta1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(content, crd))

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(crd, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_list_and_multiple_versions_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}
