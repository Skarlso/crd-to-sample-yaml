package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractSchemaTypeForVersion(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "this-is-my-name",
			},
			"spec": map[string]interface{}{
				"group": "group",
				"names": map[string]interface{}{
					"kind": "kind",
				},
				"versions": []any{
					map[string]interface{}{
						"name": "v1beta1",
						"schema": map[string]interface{}{
							"openAPIV3Schema": map[string]interface{}{
								"type":                 "object",
								"properties":           map[string]interface{}{},
								"additionalProperties": map[string]interface{}{},
								"additionalItems":      map[string]interface{}{},
								"id":                   "id",
								"title":                "title",
							},
						},
					},
				},
			},
		},
	}

	schemaType, err := ExtractSchemaType(obj)
	require.NoError(t, err)
	assert.Equal(t, "object", schemaType.Versions[0].Schema.Type)
	assert.Equal(t, "id", schemaType.Versions[0].Schema.ID)
	assert.Equal(t, "title", schemaType.Versions[0].Schema.Title)
}

func TestExtractSchemaTypeForValidation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "this-is-my-name",
			},
			"spec": map[string]interface{}{
				"group": "group",
				"names": map[string]interface{}{
					"kind": "kind",
				},
				"validation": map[string]interface{}{
					"openAPIV3Schema": map[string]interface{}{
						"type":                 "object",
						"properties":           map[string]interface{}{},
						"additionalProperties": map[string]interface{}{},
						"additionalItems":      map[string]interface{}{},
						"id":                   "id",
						"title":                "title",
					},
				},
			},
		},
	}

	schemaType, err := ExtractSchemaType(obj)
	require.NoError(t, err)
	assert.Equal(t, "this-is-my-name", schemaType.Validation.Name)
	assert.Equal(t, "object", schemaType.Validation.Schema.Type)
	assert.Equal(t, "id", schemaType.Validation.Schema.ID)
	assert.Equal(t, "title", schemaType.Validation.Schema.Title)
}
