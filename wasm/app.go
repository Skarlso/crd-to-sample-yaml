// Package main contains the application main code for the WASM codebase.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

// timeout after 2 seconds.
const timeout = 2000

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
	return app.Div().Body(
		// Options row
		app.Div().Class("row g-3 mb-4").Body(
			app.Div().Class("col-md-6").Body(
				app.Div().Class("form-check form-switch").Body(
					app.Input().Class("form-check-input").Type("checkbox").ID("enable-comments-"+v.version.Version).OnClick(v.OnCheckComment),
					app.Label().Class("form-check-label").For("enable-comments-"+v.version.Version).Body(
						app.Strong().Text("Include Comments"),
						app.Br(),
						app.Small().Class("text-muted").Text("Add helpful field descriptions"),
					),
				),
			),
			app.Div().Class("col-md-6").Body(
				app.Div().Class("form-check form-switch").Body(
					app.Input().Class("form-check-input").Type("checkbox").ID("enable-minimal-"+v.version.Version).OnClick(v.OnCheckMinimal),
					app.Label().Class("form-check-label").For("enable-minimal-"+v.version.Version).Body(
						app.Strong().Text("Minimal Output"),
						app.Br(),
						app.Small().Class("text-muted").Text("Show only required fields"),
					),
				),
			),
		),

		// YAML output container
		app.Div().Class("position-relative").Body(
			// Copy button
			app.Button().Class("copy-btn").
				ID("copy-btn-"+v.version.Version).
				DataSet("clipboard-target", "#yaml-sample-"+v.version.Version).
				Title("Copy to clipboard").
				OnClick(v.onCopyClick).Body(
				app.I().Class("fas fa-copy"),
			),

			// YAML content
			app.Pre().Class("yaml-text").Body(
				app.Code().ID("yaml-sample-"+v.version.Version).Body(app.If(v.renderErr != nil, func() app.UI {
					return app.Div().Class("text-danger").Body(
						app.I().Class("fas fa-exclamation-circle me-2"),
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
	app.Window().Get("navigator").Get("clipboard").Call("writeText", content).Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		// Show success feedback
		btn := ctx.JSSrc()
		originalHTML := btn.Get("innerHTML").String()
		btn.Set("innerHTML", `<i class="fas fa-check"></i>`)
		btn.Get("classList").Call("add", "btn-success")
		btn.Get("classList").Call("remove", "copy-btn")

		app.Window().Call("setTimeout", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			btn.Set("innerHTML", originalHTML)
			btn.Get("classList").Call("remove", "btn-success")
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
	if err := parser.ParseProperties(v.Version, buf, v.Schema, pkg.RootRequiredFields); err != nil {
		return "", err
	}

	return buf.String(), nil
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
		app.Div().Class("d-flex align-items-start").Body(
			app.Div().Class("me-3").Body(
				app.I().Class("fas fa-exclamation-triangle fa-2x text-danger"),
			),
			app.Div().Class("flex-grow-1").Body(
				app.H4().Class("alert-heading mb-3").Text("Failed to process CRD"),
				app.P().Class("mb-0").Text(err.Error()),
			),
			app.Button().Class("closebtn").Type("button").Body(
				app.I().Class("fas fa-times"),
			),
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

	wrapper := app.Div().Class("main-container")
	container := app.Div().Class("container mt-4")
	container.Body(app.Range(versions).Slice(func(i int) app.UI {
		version := versions[i]

		return app.Div().Class("card mb-5").Body(
			// Version header
			app.Div().Class("card-header bg-primary text-white").Body(
				app.Div().Class("d-flex justify-content-between align-items-center").Body(
					app.Div().Body(
						app.H2().Class("h4 mb-1 d-flex align-items-center").Body(
							app.I().Class("fas fa-cube me-2"),
							app.Text(version.Kind),
						),
						app.Small().Class("opacity-75").Text(fmt.Sprintf("%s/%s", version.Group, version.Version)),
					),
					app.Div().Body(
						app.Span().Class("badge bg-light text-dark px-3 py-2").Body(
							app.I().Class("fas fa-tag me-1"),
							app.Text(version.Version),
						),
					),
				),
			),

			// Version description
			app.If(version.Description != "", func() app.UI {
				return app.Div().Class("card-body border-bottom").Body(
					app.P().Class("text-muted mb-0").Body(
						app.I().Class("fas fa-info-circle me-2"),
						app.Text(version.Description),
					),
				)
			}),

			// YAML Sample Section
			app.Div().Class("card-body").Body(
				app.Div().Class("d-flex justify-content-between align-items-center mb-3").Body(
					app.H5().Class("mb-0 d-flex align-items-center").Body(
						app.I().Class("fas fa-file-code me-2 text-success"),
						app.Text("Generated YAML Sample"),
					),
					app.Button().Class("btn btn-outline-primary btn-sm").
						Type("button").
						DataSet("bs-toggle", "collapse").
						DataSet("bs-target", "#yaml-collapse-"+version.Version).
						Aria("expanded", "false").
						Aria("controls", "yaml-collapse-"+version.Version).Body(
						app.I().Class("fas fa-eye me-1"),
						app.Text("View Sample"),
					),
				),
				app.Div().Class("collapse").ID("yaml-collapse-"+version.Version).Body(
					&detailsView{version: &version},
				),
			),

			// Properties Schema Section
			app.Div().Class("card-body border-top").Body(
				app.H5().Class("mb-3 d-flex align-items-center").Body(
					app.I().Class("fas fa-sitemap me-2 text-info"),
					app.Text("Schema Properties"),
				),
				app.Div().Class("accordion").ID("properties-accordion-"+version.Version).Body(
					render(app.Div().Class("accordion-item"), version.Properties, "properties-accordion-"+version.Version),
				),
			),
		)
	}))

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
		// Property header with modern styling
		headerElements := []app.UI{
			app.Div().Class("col-auto").Body(
				app.H6().Class("mb-1 fw-bold text-primary").Text(prop.Name),
			),
			app.Div().Class("col-auto").Body(
				app.Span().Class("property-type").Text(prop.Type),
			),
		}

		// Add badges for special properties
		badges := []app.UI{}
		if prop.Required {
			badges = append(badges, app.Span().Class("property-type property-required me-1").Text("Required"))
		}
		if prop.Enums != nil {
			badges = append(badges, app.Span().Class("property-type property-enum me-1").Text("Enum"))
		}
		if prop.Format != "" {
			badges = append(badges, app.Span().Class("badge bg-info me-1").Text("Format: "+prop.Format))
		}
		if prop.Default != "" {
			badges = append(badges, app.Span().Class("badge bg-secondary me-1").Text("Default: "+prop.Default))
		}
		if prop.Patterns != "" {
			badges = append(badges, app.Span().Class("badge bg-warning text-dark me-1").Text("Pattern: "+prop.Patterns))
		}
		if len(badges) > 0 {
			headerElements = append(headerElements, app.Div().Class("col-12 mt-2").Body(badges...))
		}

		headerContainer := app.Div().Class("container-fluid").Body(
			app.Div().Class("row align-items-center").Body(
				headerElements...,
			),
			app.If(prop.Description != "", func() app.UI {
				return app.Div().Class("row mt-2").Body(
					app.Div().Class("col-12").Body(
						app.P().Class("text-muted mb-0 small").Text(prop.Description),
					),
				)
			}),
			app.If(prop.Enums != nil, func() app.UI {
				return app.Div().Class("row mt-2").Body(
					app.Div().Class("col-12").Body(
						app.Strong().Class("small text-info").Text("Allowed values: "),
						app.Code().Class("small").Text(strings.Join(prop.Enums, ", ")),
					),
				)
			}),
		)

		// Create header element
		var header app.UI

		if len(prop.Properties) > 0 {
			// This property has children - make it collapsible
			targetID := "accordion-collapse-for-" + prop.Name + accordionID
			button := app.Button().
				ID("accordion-button-id-"+prop.Name+accordionID).
				Class("accordion-button").
				Type("button").
				DataSets(map[string]any{
					"bs-toggle": "collapse",
					"bs-target": "#" + targetID,
				}).
				Aria("expanded", "false").
				Aria("controls", targetID).
				Body(headerContainer)

			header = app.H2().Class("accordion-header").Body(button)
			elements = append(elements, header)

			// Add collapsible content
			accordionDiv := app.Div().Class("accordion-collapse collapse").ID(targetID).DataSet("bs-parent", "#"+accordionID)
			accordionBody := app.Div().Class("accordion-body")

			element := render(app.Div().ID(prop.Name).Class("accordion-item"), prop.Properties, targetID)
			accordionBody.Body(element)
			accordionDiv.Body(accordionBody)
			elements = append(elements, accordionDiv)
		} else {
			// This property has no children - just show as a simple item
			header = app.Div().Class("accordion-item-static border rounded mb-2 p-3 bg-light").Body(headerContainer)
			elements = append(elements, header)
		}
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
		for _, item := range requiredList {
			if item == k {
				required = true

				break
			}
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
