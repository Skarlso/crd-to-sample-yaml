package pkg

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

// ExtractSchemaType makes sure the following required fields are
// present in the unstructured data and creates are own internal representation:
// - spec
// - spec.names.kind
// - spec.group
// either:
// - versions
//   - version.Schema.OpenAPIV3Schema
//
// - validation // if versions is missing
//   - validation.OpenAPIV3Schema
func ExtractSchemaType(obj *unstructured.Unstructured) (*SchemaType, error) {
	spec, err := extractValue[map[string]any](obj.Object, "spec")
	if err != nil {
		return nil, err
	}

	versions, ok := spec["versions"]
	if !ok {
		if _, ok := spec["validation"]; !ok {
			// we aren't dealing with a valid resource
			// we might want to skip it if we are going through a
			// list of YAML files in a folder, and we want to skip
			// invalid ones.
			return nil, nil
		}

		return extractValidation(obj, spec)
	}

	kind, group, err := extractGroupKind(spec)
	if err != nil {
		return nil, err
	}

	versionsList, ok := versions.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid version list type not a list: %T", versionsList)
	}

	schemaTypes := &SchemaType{
		Group: group,
		Kind:  kind,
	}

	for _, v := range versionsList {
		vMap, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid version type not a map: %T", v)
		}

		name, err := extractValue[string](vMap, "name")
		if err != nil {
			return nil, fmt.Errorf("no name found for version: %v", v)
		}

		schema, ok := vMap["schema"]
		if !ok {
			return nil, fmt.Errorf("no schema found for version: %v", v)
		}

		openAPIV3schema, err := extractValue[map[string]any](schema, "openAPIV3Schema")
		if err != nil {
			return nil, err
		}

		content, err := json.Marshal(openAPIV3schema)
		if err != nil {
			return nil, err
		}

		schemaValue := &v1beta1.JSONSchemaProps{}
		if err := json.Unmarshal(content, schemaValue); err != nil {
			return nil, err
		}

		ensureKindAndAPIVersionIsSet(schemaValue.Properties)

		version := &CRDVersion{
			Name:   name,
			Schema: schemaValue,
		}

		schemaTypes.Versions = append(schemaTypes.Versions, version)
	}

	return schemaTypes, nil
}

func extractValidation(obj *unstructured.Unstructured, specMap map[string]any) (*SchemaType, error) {
	validation, err := extractValue[map[string]any](specMap, "validation")
	if err != nil {
		return nil, err
	}

	schema, ok := validation["openAPIV3Schema"]
	if !ok {
		return nil, fmt.Errorf("openAPIV3Schema not found in validation map: %v", validation)
	}

	props := &v1beta1.JSONSchemaProps{}

	content, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(content, props); err != nil {
		return nil, err
	}

	ensureKindAndAPIVersionIsSet(props.Properties)

	kindValue, groupValue, err := extractGroupKind(specMap)
	if err != nil {
		return nil, err
	}

	return &SchemaType{
		Schema: nil,
		Validation: &Validation{
			Schema: props,
			Name:   obj.GetName(),
		},
		Group: groupValue,
		Kind:  kindValue,
	}, nil
}

func ensureKindAndAPIVersionIsSet(properties map[string]v1beta1.JSONSchemaProps) {
	if _, ok := properties["kind"]; !ok {
		properties["kind"] = v1beta1.JSONSchemaProps{}
	}

	if _, ok := properties["apiVersion"]; !ok {
		properties["apiVersion"] = v1beta1.JSONSchemaProps{}
	}
}

func extractGroupKind(specMap map[string]any) (string, string, error) {
	names, ok := specMap["names"]
	if !ok {
		return "", "", errors.New("no names found in object")
	}

	kind, err := extractValue[string](names, "kind")
	if err != nil {
		return "", "", err
	}

	group, err := extractValue[string](specMap, "group")
	if err != nil {
		return "", "", err
	}

	return kind, group, nil
}

// extractValue fetches a specific key value that we are looking for in a map.
func extractValue[T any](m any, k string) (T, error) {
	var result T

	v, ok := m.(map[string]any)
	if !ok {
		return result, fmt.Errorf("value was not of type map[string]any but: %T", m)
	}

	vv, ok := v[k]
	if !ok {
		return result, fmt.Errorf("key %s was not found in map", k)
	}

	vvv, ok := vv.(T)
	if !ok {
		return result, fmt.Errorf("value was not of type T but: %T", vvv)
	}

	return vvv, nil
}
