// Package main contains the application main code for the WASM codebase.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

// timeout after 2 seconds.
const timeout = 2000

var bulletRegex = regexp.MustCompile(`^\s*[-*+•]\s+(.+)$`)

// parseDescriptionElements converts plain text descriptions into go-app UI elements with proper paragraph and list formatting.
func parseDescriptionElements(desc string) []app.UI {
	if desc == "" {
		return nil
	}

	var elements []app.UI

	// Split by double newlines to identify paragraphs
	paragraphs := strings.SplitSeq(strings.TrimSpace(desc), "\n\n")

	for para := range paragraphs {
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
					// Add accumulated non-list lines as paragraph first
					elements = append(elements, app.P().Text(strings.Join(nonListLines, " ")))
					nonListLines = nil
				}

				inList = true
				matches := bulletRegex.FindStringSubmatch(line)
				listItems = append(listItems, matches[1])
			} else {
				if inList {
					// End the list and start new paragraph
					if len(listItems) > 0 {
						var liElements []app.UI
						for _, item := range listItems {
							liElements = append(liElements, app.Li().Text(item))
						}

						elements = append(elements, app.Ul().Body(liElements...))
						listItems = nil
					}

					inList = false
				}

				nonListLines = append(nonListLines, line)
			}
		}

		// Handle remaining content
		if inList && len(listItems) > 0 {
			var liElements []app.UI
			for _, item := range listItems {
				liElements = append(liElements, app.Li().Text(item))
			}

			elements = append(elements, app.Ul().Body(liElements...))
		} else if len(nonListLines) > 0 {
			elements = append(elements, app.P().Text(strings.Join(nonListLines, " ")))
		}
	}

	return elements
}

// crdView is the main component to display a rendered CRD.
type crdView struct {
	app.Compo

	preRenderErr error

	crds        []*pkg.SchemaType
	comment     bool
	minimal     bool
	originalURL string

	navigateBackOnClick func(ctx app.Context, _ app.Event)
}

type detailsView struct {
	app.Compo

	content   string
	comment   bool
	minimal   bool
	renderErr error
	version   *Version
}

func (v *detailsView) OnMount(_ app.Context) {
	content, err := v.version.generateYAMLDetails(v.comment, v.minimal)
	if err != nil {
		v.renderErr = err
		v.content = ""

		return
	}

	v.content = content
}

func (v *detailsView) Render() app.UI {
	return app.Div().Class("stack gap-4").Body(
		app.Div().Class("cols").Body(
			app.Label().Class("switch").For("enable-comments-"+v.version.Version).Body(
				app.Input().Type("checkbox").ID("enable-comments-"+v.version.Version).OnClick(v.OnCheckComment),
				app.Span().Class("switch-track"),
				app.Span().Body(
					app.Span().Class("strong").Text("Include comments"),
					app.Br(),
					app.Span().Class("hint").Text("Add field descriptions"),
				),
			),
			app.Label().Class("switch").For("enable-minimal-"+v.version.Version).Body(
				app.Input().Type("checkbox").ID("enable-minimal-"+v.version.Version).OnClick(v.OnCheckMinimal),
				app.Span().Class("switch-track"),
				app.Span().Body(
					app.Span().Class("strong").Text("Minimal output"),
					app.Br(),
					app.Span().Class("hint").Text("Only required fields"),
				),
			),
		),

		app.Div().Class("code-wrap").Body(
			app.Button().Class("btn-icon").
				ID("copy-btn-"+v.version.Version).
				Title("Copy to clipboard").
				Aria("label", "Copy YAML to clipboard").
				OnClick(v.onCopyClick).Body(icon("copy")),

			app.Pre().Class("code").Body(
				app.Code().ID("yaml-sample-"+v.version.Version).Body(app.If(v.renderErr != nil, func() app.UI {
					return app.Span().Class("center").Style("color", "var(--danger)").Body(
						icon("alert-circle"),
						app.Text(v.renderErr.Error()),
					)
				}).Else(func() app.UI {
					return app.Text(v.content)
				})),
			),
		),
	)
}

func (v *detailsView) OnCheckComment(_ app.Context, _ app.Event) {
	v.comment = !v.comment

	content, err := v.version.generateYAMLDetails(v.comment, v.minimal)
	if err != nil {
		v.renderErr = err
		v.content = ""

		return
	}

	v.content = content
}

func (v *detailsView) OnCheckMinimal(_ app.Context, _ app.Event) {
	v.minimal = !v.minimal

	content, err := v.version.generateYAMLDetails(v.comment, v.minimal)
	if err != nil {
		v.renderErr = err
		v.content = ""

		return
	}

	v.content = content
}

func (v *detailsView) onCopyClick(ctx app.Context, _ app.Event) {
	// Use the Clipboard API to copy the text
	content := v.content
	if content == "" {
		return
	}

	// Use JavaScript's clipboard API
	app.Window().Get("navigator").Get("clipboard").Call("writeText", content).Call("then", app.FuncOf(func(this app.Value, args []app.Value) any {
		// Show success feedback
		btn := ctx.JSSrc()
		originalHTML := btn.Get("innerHTML").String()
		btn.Set("innerHTML", checkIconHTML)
		btn.Get("classList").Call("add", "is-copied")

		app.Window().Call("setTimeout", app.FuncOf(func(this app.Value, args []app.Value) any {
			btn.Set("innerHTML", originalHTML)
			btn.Get("classList").Call("remove", "is-copied")
			btn.Get("classList").Call("add", "copy-btn")

			return nil
		}), timeout)

		return nil
	}))
}

// Version wraps a top level version resource which contains the underlying openAPIV3Schema.
type Version struct {
	Version     string
	Kind        string
	Group       string
	Properties  []*Property
	Description string
	Schema      map[string]v1beta1.JSONSchemaProps
}

func (v *Version) generateYAMLDetails(comment bool, minimal bool) (string, error) {
	buf := bytes.NewBuffer(nil)

	parser := pkg.NewParser(v.Group, v.Kind, comment, minimal, true)
	err := parser.ParseProperties(v.Version, buf, v.Schema, pkg.RootRequiredFields)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// DiffLine represents a single line in a diff with its type.
type DiffLine struct {
	Type    string // "added", "removed", "unchanged"
	Content string
	LineNum int
}

// simpleDiff performs a basic line-by-line diff between two strings.
func simpleDiff(oldContent, newContent string) []DiffLine {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var result []DiffLine

	oldIdx, newIdx := 0, 0

	for oldIdx < len(oldLines) && newIdx < len(newLines) {
		if oldLines[oldIdx] == newLines[newIdx] {
			result = append(result, DiffLine{
				Type:    "unchanged",
				Content: oldLines[oldIdx],
				LineNum: newIdx + 1,
			})
			oldIdx++
			newIdx++
		} else {
			// Look ahead to find matching lines
			matchFound := false

			for lookAhead := 1; lookAhead < 3 && (newIdx+lookAhead) < len(newLines); lookAhead++ {
				if oldLines[oldIdx] == newLines[newIdx+lookAhead] {
					// Add the new lines before the match
					for i := range lookAhead {
						result = append(result, DiffLine{
							Type:    "added",
							Content: newLines[newIdx+i],
							LineNum: newIdx + i + 1,
						})
					}

					newIdx += lookAhead
					matchFound = true

					break
				}
			}

			if !matchFound {
				// Look ahead in old lines
				for lookAhead := 1; lookAhead < 3 && (oldIdx+lookAhead) < len(oldLines); lookAhead++ {
					if oldLines[oldIdx+lookAhead] == newLines[newIdx] {
						// Add the removed lines
						for i := range lookAhead {
							result = append(result, DiffLine{
								Type:    "removed",
								Content: oldLines[oldIdx+i],
								LineNum: oldIdx + i + 1,
							})
						}

						oldIdx += lookAhead
						matchFound = true

						break
					}
				}
			}

			if !matchFound {
				result = append(result, DiffLine{
					Type:    "removed",
					Content: oldLines[oldIdx],
					LineNum: oldIdx + 1,
				})
				result = append(result, DiffLine{
					Type:    "added",
					Content: newLines[newIdx],
					LineNum: newIdx + 1,
				})
				oldIdx++
				newIdx++
			}
		}
	}

	// Add remaining lines
	for oldIdx < len(oldLines) {
		result = append(result, DiffLine{
			Type:    "removed",
			Content: oldLines[oldIdx],
			LineNum: oldIdx + 1,
		})
		oldIdx++
	}

	for newIdx < len(newLines) {
		result = append(result, DiffLine{
			Type:    "added",
			Content: newLines[newIdx],
			LineNum: newIdx + 1,
		})
		newIdx++
	}

	return result
}

// diffView component for comparing two versions.
type diffView struct {
	app.Compo

	versions         []Version
	selectedVersion1 int
	selectedVersion2 int
	diffLines        []DiffLine
	comment          bool
	minimal          bool
	renderErr        error
}

func (d *diffView) OnMount(_ app.Context) {
	minVersions := 2
	if len(d.versions) >= minVersions {
		d.selectedVersion1 = 0
		d.selectedVersion2 = 1
		d.updateDiff()
	}
}

func (d *diffView) updateDiff() {
	if d.selectedVersion1 >= len(d.versions) || d.selectedVersion2 >= len(d.versions) {
		d.renderErr = errors.New("invalid version selection")

		return
	}

	version1 := &d.versions[d.selectedVersion1]
	version2 := &d.versions[d.selectedVersion2]

	yaml1, err := version1.generateYAMLDetails(d.comment, d.minimal)
	if err != nil {
		d.renderErr = fmt.Errorf("failed to generate YAML for version 1: %w", err)

		return
	}

	yaml2, err := version2.generateYAMLDetails(d.comment, d.minimal)
	if err != nil {
		d.renderErr = fmt.Errorf("failed to generate YAML for version 2: %w", err)

		return
	}

	d.diffLines = simpleDiff(yaml1, yaml2)
	d.renderErr = nil
}

func (d *diffView) onVersion1Change(ctx app.Context, _ app.Event) {
	selectedValue := ctx.JSSrc().Get("value").String()
	for i, version := range d.versions {
		if version.Version == selectedValue {
			d.selectedVersion1 = i

			break
		}
	}

	d.updateDiff()
}

func (d *diffView) onVersion2Change(ctx app.Context, _ app.Event) {
	selectedValue := ctx.JSSrc().Get("value").String()
	for i, version := range d.versions {
		if version.Version == selectedValue {
			d.selectedVersion2 = i

			break
		}
	}

	d.updateDiff()
}

func (d *diffView) onCommentToggle(_ app.Context, _ app.Event) {
	d.comment = !d.comment
	d.updateDiff()
}

func (d *diffView) onMinimalToggle(_ app.Context, _ app.Event) {
	d.minimal = !d.minimal
	d.updateDiff()
}

func (d *diffView) Render() app.UI {
	minVersions := 2
	if len(d.versions) < minVersions {
		return app.Div()
	}

	versionField := func(id, label string, selected int, onChange app.EventHandler) app.UI {
		return app.Div().Class("field").Body(
			app.Label().Class("label").For(id).Body(icon("tag"), app.Text(label)),
			app.Select().Class("select").ID(id).OnChange(onChange).Body(
				app.Range(d.versions).Slice(func(i int) app.UI {
					option := app.Option().Value(d.versions[i].Version).Text(d.versions[i].Version)
					if i == selected {
						option = option.Selected(true)
					}

					return option
				}),
			),
		)
	}

	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-head").Body(
			icon("git-compare"),
			app.H2().Class("grow").Text("Version Diff"),
			app.Button().Class("btn btn-ghost btn-sm").
				Type("button").
				DataSets(map[string]any{"toggle": "collapse", "target": "#diff-collapse"}).
				Aria("expanded", "false").
				Aria("controls", "diff-collapse").Body(
				icon("eye"),
				app.Text("Show Diff"),
			),
		),

		app.Div().Class("collapse").ID("diff-collapse").Body(
			app.Div().Body(
				app.Div().Class("card-body stack gap-4").Body(
					app.Div().Class("cols").Body(
						versionField("version1-select", "Version A", d.selectedVersion1, d.onVersion1Change),
						versionField("version2-select", "Version B", d.selectedVersion2, d.onVersion2Change),
						app.Label().Class("switch").For("diff-comments").Body(
							app.Input().Type("checkbox").ID("diff-comments").OnClick(d.onCommentToggle),
							app.Span().Class("switch-track"),
							app.Span().Text("Comments"),
						),
						app.Label().Class("switch").For("diff-minimal").Body(
							app.Input().Type("checkbox").ID("diff-minimal").OnClick(d.onMinimalToggle),
							app.Span().Class("switch-track"),
							app.Span().Text("Minimal"),
						),
					),

					app.If(d.renderErr != nil, func() app.UI {
						return app.Div().Class("alert alert-danger").Body(
							icon("alert-circle"),
							app.Text(d.renderErr.Error()),
						)
					}).Else(func() app.UI {
						return app.Div().Class("diff").Body(
							app.Range(d.diffLines).Slice(func(i int) app.UI {
								line := d.diffLines[i]

								var lineClass, marker string

								switch line.Type {
								case "added":
									lineClass, marker = "diff-add", "plus"
								case "removed":
									lineClass, marker = "diff-del", "minus"
								default:
									lineClass, marker = "diff-same", ""
								}

								return app.Div().Class("diff-line "+lineClass).Body(
									app.If(marker != "", func() app.UI { return icon(marker) }),
									app.Text(line.Content),
								)
							}),
						)
					}),
				),
			),
		),
	)
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
	Enums       []string
}

func (h *crdView) buildError(err error) app.UI {
	return app.Div().Class("alert alert-danger fade-in").Role("alert").Body(
		icon("alert-triangle", "icon-lg"),
		app.Div().Class("grow").Body(
			app.P().Class("strong mb-2").Text("Failed to process CRD"),
			app.P().Text(err.Error()),
		),
	)
}

func (h *crdView) OnNav(ctx app.Context) {
	if !strings.Contains(ctx.Page().URL().String(), "share") {
		return
	}

	u := ctx.Page().URL().Query().Get("url")
	if u == "" {
		h.preRenderErr = errors.New(
			"url parameter has to be define in the following format: " +
				"/share?url=https://example.com/crd.yaml")

		return
	}

	if _, err := url.Parse(u); err != nil {
		h.preRenderErr = fmt.Errorf("invald url provided in query: %w", err)

		return
	}

	// authentication is not available here.
	f := fetcher.NewFetcher(http.DefaultClient, "", "", "")

	content, err := f.Fetch(u)
	if err != nil {
		h.preRenderErr = err

		return
	}

	crd, err := renderCRDContent(content)
	if err != nil {
		h.preRenderErr = err

		return
	}

	// Store the original URL for shareable link
	h.originalURL = u
	h.crds = append(h.crds, crd)
}

// The Render method is where the component appearance is defined.
func (h *crdView) Render() app.UI {
	if h.preRenderErr != nil {
		return h.buildError(h.preRenderErr)
	}

	versions := make([]Version, 0)

	for _, schemaType := range h.crds {
		for _, version := range schemaType.Versions {
			v, err := h.generate(schemaType, version.Schema, schemaType.Kind+"-"+version.Name)
			if err != nil {
				return h.buildError(err)
			}

			versions = append(versions, v)
		}

		// Parse validation instead.
		if len(schemaType.Versions) == 0 && schemaType.Validation != nil {
			v, err := h.generate(schemaType, schemaType.Validation.Schema, schemaType.Kind+"-"+schemaType.Validation.Name)
			if err != nil {
				return h.buildError(err)
			}

			versions = append(versions, v)
		}
	}

	wrapper := app.Div()
	container := app.Div().Class("container mt-4 mb-5")

	// Build the content array
	var content []app.UI //nolint:prealloc // avoid preallocation

	// Add diff view if there are multiple versions
	minVersion := 2
	if len(versions) >= minVersion {
		content = append(content, &diffView{versions: versions})
	}

	// Add version cards
	for i, version := range versions {
		content = append(content, app.Div().Class("card mb-5").Body(
			app.Div().Class("card-head").Body(
				icon("box"),
				app.Div().Class("grow").Body(
					app.H2().Text(version.Kind),
					app.Div().Class("card-sub").Text(fmt.Sprintf("%s/%s", version.Group, version.Version)),
				),
				app.Span().Class("badge").Text(version.Version),
			),

			app.If(version.Description != "", func() app.UI {
				return app.Div().Class("card-body").Body(
					app.Div().Class("muted").Body(parseDescriptionElements(version.Description)...),
				)
			}),

			app.Div().Class("card-body").Body(
				app.Div().Class("between mb-3").Body(
					app.P().Class("section-title").Body(
						icon("file-code"),
						app.Text("Generated YAML Sample"),
					),
					app.Button().Class("btn btn-ghost btn-sm").
						Type("button").
						DataSets(map[string]any{
							"toggle": "collapse",
							"target": "#yaml-collapse-" + version.Version,
						}).
						Aria("expanded", "false").
						Aria("controls", "yaml-collapse-"+version.Version).Body(
						icon("eye"),
						app.Text("View Sample"),
					),
				),
				app.Div().Class("collapse").ID("yaml-collapse-"+version.Version).Body(
					app.Div().Body(&detailsView{version: &version}),
				),
			),

			app.Div().Class("card-body").Body(
				app.P().Class("section-title mb-3").Body(
					icon("network"),
					app.Text("Schema Properties"),
				),
				app.Div().ID("properties-accordion-"+version.Version).Body(
					render(app.Div(), version.Properties, "properties-accordion-"+version.Version),
				),
			),
		))
		_ = i // avoid unused variable
	}

	container.Body(content...)

	return wrapper.Body(
		&header{titleOnClick: h.navigateBackOnClick, hidden: false, shareURL: h.originalURL, shareOnClick: h.onShareClick},
		container,
	)
}

func (h *crdView) generate(crd *pkg.SchemaType, properties *v1beta1.JSONSchemaProps, name string) (Version, error) {
	out, err := parseCRD(properties.Properties, name, pkg.RootRequiredFields, h.minimal)
	if err != nil {
		return Version{}, err
	}

	return Version{
		Version:     name,
		Schema:      properties.Properties,
		Properties:  out,
		Kind:        crd.Kind,
		Group:       crd.Group,
		Description: properties.Description,
	}, nil
}

func (h *crdView) onShareClick(ctx app.Context, _ app.Event) {
	if h.originalURL == "" {
		return
	}

	pageURL := ctx.Page().URL()

	protocol := "https"
	if pageURL.Scheme != "" {
		protocol = pageURL.Scheme
	}

	shareURL := fmt.Sprintf("%s://%s/share?url=%s", protocol, pageURL.Host, url.QueryEscape(h.originalURL))

	// Use JavaScript clipboard API
	app.Window().Get("navigator").Get("clipboard").Call("writeText", shareURL)
}

func render(d app.UI, p []*Property, accordionID string) app.UI {
	elements := make([]app.UI, 0, len(p))
	for _, prop := range p {
		badges := []app.UI{
			app.Span().Class("badge").Text(prop.Type),
		}

		if prop.Required {
			badges = append(badges, app.Span().Class("badge badge-required").Text("Required"))
		}

		if prop.Enums != nil {
			badges = append(badges, app.Span().Class("badge badge-enum").Text("Enum"))
		}

		if prop.Format != "" {
			badges = append(badges, app.Span().Class("badge").Text("Format: "+prop.Format))
		}

		if prop.Default != "" {
			badges = append(badges, app.Span().Class("badge").Text("Default: "+prop.Default))
		}

		if prop.Patterns != "" {
			badges = append(badges, app.Span().Class("badge").Text("Pattern: "+prop.Patterns))
		}

		headerContainer := app.Div().Class("grow").Body(
			app.Div().Class("prop-head").Body(
				append([]app.UI{app.Span().Class("prop-name").Text(prop.Name)}, badges...)...,
			),
			app.If(prop.Description != "", func() app.UI {
				return app.Div().Class("prop-desc").Body(parseDescriptionElements(prop.Description)...)
			}),
			app.If(prop.Enums != nil, func() app.UI {
				return app.Div().Class("prop-desc").Body(
					app.Span().Class("strong").Text("Allowed values: "),
					app.Code().Text(strings.Join(prop.Enums, ", ")),
				)
			}),
		)

		if len(prop.Properties) > 0 {
			targetID := "node-" + prop.Name + accordionID

			elements = append(elements,
				app.Div().Class("node").Body(
					app.Button().
						Class("node-toggle").
						Type("button").
						DataSets(map[string]any{
							"toggle": "collapse",
							"target": "#" + targetID,
						}).
						Aria("expanded", "false").
						Aria("controls", targetID).
						Body(headerContainer),

					app.Div().Class("collapse").ID(targetID).
						DataSet("parent", "#"+accordionID).Body(
						app.Div().Body(
							app.Div().Class("node-body").Body(
								render(app.Div(), prop.Properties, targetID),
							),
						),
					),
				),
			)

			continue
		}

		elements = append(elements, app.Div().Class("node node-leaf").Body(headerContainer))
	}

	// add all the elements and return the div
	//nolint: gocritic // type switch
	switch t := d.(type) {
	case app.HTMLDiv:
		t.Body(elements...)
		d = t
	}

	return d
}

// parseCRD takes the properties and constructs a linked list out of the embedded properties that the recursive
// template can call and construct linked divs.
func parseCRD(properties map[string]v1beta1.JSONSchemaProps, version string, requiredList []string, minimal bool) ([]*Property, error) {
	sortedKeys := make([]string, 0, len(properties))
	output := make([]*Property, 0, len(properties))

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

		if slices.Contains(requiredList, k) {
			required = true
		}

		// skip if only minimal is required
		if minimal && !required {
			continue
		}

		var enums []string

		if v.Enum != nil {
			for _, e := range v.Enum {
				enums = append(enums, string(e.Raw))
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
			Enums:       enums,
		}
		if v.Default != nil {
			p.Default = string(v.Default.Raw)
		}

		switch {
		case len(properties[k].Properties) > 0:
			requiredList = v.Required

			out, err := parseCRD(properties[k].Properties, version, requiredList, minimal)
			if err != nil {
				return nil, err
			}

			p.Properties = out
		case properties[k].Type == "array" && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0:
			requiredList = v.Required

			out, err := parseCRD(properties[k].Items.Schema.Properties, version, properties[k].Items.Schema.Required, minimal)
			if err != nil {
				return nil, err
			}

			p.Properties = out
		case properties[k].AdditionalProperties != nil && properties[k].AdditionalProperties.Schema != nil:
			requiredList = v.Required

			out, err := parseCRD(properties[k].AdditionalProperties.Schema.Properties, version, properties[k].AdditionalProperties.Schema.Required, minimal)
			if err != nil {
				return nil, err
			}

			p.Properties = out
		}

		output = append(output, p)
	}

	return output, nil
}
