package pkg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"slices"
	"sort"
)

type Index struct {
	Page []ViewPage
}

// Version wraps a top level version resource which contains the underlying openAPIV3Schema.
type Version struct {
	Version     string
	Kind        string
	Group       string
	Properties  []*Property
	Description string
	YAML        string
}

// ViewPage is the template for view.html.
type ViewPage struct {
	Title    string
	Versions []Version
}

var (
	//go:embed templates
	files     embed.FS
	templates map[string]*template.Template
)

// LoadTemplates creates a map of loaded templates that are primed and ready to be rendered.
func LoadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, "templates")
	if err != nil {
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}
		pt, err := template.ParseFS(files, "templates/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}

	return nil
}

// RenderContent creates an HTML website from the CRD content.
func RenderContent(w io.WriteCloser, crds []*SchemaType, comments, minimal, random bool) (err error) {
	allViews := make([]ViewPage, 0, len(crds))

	for _, crd := range crds {
		versions := make([]Version, 0)
		parser := NewParser(crd.Group, crd.Kind, comments, minimal, random)

		for _, version := range crd.Versions {
			v, err := generate(version.Name, crd.Group, crd.Kind, version.Schema, minimal, parser)
			if err != nil {
				return fmt.Errorf("failed to generate yaml sample: %w", err)
			}

			versions = append(versions, v)
		}

		// parse validation instead
		if len(versions) == 0 && crd.Validation != nil {
			version, err := generate(crd.Validation.Name, crd.Group, crd.Kind, crd.Validation.Schema, minimal, parser)
			if err != nil {
				return fmt.Errorf("failed to generate yaml sample: %w", err)
			}

			versions = append(versions, version)
		} else if len(versions) == 0 {
			continue
		}

		view := ViewPage{
			Title:    crd.Kind,
			Versions: versions,
		}

		allViews = append(allViews, view)
	}

	t := templates["view.html"]

	index := Index{
		Page: allViews,
	}

	if err := t.Execute(w, index); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func generate(name, group, kind string, properties *JSONSchemaProps, minimal bool, parser *Parser) (Version, error) {
	out, err := parseCRD(properties.Properties, name, minimal, RootRequiredFields)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse properties: %w", err)
	}
	var buffer []byte
	buf := bytes.NewBuffer(buffer)
	if err := parser.ParseProperties(name, buf, properties.Properties); err != nil {
		return Version{}, fmt.Errorf("failed to generate yaml sample: %w", err)
	}

	return Version{
		Version:     name,
		Properties:  out,
		Kind:        kind,
		Group:       group,
		Description: properties.Description,
		YAML:        buf.String(),
	}, nil
}

// Property builds up a Tree structure of embedded things.
type Property struct {
	Name        string
	Description string
	Type        string
	Nullable    bool
	Patterns    string
	Format      string
	Indent      int
	Version     string
	Default     string
	Required    bool
	Properties  []*Property
}

// parseCRD takes the properties and constructs a linked list out of the embedded properties that the recursive
// template can call and construct linked divs.
func parseCRD(properties map[string]JSONSchemaProps, version string, minimal bool, requiredList []string) ([]*Property, error) {
	output := make([]*Property, 0, len(properties))
	sortedKeys := make([]string, 0, len(properties))

	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		if minimal {
			if !slices.Contains(requiredList, k) {
				continue
			}
		}
		// Create the Property with the values necessary.
		// Check if there are properties for it in Properties or in Array -> Properties.
		// If yes, call parseCRD and add the result to the created properties Properties list.
		// If not, or if we are done, add this new property to the list of properties and return it.
		v := properties[k]
		required := false
		for _, item := range requiredList {
			if item == k {
				required = true

				break
			}
		}
		p := &Property{
			Name:        k,
			Type:        v.Type,
			Description: v.Description,
			Patterns:    v.Pattern,
			Format:      v.Format,
			Nullable:    v.Nullable,
			Version:     version,
			Required:    required,
		}
		if v.Default != nil {
			p.Default = string(v.Default.Raw)
		}

		switch {
		case len(properties[k].Properties) > 0 && properties[k].AdditionalProperties == nil:
			requiredList = v.Required
			out, err := parseCRD(properties[k].Properties, version, minimal, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		case properties[k].Type == array && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0:
			requiredList = v.Required
			out, err := parseCRD(properties[k].Items.Schema.Properties, version, minimal, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		case properties[k].AdditionalProperties != nil:
			requiredList = v.Required
			out, err := parseCRD(properties[k].AdditionalProperties.Schema.Properties, version, minimal, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		}

		output = append(output, p)
	}

	return output, nil
}
