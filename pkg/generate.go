package pkg

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Generate takes a CRD content and path, and outputs
func Generate(crd *v1beta1.CustomResourceDefinition, w io.WriteCloser) (err error) {
	defer func() {
		if cerr := w.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	for i, version := range crd.Spec.Versions {
		if err := parseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, w, 0, false); err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
		}

		if i < len(crd.Spec.Versions)-1 {
			if _, err := w.Write([]byte("\n---\n")); err != nil {
				return fmt.Errorf("failed to write yaml delimiter to writer: %w", err)
			}
		}
	}

	return nil
}

type writer struct {
	err error
}

func (w *writer) write(wc io.Writer, msg string) {
	if w.err != nil {
		return
	}
	_, w.err = wc.Write([]byte(msg))
}

// parseProperties takes a writer and puts out any information / properties it encounters during the runs.
// It will recursively parse every "properties:" and "additionalProperties:". Using the types, it will also output
// some sample data based on those types.
func parseProperties(group, version, kind string, properties map[string]v1beta1.JSONSchemaProps, file io.Writer, indent int, inArray bool) error {
	var sortedKeys []string
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	w := &writer{}
	for _, k := range sortedKeys {
		if inArray {
			w.write(file, fmt.Sprintf("%s:", k))
			inArray = false
		} else {
			w.write(file, fmt.Sprintf("%s%s:", strings.Repeat(" ", indent), k))
		}
		if len(properties[k].Properties) == 0 && properties[k].AdditionalProperties == nil {
			if k == "apiVersion" {
				w.write(file, fmt.Sprintf(" %s/%s\n", group, version))
				continue
			}
			if k == "kind" {
				w.write(file, fmt.Sprintf(" %s\n", kind))
				continue
			}
			// If we are dealing with an array, and we have properties to parse
			// we need to reparse all of them again.
			var result string
			if properties[k].Type == "array" && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0 {
				w.write(file, fmt.Sprintf("\n%s- ", strings.Repeat(" ", indent)))
				if err := parseProperties(group, version, kind, properties[k].Items.Schema.Properties, file, indent+2, true); err != nil {
					return err
				}
			} else {
				result = outputValueType(properties[k])
				w.write(file, fmt.Sprintf(" %s\n", result))
			}
		} else if len(properties[k].Properties) > 0 {
			w.write(file, "\n")
			// recursively parse all sub-properties
			if err := parseProperties(group, version, kind, properties[k].Properties, file, indent+2, false); err != nil {
				return err
			}
		} else if properties[k].AdditionalProperties != nil {
			w.write(file, "\n")

			if err := parseProperties(group, version, kind, properties[k].AdditionalProperties.Schema.Properties, file, indent+2, false); err != nil {
				return err
			}
		}
	}
	if w.err != nil {
		return fmt.Errorf("failed to write to file: %w", w.err)
	}
	return nil
}

// outputValueType generate an output value based on the given type.
func outputValueType(v v1beta1.JSONSchemaProps) string {
	if v.Default != nil {
		return string(v.Default.Raw)
	}

	switch v.Type {
	case "string":
		return "string"
	case "integer":
		return "1"
	case "boolean":
		return "true"
	case "object":
		return "{}"
	case "array": // deal with arrays of other types that weren't objects
		t := v.Items.Schema.Type
		var s string
		if t == "string" {
			s = fmt.Sprintf("[\"%s\"]", t)
		} else {
			s = fmt.Sprintf("[%s]", t)
		}
		return s
	}
	return v.Type
}
