package pkg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

var bulletRegex = regexp.MustCompile(`^\s*[-*+•]\s+(.+)$`)

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
	Conditions  []ConditionInfo
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

// parseDescription converts plain text descriptions into HTML with proper paragraph and list formatting.
func parseDescription(desc string) template.HTML {
	if desc == "" {
		return ""
	}

	// Split by double newlines to identify paragraphs
	paragraphs := strings.Split(strings.TrimSpace(desc), "\n\n")

	var result strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		lines := strings.Split(para, "\n")

		var (
			listItems    []string
			nonListLines []string
		)

		inList := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if bulletRegex.MatchString(line) {
				if !inList && len(nonListLines) > 0 {
					// Process accumulated non-list lines as paragraph first
					result.WriteString("<p>")
					result.WriteString(template.HTMLEscapeString(strings.Join(nonListLines, " ")))
					result.WriteString("</p>\n")

					nonListLines = nil
				}

				inList = true
				matches := bulletRegex.FindStringSubmatch(line)
				listItems = append(listItems, matches[1])
			} else {
				if inList {
					// End the list and start new paragraph
					if len(listItems) > 0 {
						result.WriteString("<ul>\n")

						for _, item := range listItems {
							result.WriteString("<li>")
							result.WriteString(template.HTMLEscapeString(item))
							result.WriteString("</li>\n")
						}

						result.WriteString("</ul>\n")

						listItems = nil
					}

					inList = false
				}

				nonListLines = append(nonListLines, line)
			}
		}

		// Handle remaining content
		if inList && len(listItems) > 0 {
			result.WriteString("<ul>\n")

			for _, item := range listItems {
				result.WriteString("<li>")
				result.WriteString(template.HTMLEscapeString(item))
				result.WriteString("</li>\n")
			}

			result.WriteString("</ul>\n")
		} else if len(nonListLines) > 0 {
			result.WriteString("<p>")
			result.WriteString(template.HTMLEscapeString(strings.Join(nonListLines, " ")))
			result.WriteString("</p>\n")
		}
	}

	return template.HTML(result.String()) //nolint:gosec // desc is escaped elsewhere during formatting.
}

// LoadTemplates creates a map of loaded templates that are primed and ready to be rendered.
func LoadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	tmplFiles, err := fs.ReadDir(files, "templates")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"parseDescription": parseDescription,
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.New(tmpl.Name()).Funcs(funcMap).ParseFS(files, "templates/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}

	return nil
}

// Group defines a single group with a list of rendered versions.
type Group struct {
	Name string
	Page []ViewPage
}

// GroupPage will have a list of groups and inside these groups
// will be a list of page views.
type GroupPage struct {
	Groups    []Group
	CustomCSS template.CSS
}

type RenderOpts struct {
	Comments  bool
	Minimal   bool
	Random    bool
	CustomCSS string
}

// RenderContent creates an HTML website from the CRD content.
func RenderContent(w io.WriteCloser, crds []*SchemaType, opts RenderOpts) (err error) {
	defer func() {
		err := w.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close output file: %s", err.Error())
		}
	}()

	groups := buildUpGroup(crds)

	allGroups := make([]Group, 0)
	for name, group := range groups {
		allViews := make([]ViewPage, 0, len(group))

		for _, crd := range group {
			versions := make([]Version, 0)
			parser := NewParser(crd.Group, crd.Kind, opts.Comments, opts.Minimal, opts.Random)

			for _, version := range crd.Versions {
				v, err := generate(version.Name, crd.Group, crd.Kind, version.Schema, opts.Minimal, parser)
				if err != nil {
					return fmt.Errorf("failed to generate yaml sample: %w", err)
				}

				v.Conditions = crd.Conditions
				versions = append(versions, v)
			}

			// parse validation instead
			if len(versions) == 0 && crd.Validation != nil {
				version, err := generate(crd.Validation.Name, crd.Group, crd.Kind, crd.Validation.Schema, opts.Minimal, parser)
				if err != nil {
					return fmt.Errorf("failed to generate yaml sample: %w", err)
				}

				version.Conditions = crd.Conditions
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

		allGroups = append(allGroups, Group{
			Name: name,
			Page: allViews,
		})
	}

	t := templates["view_with_groups.html"]

	index := GroupPage{
		Groups:    allGroups,
		CustomCSS: template.CSS(opts.CustomCSS), //nolint:gosec // opts.CustomCSS is escaped and sanitized input
	}

	if err := t.Execute(w, index); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func buildUpGroup(crds []*SchemaType) map[string][]*SchemaType {
	result := map[string][]*SchemaType{}

	for _, crd := range crds {
		if crd.Rendering.Group == "" {
			crd.Rendering.Group = crd.Group
		}

		result[crd.Rendering.Group] = append(result[crd.Rendering.Group], crd)
	}

	return result
}

func generate(name, group, kind string, properties *v1beta1.JSONSchemaProps, minimal bool, parser *Parser) (Version, error) {
	out, err := parseCRD(properties.Properties, name, minimal, group, kind, RootRequiredFields, 0)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse properties: %w", err)
	}

	var buffer []byte

	buf := bytes.NewBuffer(buffer)
	if err := parser.ParseProperties(name, buf, properties.Properties, RootRequiredFields); err != nil {
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
	Examples    string
	Type        string
	Nullable    bool
	Patterns    string
	Format      string
	Indent      int
	Version     string
	Default     string
	Required    bool
	Properties  []*Property
	Enums       string
}

// parseCRD takes the properties and constructs a linked list out of the embedded properties that the recursive
// template can call and construct linked divs.
func parseCRD(properties map[string]v1beta1.JSONSchemaProps, version string, minimal bool, group string, kind string, requiredList []string, depth int) ([]*Property, error) {
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

		if slices.Contains(requiredList, k) {
			required = true
		}

		description := v.Description
		if description == "" && depth == 0 {
			if k == "apiVersion" {
				description = fmt.Sprintf("%s/%s", group, version)
			}

			if k == "kind" {
				description = kind
			}
		}

		var enums []string
		for _, e := range v.Enum {
			enums = append(enums, string(e.Raw))
		}

		p := &Property{
			Name:        k,
			Type:        v.Type,
			Description: description,
			Patterns:    v.Pattern,
			Format:      v.Format,
			Nullable:    v.Nullable,
			Version:     version,
			Required:    required,
			Enums:       strings.Join(enums, ", "),
		}
		if v.Default != nil {
			p.Default = string(v.Default.Raw)
		}

		if v.Example != nil {
			p.Examples = string(v.Example.Raw)
		}

		switch {
		case len(properties[k].Properties) > 0:
			requiredList = v.Required
			depth++

			out, err := parseCRD(properties[k].Properties, version, minimal, group, kind, requiredList, depth)
			if err != nil {
				return nil, err
			}

			depth--
			p.Properties = out
		case properties[k].Type == array && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0:
			depth++
			requiredList = v.Required

			out, err := parseCRD(properties[k].Items.Schema.Properties, version, minimal, group, kind, requiredList, depth)
			if err != nil {
				return nil, err
			}

			depth--
			p.Properties = out
		case properties[k].AdditionalProperties != nil && properties[k].AdditionalProperties.Schema != nil:
			depth++
			requiredList = v.Required

			out, err := parseCRD(properties[k].AdditionalProperties.Schema.Properties, version, minimal, group, kind, requiredList, depth)
			if err != nil {
				return nil, err
			}

			depth--
			p.Properties = out
		}

		output = append(output, p)
	}

	return output, nil
}
