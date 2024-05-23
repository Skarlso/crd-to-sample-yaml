package pkg

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const array = "array"

var rootRequiredFields = []string{"apiVersion", "kind", "spec"}

// Generate takes a CRD content and path, and outputs.
func Generate(crd *v1beta1.CustomResourceDefinition, w io.WriteCloser, enableComments, minimal bool) (err error) {
	defer func() {
		if cerr := w.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()
	parser := NewParser(crd.Spec.Group, crd.Spec.Names.Kind, enableComments, minimal)
	for i, version := range crd.Spec.Versions {
		if err := parser.ParseProperties(version.Name, w, version.Schema.OpenAPIV3Schema.Properties, rootRequiredFields); err != nil {
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

type Parser struct {
	comments     bool
	inArray      bool
	indent       int
	group        string
	kind         string
	onlyRequired bool
}

// NewParser creates a new parser contains most of the things that do not change over each call.
func NewParser(group, kind string, comments, requiredOnly bool) *Parser {
	return &Parser{
		group:        group,
		kind:         kind,
		comments:     comments,
		onlyRequired: requiredOnly,
	}
}

// ParseProperties takes a writer and puts out any information / properties it encounters during the runs.
// It will recursively parse every "properties:" and "additionalProperties:". Using the types, it will also output
// some sample data based on those types.
func (p *Parser) ParseProperties(version string, file io.Writer, properties map[string]v1beta1.JSONSchemaProps, requiredFields []string) error {
	sortedKeys := make([]string, 0, len(properties))
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	w := &writer{}
	for _, k := range sortedKeys {
		// if field is not required, skip the entire flow.
		if p.onlyRequired {
			if !slices.Contains(requiredFields, k) {
				continue
			}
		}

		if p.inArray {
			w.write(file, k+":")
			p.inArray = false
		} else {
			if p.comments && properties[k].Description != "" {
				comment := strings.Builder{}
				multiLine := strings.Split(properties[k].Description, "\n")
				for _, line := range multiLine {
					comment.WriteString(fmt.Sprintf("%s# %s\n", strings.Repeat(" ", p.indent), line))
				}

				w.write(file, comment.String())
			}

			w.write(file, fmt.Sprintf("%s%s:", strings.Repeat(" ", p.indent), k))
		}
		switch {
		case len(properties[k].Properties) == 0 && properties[k].AdditionalProperties == nil:
			if k == "apiVersion" {
				w.write(file, fmt.Sprintf(" %s/%s\n", p.group, version))

				continue
			}
			if k == "kind" {
				w.write(file, fmt.Sprintf(" %s\n", p.kind))

				continue
			}
			// If we are dealing with an array, and we have properties to parse
			// we need to reparse all of them again.
			var result string
			if properties[k].Type == array && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0 {
				w.write(file, fmt.Sprintf("\n%s- ", strings.Repeat(" ", p.indent)))
				p.indent += 2
				p.inArray = true
				if err := p.ParseProperties(version, file, properties[k].Items.Schema.Properties, properties[k].Items.Schema.Required); err != nil {
					return err
				}
				p.indent -= 2
			} else {
				result = outputValueType(properties[k])
				w.write(file, fmt.Sprintf(" %s\n", result))
			}
		case len(properties[k].Properties) > 0:
			w.write(file, "\n")
			// recursively parse all sub-properties
			p.indent += 2
			if err := p.ParseProperties(version, file, properties[k].Properties, properties[k].Required); err != nil {
				return err
			}
			p.indent -= 2
		case properties[k].AdditionalProperties != nil:
			if len(properties[k].AdditionalProperties.Schema.Properties) == 0 {
				w.write(file, " {}\n")
			} else {
				w.write(file, "\n")

				p.indent += 2
				if err := p.ParseProperties(version, file, properties[k].AdditionalProperties.Schema.Properties, properties[k].AdditionalProperties.Schema.Required); err != nil {
					return err
				}
				p.indent -= 2
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

	if v.Example != nil {
		return string(v.Example.Raw)
	}

	st := "string"
	switch v.Type {
	case st:
		return st
	case "integer":
		return "1"
	case "boolean":
		return "true"
	case "object":
		return "{}"
	case array: // deal with arrays of other types that weren't objects
		t := v.Items.Schema.Type
		var s string
		if t == st {
			s = fmt.Sprintf("[\"%s\"]", t)
		} else {
			s = fmt.Sprintf("[%s]", t)
		}

		return s
	}

	return v.Type
}
