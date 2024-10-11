package matches

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

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

	if crd.Spec.Validation != nil && len(crd.Spec.Versions) == 0 {
		return ValidateCRDValidation(crd, sampleFile)
	}

	availableVersions := make([]string, 0, len(crd.Spec.Versions))

	// Add checking out the api version from the provided template and only eval against that.
	for _, v := range crd.Spec.Versions {
		availableVersions = append(availableVersions, v.Name)

		// Make sure we are only testing versions that equal to the CRD's version.
		// This is important in case there are multiple versions in the CRD.
		if obj.GroupVersionKind().Version == v.Name {
			if err := validate(v.Schema.OpenAPIV3Schema, obj, crd.Spec.Names.Kind, v.Name); err != nil {
				return fmt.Errorf("failed to validate kind %s and version %s: %w", crd.Spec.Names.Kind, v.Name, err)
			}

			return nil
		}
	}

	return fmt.Errorf(
		"version of the snapshot %s not found amongst the available testing versions of the CRD %s",
		obj.GroupVersionKind().Version,
		strings.Join(availableVersions, ","),
	)
}

func ValidateCRDValidation(crd *apiextensions.CustomResourceDefinition, sampleFile []byte) error {
	reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(sampleFile), maxBufferSize)
	obj := &unstructured.Unstructured{}
	err := reader.Decode(obj)
	if err != nil {
		return fmt.Errorf("failed to decode sample file: %w", err)
	}

	return validate(crd.Spec.Validation.OpenAPIV3Schema, obj, crd.Spec.Names.Kind, crd.Name)
}

func validate(props *apiextensions.JSONSchemaProps, obj *unstructured.Unstructured, kind, name string) error {
	eval, _, err := validation.NewSchemaValidator(props)
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
		return fmt.Errorf("failed to validate kind %s and version %s: %w", kind, name, resultErrors)
	}

	return nil
}
