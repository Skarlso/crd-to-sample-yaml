package pkg

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

func TestGenerate(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithTemplateDelimiter(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_template_start_character_default_value_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithComments(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, true, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_comments_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimal(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, true, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimalWithExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_example.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, true, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_minimal_example_with_example_for_field_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateMinimalWithNoRequiredFields(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_minimal_no_required_fields.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, true, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_minimal_no_required_fields_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithAdditionalProperties(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_additional_properties_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithValidation(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_validation.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	schemaType.Validation.Schema.Properties["kind"] = v1beta1.JSONSchemaProps{}
	schemaType.Validation.Schema.Properties["apiVersion"] = v1beta1.JSONSchemaProps{}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

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

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_with_list_and_multiple_versions_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}

func TestGenerateWithDifferentCRDType(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "sample_crd_different_crd_type.yaml"))
	require.NoError(t, err)

	crd := &unstructured.Unstructured{}
	require.NoError(t, yaml.Unmarshal(content, crd))
	schemaType, err := ExtractSchemaType(crd)
	require.NoError(t, err)

	var output []byte
	buffer := bytes.NewBuffer(output)
	nopCloser := &WriteNoOpCloser{w: buffer}
	require.NoError(t, Generate(schemaType, nopCloser, false, false, true))

	golden, err := os.ReadFile(filepath.Join("testdata", "sample_crd_different_crd_type_golden.yaml"))
	require.NoError(t, err)

	assert.Equal(t, string(golden), buffer.String())
}
