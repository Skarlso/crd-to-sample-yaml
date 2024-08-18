package matches

import (
	"bytes"
	"errors"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const maxBufferSize = 2048

// Validate takes a source CRD and a sample file and validates its contents against the CRD definition.
func Validate(sourceCRD []byte, sampleFile []byte) error {
	reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(sampleFile), maxBufferSize)
	obj := &unstructured.Unstructured{}
	err := reader.Decode(obj)
	if err != nil {
		return fmt.Errorf("failed to decode sample file: %w", err)
	}

	crd := &apiextensions.CustomResourceDefinition{}
	if err := yaml.Unmarshal(sourceCRD, crd); err != nil {
		return errors.New("failed to unmarshal into custom resource definition")
	}

	// Add checking out the api version from the provided template and only eval against that.
	// TODO: this should be a specific version instead.
	for _, v := range crd.Spec.Versions {
		eval, _, err := validation.NewSchemaValidator(v.Schema.OpenAPIV3Schema)
		if err != nil {
			return fmt.Errorf("invalid schema: %w", err)
		}

		var resultErrors error
		result := eval.Validate(obj)
		for _, e := range result.Errors {
			resultErrors = errors.Join(resultErrors, e)
		}

		for _, e := range result.Warnings {
			resultErrors = errors.Join(resultErrors, e)
		}

		if resultErrors != nil {
			return fmt.Errorf("failed to validate kind %s: %w", crd.Spec.Names.Kind, resultErrors)
		}
	}

	return nil
}
