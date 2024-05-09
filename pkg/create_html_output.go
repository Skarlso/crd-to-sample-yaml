package pkg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"sort"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

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
func RenderContent(w io.Writer, crdContent []byte, comments bool) error {
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(crdContent, crd); err != nil {
		return fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}
	versions := make([]Version, 0)
	for _, version := range crd.Spec.Versions {
		out, err := parseCRD(version.Schema.OpenAPIV3Schema.Properties, version.Name, version.Schema.OpenAPIV3Schema.Required)
		if err != nil {
			return fmt.Errorf("failed to parse properties: %w", err)
		}
		var buffer []byte
		buf := bytes.NewBuffer(buffer)
		if err := ParseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, buf, 0, false, comments); err != nil {
			return fmt.Errorf("failed to generate yaml sample: %w", err)
		}
		versions = append(versions, Version{
			Version:     version.Name,
			Properties:  out,
			Kind:        crd.Spec.Names.Kind,
			Group:       crd.Spec.Group,
			Description: version.Schema.OpenAPIV3Schema.Description,
			YAML:        buf.String(),
		})
	}
	view := ViewPage{
		Versions: versions,
	}
	t := templates["view.html"]
	if err := t.Execute(w, view); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
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
func parseCRD(properties map[string]v1beta1.JSONSchemaProps, version string, requiredList []string) ([]*Property, error) {
	output := make([]*Property, 0, len(properties))
	sortedKeys := make([]string, 0, len(properties))

	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
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
			out, err := parseCRD(properties[k].Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		case properties[k].Type == array && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0:
			requiredList = v.Required
			out, err := parseCRD(properties[k].Items.Schema.Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		case properties[k].AdditionalProperties != nil:
			requiredList = v.Required
			out, err := parseCRD(properties[k].AdditionalProperties.Schema.Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		}

		output = append(output, p)
	}

	return output, nil
}