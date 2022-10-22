package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func Generate(content []byte, path string) error {
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition")
	}

	for _, version := range crd.Spec.Versions {
		outputLocation := filepath.Join(path, fmt.Sprintf("%s_%s.yaml", crd.Spec.Names.Kind, version.Name))
		outputFile, err := os.Create(outputLocation)
		if err != nil {
			return fmt.Errorf("failed to create file at: '%s': %w", outputLocation, err)
		}
		if err := parseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, outputFile, 0, false); err != nil {
			outputFile.Close()
			return fmt.Errorf("failed to parse properties: %w", err)
		}
		outputFile.Close()
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
		if len(properties[k].Properties) == 0 {
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
		}
	}
	if w.err != nil {
		return fmt.Errorf("failed to write to file: %w", w.err)
	}
	return nil
}

func outputValueType(v v1beta1.JSONSchemaProps) string {
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
