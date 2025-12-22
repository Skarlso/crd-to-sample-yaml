package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

const maximumBytes = 200 * 1000 // 200KB

// index is the main page that contains the textarea and the submit button.
// It will also deal with navigation and user submits.
type index struct {
	app.Compo

	crds      []*pkg.SchemaType
	isMounted bool
	err       error
	comments  bool
	minimal   bool
	lastURL   string
}

func (i *index) buildError() app.UI {
	return app.Div().Class("alert alert-danger fade-in").Role("alert").Body(
		app.Div().Class("d-flex align-items-start").Body(
			app.Div().Class("me-3").Body(
				app.I().Class("fas fa-exclamation-triangle fa-2x text-danger"),
			),
			app.Div().Class("flex-grow-1").Body(
				app.H4().Class("alert-heading d-flex align-items-center mb-3").Body(
					app.Text("Something went wrong!"),
				),
				app.P().Class("mb-3").Text(i.err.Error()),
				app.Hr(),
				app.P().Class("mb-0 small").Body(
					app.Strong().Text("Suggestions:"),
					app.Br(),
					app.Text("• Check that your CRD is valid YAML"),
					app.Br(),
					app.Text("• Ensure the URL is accessible"),
					app.Br(),
					app.Text("• Verify authentication credentials if required"),
				),
			),
			app.Button().Class("closebtn").Type("button").OnClick(i.dismissError).Body(
				app.I().Class("fas fa-times"),
			),
		),
	)
}

func (i *index) dismissError(_ app.Context, _ app.Event) {
	i.err = nil
}

// header is the site header.
type header struct {
	app.Compo

	titleOnClick func(ctx app.Context, _ app.Event)
	hidden       bool
	shareURL     string
	shareOnClick func(ctx app.Context, _ app.Event)
}

func (h *header) Render() app.UI {
	return app.Nav().Class("navbar navbar-expand-lg").Body(
		app.Div().Class("container-fluid").Body(
			// Brand
			app.Button().Class("navbar-brand d-flex align-items-center btn btn-link border-0 text-decoration-none").OnClick(h.titleOnClick).Body(
				app.I().Class("fas fa-code me-2").Style("font-size", "1.5rem"),
				app.Span().Text("CRD to YAML"),
			),

			// Mobile toggle button
			app.Button().Class("navbar-toggler").Type("button").
				DataSet("bs-toggle", "collapse").
				DataSet("bs-target", "#navbarNav").
				Aria("controls", "navbarNav").
				Aria("expanded", "false").
				Aria("label", "Toggle navigation").Body(
				app.Span().Class("navbar-toggler-icon"),
			),

			// Collapsible navbar content
			app.Div().Class("collapse navbar-collapse").ID("navbarNav").Body(
				app.Ul().Class("navbar-nav ms-auto").Body(
					// Back button
					app.Li().Class("nav-item").Hidden(h.hidden).Body(
						app.Button().Class("nav-link icon-btn me-2 btn btn-link border-0").
							OnClick(h.titleOnClick).
							Title("Back to Home").Body(
							app.I().Class("fas fa-arrow-left"),
						),
					),

					// Share button
					app.Li().Class("nav-item").Hidden(h.hidden || h.shareURL == "").Body(
						app.Button().Class("nav-link icon-btn me-2 btn btn-link border-0").
							OnClick(h.shareOnClick).
							Title("Share this CRD").Body(
							app.I().Class("fas fa-share-alt"),
						),
					),

					// GitHub link
					app.Li().Class("nav-item").Body(
						app.A().Class("nav-link icon-btn").
							Href("https://github.com/Skarlso/crd-to-sample-yaml").
							Target("_blank").
							Title("View on GitHub").Body(
							app.I().Class("fab fa-github"),
						),
					),
				),
			),
		),
	)
}

// textarea is the textarea component that is used to supply the CRD content.
type textarea struct {
	app.Compo
}

func (t *textarea) Render() app.UI {
	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-header").Body(
			app.H5().Class("card-title mb-0 d-flex align-items-center").Body(
				app.I().Class("fas fa-file-code me-2 text-primary"),
				app.Text("CRD Definition"),
			),
		),
		app.Div().Class("card-body").Body(
			app.Div().Class("form-floating").Body(
				app.Textarea().
					Class("form-control").
					ID("crd_data").
					Name("crd_data").
					Placeholder("Paste your Kubernetes CRD definition here...").
					Style("min-height", "200px"),
				app.Label().For("crd_data").Text("Paste your Kubernetes CRD definition here..."),
			),
			app.Div().Class("form-text mt-2").Body(
				app.I().Class("fas fa-info-circle me-1 text-info"),
				app.Text("Supports YAML format. Maximum size: 200KB"),
			),
		),
	)
}

// input is the input button.
type input struct {
	app.Compo
}

func (i *input) Render() app.UI {
	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-header").Body(
			app.H5().Class("card-title mb-0 d-flex align-items-center").Body(
				app.I().Class("fas fa-link me-2 text-success"),
				app.Text("Fetch from URL"),
			),
		),
		app.Div().Class("card-body").Body(
			app.Div().Class("row g-3").Body(
				// URL input
				app.Div().Class("col-12").Body(
					app.Div().Class("form-floating").Body(
						app.Input().
							Class("form-control url_to_crd").
							Type("url").
							ID("url_to_crd").
							Name("url_to_crd").
							Placeholder("https://example.com/crd.yaml"),
						app.Label().For("url_to_crd").Text("CRD URL"),
					),
				),

				// Authentication section
				app.Div().Class("col-12").Body(
					app.Div().Class("border rounded p-3 bg-light").Body(
						app.H6().Class("text-muted mb-3 d-flex align-items-center").Body(
							app.I().Class("fas fa-shield-alt me-2"),
							app.Text("Authentication (Optional)"),
						),
						app.Div().Class("row g-2").Body(
							app.Div().Class("col-md-4").Body(
								app.Div().Class("form-floating").Body(
									app.Input().
										Class("form-control url_username").
										Type("text").
										ID("url_username").
										Placeholder("Username"),
									app.Label().For("url_username").Text("Username"),
								),
							),
							app.Div().Class("col-md-4").Body(
								app.Div().Class("form-floating").Body(
									app.Input().
										Class("form-control url_password").
										Type("password").
										ID("url_password").
										Placeholder("Password"),
									app.Label().For("url_password").Text("Password"),
								),
							),
							app.Div().Class("col-md-4").Body(
								app.Div().Class("form-floating").Body(
									app.Input().
										Class("form-control url_token").
										Type("password").
										ID("url_token").
										Placeholder("Token"),
									app.Label().For("url_token").Text("Access Token"),
								),
							),
						),
					),
				),
			),
			app.Div().Class("form-text mt-2").Body(
				app.I().Class("fas fa-info-circle me-1 text-info"),
				app.Text("Supports public URLs and authenticated repositories (GitHub, GitLab, etc.)"),
			),
		),
	)
}

// form is the form in which the user will submit their input.
type form struct {
	app.Compo

	formHandler         app.EventHandler
	checkHandlerMinimal app.EventHandler
	checkHandlerComment app.EventHandler
}

func (f *form) Render() app.UI {
	return app.Div().Class("container-fluid").Body(
		app.Div().Class("row justify-content-center").Body(
			app.Div().Class("col-lg-10 col-xl-8").Body(
				&textarea{},
				app.Div().Class("text-center mb-3").Body(
					app.Span().Class("badge bg-secondary px-3 py-2").Body(
						app.I().Class("fas fa-exchange-alt me-2"),
						app.Text("OR"),
					),
				),
				&input{},
				&checkBox{checkHandlerComment: f.checkHandlerComment, checkHandlerMinimal: f.checkHandlerMinimal},
				app.Div().Class("d-grid gap-2 mt-4").Body(
					app.Button().Class("btn btn-primary btn-lg").Type("submit").
						ID("submit-btn").
						OnClick(f.formHandler).Body(
						app.I().Class("fas fa-magic me-2"),
						app.Text("Generate YAML Sample"),
					),
				),
			),
		),
	)
}

func renderCRDContent(content []byte) (*pkg.SchemaType, error) {
	content, err := sanitize.Sanitize(content)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize content: %w", err)
	}

	crd := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into custom resource definition: %w", err)
	}

	schemaType, err := pkg.ExtractSchemaType(crd)
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema type: %w", err)
	}

	if schemaType == nil {
		return nil, nil
	}

	return schemaType, nil
}

func (i *index) OnClick(ctx app.Context, _ app.Event) {
	// Add loading state to button
	submitBtn := app.Window().GetElementByID("submit-btn")
	submitBtn.Set("disabled", true)
	submitBtn.Get("classList").Call("add", "btn-loading")
	originalText := submitBtn.Get("innerHTML").String()
	submitBtn.Set("innerHTML", `<span class="loading-spinner me-2"></span>Processing...`)

	defer func() {
		submitBtn.Set("disabled", false)
		submitBtn.Get("classList").Call("remove", "btn-loading")
		submitBtn.Set("innerHTML", originalText)
	}()

	ta := app.Window().GetElementByID("crd_data").Get("value")
	if v := ta.String(); v != "" {
		if len(v) > maximumBytes {
			i.err = errors.New("content exceeds maximum length of 200KB")

			return
		}

		crd, err := renderCRDContent([]byte(v))
		if err != nil {
			i.err = err

			return
		}

		i.crds = append(i.crds, crd)

		// Scroll to top after successful CRD processing
		app.Window().Call("scrollTo", 0, 0)

		return
	}

	username := app.Window().GetElementByID("url_username").Get("value")
	password := app.Window().GetElementByID("url_password").Get("value")
	token := app.Window().GetElementByID("url_token").Get("value")

	inp := app.Window().GetElementByID("url_to_crd").Get("value")
	if inp.String() == "" {
		return
	}

	f := fetcher.NewFetcher(http.DefaultClient, username.String(), password.String(), token.String())

	content, err := f.Fetch(inp.String())
	if err != nil {
		i.err = fmt.Errorf("failed to fetch CRD content: %w", err)

		return
	}

	if len(content) > maximumBytes {
		i.err = errors.New("content exceeds maximum length of 200KB")

		return
	}

	// Store the URL for shareable link
	i.lastURL = inp.String()

	crd, err := renderCRDContent(content)
	if err != nil {
		i.err = err

		return
	}

	i.crds = append(i.crds, crd)

	// Scroll to top after successful CRD processing
	app.Window().Call("scrollTo", 0, 0)
}

// checkBox defines if comments should be generated for the sample YAML output.
type checkBox struct {
	app.Compo

	checkHandlerComment app.EventHandler
	checkHandlerMinimal app.EventHandler
}

func (c *checkBox) Render() app.UI {
	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-header").Body(
			app.H5().Class("card-title mb-0 d-flex align-items-center").Body(
				app.I().Class("fas fa-cogs me-2 text-warning"),
				app.Text("Output Options"),
			),
		),
		app.Div().Class("card-body").Body(
			app.Div().Class("row g-3").Body(
				app.Div().Class("col-md-6").Body(
					app.Div().Class("form-check form-switch").Body(
						app.Input().Class("form-check-input").Type("checkbox").ID("enable-comments").OnClick(c.checkHandlerComment),
						app.Label().Class("form-check-label").For("enable-comments").Body(
							app.Strong().Text("Include Comments"),
							app.Br(),
							app.Small().Class("text-muted").Text("Add helpful comments to the generated YAML"),
						),
					),
				),
				app.Div().Class("col-md-6").Body(
					app.Div().Class("form-check form-switch").Body(
						app.Input().Class("form-check-input").Type("checkbox").ID("enable-minimal").OnClick(c.checkHandlerMinimal),
						app.Label().Class("form-check-label").For("enable-minimal").Body(
							app.Strong().Text("Minimal Output"),
							app.Br(),
							app.Small().Class("text-muted").Text("Show only required fields"),
						),
					),
				),
			),
		),
	)
}

func (i *index) OnCheckComment(_ app.Context, _ app.Event) {
	i.comments = !i.comments
}

func (i *index) OnCheckMinimal(_ app.Context, _ app.Event) {
	i.minimal = !i.minimal
}

func (i *index) OnMount(_ app.Context) {
	i.isMounted = true
}

func (i *index) NavBackOnClick(_ app.Context, _ app.Event) {
	i.crds = nil
	i.minimal = false
	i.comments = false
	i.lastURL = ""
}

type editView struct {
	app.Compo

	content []byte
}

func (e *editView) OnInput(ctx app.Context, _ app.Event) {
	content := ctx.JSSrc().Get("value").String()

	crd := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(content), crd); err != nil {
		e.content = []byte("invalid CRD content")

		return
	}

	schemaType, err := pkg.ExtractSchemaType(crd)
	if err != nil {
		e.content = []byte("invalid CRD content")

		return
	}

	e.content = nil

	parser := pkg.NewParser(schemaType.Group, schemaType.Kind, false, false, false)
	for _, version := range schemaType.Versions {
		e.content = append(e.content, []byte("---\n")...)

		var buffer []byte

		buf := bytes.NewBuffer(buffer)
		err := parser.ParseProperties(version.Name, buf, version.Schema.Properties, pkg.RootRequiredFields)
		if err != nil {
			e.content = []byte(err.Error())

			return
		}

		e.content = append(e.content, buf.Bytes()...)
	}
}

func (e *editView) Render() app.UI {
	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-header").Body(
			app.H5().Class("card-title mb-0 d-flex align-items-center").Body(
				app.I().Class("fas fa-edit me-2 text-info"),
				app.Text("Live CRD Editor"),
			),
			app.Small().Class("text-muted").Text("Type your CRD definition and see the YAML output in real-time"),
		),
		app.Div().Class("card-body p-0").Body(
			app.Div().Class("row g-0").Body(
				app.Div().Class("col-md-6 border-end").Body(
					app.Div().Class("p-3").Body(
						app.Label().Class("form-label fw-bold mb-2").Body(
							app.I().Class("fas fa-code me-1"),
							app.Text("CRD Input"),
						),
						app.Textarea().
							Class("form-control border-0").
							Style("height", "400px").
							Style("resize", "none").
							Placeholder("Start typing your CRD definition...").
							ID("input-area").
							OnInput(e.OnInput),
					),
				),
				app.Div().Class("col-md-6").Body(
					app.Div().Class("p-3").Body(
						app.Label().Class("form-label fw-bold mb-2").Body(
							app.I().Class("fas fa-file-alt me-1"),
							app.Text("YAML Output"),
						),
						app.Pre().Class("yaml-text border-0").Style("height", "400px").Style("margin", "0").Body(
							app.Code().Text(string(e.content)),
						),
					),
				),
			),
		),
	)
}

func (i *index) Render() app.UI {
	// Prevent double rendering components.
	if i.isMounted {
		return app.Main().Body(
			app.Div().Class("main-container").Body(func() app.UI {
				if i.err != nil {
					return app.Div().Body(
						&header{titleOnClick: i.NavBackOnClick, hidden: true},
						app.Div().Class("container mt-4").Body(i.buildError()),
					)
				}

				if len(i.crds) > 0 {
					return &crdView{crds: i.crds, comment: i.comments, minimal: i.minimal, originalURL: i.lastURL, navigateBackOnClick: i.NavBackOnClick}
				}

				return app.Div().Body(
					&header{titleOnClick: i.NavBackOnClick, hidden: true},
					app.Div().Class("container mt-4").Body(
						app.Div().Class("row justify-content-center mb-5").Body(
							app.Div().Class("col-lg-8 text-center").Body(
								app.H1().Class("display-4 fw-bold mb-3").Body(
									app.I().Class("fas fa-cube me-3 text-primary"),
									app.Text("CRD to YAML Generator"),
								),
								app.P().Class("lead text-muted mb-4").Text("Transform Kubernetes Custom Resource Definitions into sample YAML configurations with ease"),
								app.Div().Class("d-flex justify-content-center gap-3 mb-4").Body(
									app.Span().Class("badge bg-primary-subtle text-primary px-3 py-2").Body(
										app.I().Class("fas fa-rocket me-1"),
										app.Text("Fast & Easy"),
									),
									app.Span().Class("badge bg-success-subtle text-success px-3 py-2").Body(
										app.I().Class("fas fa-shield-alt me-1"),
										app.Text("Secure"),
									),
									app.Span().Class("badge bg-info-subtle text-info px-3 py-2").Body(
										app.I().Class("fas fa-mobile-alt me-1"),
										app.Text("Responsive"),
									),
								),
							),
						),
						&editView{},
						app.Div().Class("text-center mb-4").Body(
							app.H3().Class("h4 text-muted mb-3").Body(
								app.I().Class("fas fa-upload me-2"),
								app.Text("Upload or Fetch CRD"),
							),
							app.P().Class("text-muted").Text("Choose how you want to provide your CRD definition"),
						),
						&form{formHandler: i.OnClick, checkHandlerComment: i.OnCheckComment, checkHandlerMinimal: i.OnCheckMinimal},
					),
				)
			}()))
	}

	return app.Main()
}
