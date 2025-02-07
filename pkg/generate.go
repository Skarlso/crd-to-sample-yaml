package pkg

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

const array = "array"

var RootRequiredFields = []string{"apiVersion", "kind", "spec", "metadata", "status"}

// Generate takes a CRD content and path, and outputs.
func Generate(crd *SchemaType, w io.WriteCloser, enableComments, minimal, skipRandom bool) (err error) {
	defer func() {
		if err := w.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close output file: %s", err.Error())
		}
	}()

	parser := NewParser(crd.Group, crd.Kind, enableComments, minimal, skipRandom)
	for i, version := range crd.Versions {
		if err := parser.ParseProperties(version.Name, w, version.Schema.Properties, RootRequiredFields); err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
		}

		if i < len(crd.Versions)-1 {
			if _, err := w.Write([]byte("\n---\n")); err != nil {
				return fmt.Errorf("failed to write yaml delimiter to writer: %w", err)
			}
		}
	}

	// Parse validation instead
	if len(crd.Versions) == 0 && crd.Validation != nil {
		if err := parser.ParseProperties(crd.Validation.Name, w, crd.Validation.Schema.Properties, RootRequiredFields); err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
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
	skipRandom   bool
}

// NewParser creates a new parser contains most of the things that do not change over each call.
func NewParser(group, kind string, comments, requiredOnly, skipRandom bool) *Parser {
	return &Parser{
		group:        group,
		kind:         kind,
		comments:     comments,
		onlyRequired: requiredOnly,
		skipRandom:   skipRandom,
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
		// skip the entire key if it's not required
		if p.onlyRequired && !slices.Contains(requiredFields, k) {
			// we can't put a closing `{}` in case there are no more
			// properties because the last entry is a new line in the writer.
			// This means that putting a `{}` here would put it after the new line.
			// Whereas we want it before the new line.
			continue
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
			// only set kind at the first level, after that it mist be something else.
			if k == "kind" && p.indent == 0 {
				w.write(file, fmt.Sprintf(" %s\n", p.kind))

				continue
			}
			// If we are dealing with an array, and we have properties to parse
			// we need to reparse all of them again.
			if properties[k].Type == array && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0 {
				w.write(file, fmt.Sprintf("\n%s- ", strings.Repeat(" ", p.indent)))
				p.indent += 2
				p.inArray = true

				if err := p.ParseProperties(version, file, properties[k].Items.Schema.Properties, properties[k].Items.Schema.Required); err != nil {
					return err
				}
				p.indent -= 2
			} else {
				w.write(file, fmt.Sprintf(" %s\n", outputValueType(properties[k], p.skipRandom)))
			}
		case len(properties[k].Properties) > 0:
			// recursively parse all sub-properties
			p.indent += 2
			if p.onlyRequired && p.emptyAfterTrimRequired(properties[k].Properties, properties[k].Required) {
				p.indent -= 2
				w.write(file, " {}\n")

				continue
			}

			w.write(file, "\n")
			if err := p.ParseProperties(version, file, properties[k].Properties, properties[k].Required); err != nil {
				return err
			}
			p.indent -= 2
		case properties[k].AdditionalProperties != nil:
			// if there are no properties defined but only additional properties, we will not generate the
			// additional properties because they are forbidden fields by the Schema Validation.
			if len(properties[k].Properties) == 0 ||
				(properties[k].AdditionalProperties.Schema == nil || len(properties[k].AdditionalProperties.Schema.Properties) == 0) {
				w.write(file, " {}\n")
			} else {
				p.indent += 2

				// This will eventually end up in case len(properties[k].Properties) > 0: hence, we don't need
				// an empty check in here. That would be redundant.
				if err := p.ParseProperties(
					version,
					file,
					properties[k].AdditionalProperties.Schema.Properties,
					properties[k].AdditionalProperties.Schema.Required,
				); err != nil {
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

// deletes properties from the properties that aren't required.
func (p *Parser) emptyAfterTrimRequired(properties map[string]v1beta1.JSONSchemaProps, required []string) bool {
	// we don't want to modify the original properties because that causes
	// various problems later down the line when we are trying to run
	// the same generate on the same values twice.
	for k := range properties {
		if slices.Contains(required, k) {
			return false
		}
	}

	return true
}

// outputValueType generate an output value based on the given type.
func outputValueType(v v1beta1.JSONSchemaProps, skipRandom bool) string {
	if v.Default != nil {
		return string(v.Default.Raw)
	}

	if v.Example != nil {
		return string(v.Example.Raw)
	}

	if v.Pattern != "" && !skipRandom {
		// if it's a valid regex, let's return a value that matches the regex
		// if not, we don't care
		if _, err := regexp.Compile(v.Pattern); err == nil {
			return gofakeit.Regex(v.Pattern) + " # " + v.Pattern
		}
	}

	if v.Enum != nil {
		var value []string
		for _, ev := range v.Enum {
			value = append(value, string(ev.Raw))
		}

		return string(v.Enum[0].Raw) + " # " + strings.Join(value, ", ")
	}

	st := "string"
	switch v.Type {
	case st:
		if v.Format == "date-time" {
			return "2024-10-11T12:48:44Z"
		}

		if v.Format == "byte" {
			return base64.StdEncoding.EncodeToString([]byte("string"))
		}

		return st
	case "integer":
		if v.Minimum != nil {
			return strconv.Itoa(int(*v.Minimum))
		}

		return "1"
	case "boolean":
		return "true"
	case "object":
		return "{}"
	case array: // deal with arrays of other types that weren't objects
		if v.Items.Schema != nil {
			t := v.Items.Schema.Type
			var s string
			var items []string
			if v.MinItems != nil {
				for range int(*v.MinItems) {
					items = append(items, t)
				}
			}
			s = fmt.Sprintf("[%s] # minItems %d of type %s", strings.Join(items, ","), len(items), t)

			return s
		}

		return "[]"
	}

	return v.Type
}
