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
		icon("alert-triangle", "icon-lg"),
		app.Div().Class("grow").Body(
			app.P().Class("strong mb-2").Text("Something went wrong"),
			app.P().Class("mb-3").Text(i.err.Error()),
			app.P().Class("small").Body(
				app.Text("Check the CRD is valid YAML, the URL is reachable, "),
				app.Text("and any credentials are correct."),
			),
		),
		app.Button().Class("alert-close").Type("button").Aria("label", "Dismiss").
			OnClick(i.dismissError).Body(icon("x")),
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
	return app.Nav().Class("navbar").Body(
		app.Button().Class("brand").OnClick(h.titleOnClick).Body(
			icon("code"),
			app.Span().Text("CRD to YAML"),
		),

		app.Div().Class("nav-actions").Body(
			app.If(!h.hidden, func() app.UI {
				return app.Button().Class("btn-icon").Type("button").
					OnClick(h.titleOnClick).Title("Back to home").Body(icon("arrow-left"))
			}),
			app.If(!h.hidden && h.shareURL != "", func() app.UI {
				return app.Button().Class("btn-icon").Type("button").
					OnClick(h.shareOnClick).Title("Share this CRD").Body(icon("share"))
			}),
			app.Button().Class("btn-icon").Type("button").
				DataSet("theme-toggle", "true").
				Title("Toggle light and dark theme").
				Aria("label", "Toggle theme").Body(icon("contrast")),
			app.A().Class("btn-icon").
				Href("https://github.com/Skarlso/crd-to-sample-yaml").
				Target("_blank").
				Title("View on GitHub").Body(icon("github")),
		),
	)
}

// textarea is the textarea component that is used to supply the CRD content.
type textarea struct {
	app.Compo
}

func (t *textarea) Render() app.UI {
	return app.Div().Class("card mb-4").Body(
		app.Div().Class("card-head").Body(
			icon("file-code"),
			app.H2().Text("CRD Definition"),
		),
		app.Div().Class("card-body stack gap-2").Body(
			app.Textarea().
				Class("textarea").
				ID("crd_data").
				Name("crd_data").
				Aria("label", "CRD definition").
				Placeholder("Paste your Kubernetes CRD definition here..."),
			app.P().Class("hint center").Body(
				icon("info"),
				app.Text("YAML format, up to 200KB"),
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
		app.Div().Class("card-head").Body(
			icon("link"),
			app.H2().Text("Fetch from URL"),
		),
		app.Div().Class("card-body stack gap-4").Body(
			app.Div().Class("field").Body(
				app.Label().Class("label").For("url_to_crd").Text("CRD URL"),
				app.Input().
					Class("input url_to_crd").
					Type("url").
					ID("url_to_crd").
					Name("url_to_crd").
					Placeholder("https://example.com/crd.yaml"),
			),

			app.Details().Class("auth").Body(
				app.Summary().Class("auth-summary").Body(
					icon("shield"),
					app.Text("Authentication (optional)"),
				),
				app.Div().Class("cols mt-3").Body(
					app.Div().Class("field").Body(
						app.Label().Class("label").For("url_username").Text("Username"),
						app.Input().Class("input url_username").Type("text").ID("url_username"),
					),
					app.Div().Class("field").Body(
						app.Label().Class("label").For("url_password").Text("Password"),
						app.Input().Class("input url_password").Type("password").ID("url_password"),
					),
					app.Div().Class("field").Body(
						app.Label().Class("label").For("url_token").Text("Access token"),
						app.Input().Class("input url_token").Type("password").ID("url_token"),
					),
				),
			),

			app.P().Class("hint center").Body(
				icon("info"),
				app.Text("Public URLs and authenticated repositories (GitHub, GitLab, ...)"),
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
	return app.Div().Class("stack").Body(
		&textarea{},
		app.Div().Class("or-divider").Body(app.Span().Text("or")),
		&input{},
		&checkBox{checkHandlerComment: f.checkHandlerComment, checkHandlerMinimal: f.checkHandlerMinimal},
		app.Button().Class("btn btn-primary btn-lg btn-block mt-4").Type("submit").
			ID("submit-btn").
			OnClick(f.formHandler).Body(
			icon("sparkles"),
			app.Text("Generate YAML Sample"),
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
		app.Div().Class("card-head").Body(
			icon("settings"),
			app.H2().Text("Output Options"),
		),
		app.Div().Class("card-body cols").Body(
			app.Label().Class("switch").For("enable-comments").Body(
				app.Input().Type("checkbox").ID("enable-comments").OnClick(c.checkHandlerComment),
				app.Span().Class("switch-track"),
				app.Span().Body(
					app.Span().Class("strong").Text("Include comments"),
					app.Br(),
					app.Span().Class("hint").Text("Add field descriptions to the YAML"),
				),
			),
			app.Label().Class("switch").For("enable-minimal").Body(
				app.Input().Type("checkbox").ID("enable-minimal").OnClick(c.checkHandlerMinimal),
				app.Span().Class("switch-track"),
				app.Span().Body(
					app.Span().Class("strong").Text("Minimal output"),
					app.Br(),
					app.Span().Class("hint").Text("Only required fields"),
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
		app.Div().Class("card-head").Body(
			icon("edit"),
			app.Div().Body(
				app.H2().Text("Live CRD Editor"),
				app.P().Class("hint").Text("Type a CRD and see the YAML update as you go"),
			),
		),
		app.Div().Class("split").Body(
			app.Div().Class("card-body stack gap-2").Body(
				app.P().Class("section-title").Body(icon("code"), app.Text("CRD Input")),
				app.Textarea().
					Class("textarea editor-pane").
					Placeholder("Start typing your CRD definition...").
					ID("input-area").
					Aria("label", "CRD input").
					OnInput(e.OnInput),
			),
			app.Div().Class("card-body stack gap-2").Body(
				app.P().Class("section-title").Body(icon("file-text"), app.Text("YAML Output")),
				app.Pre().Class("code editor-pane").Body(
					app.Code().Text(string(e.content)),
				),
			),
		),
	)
}

func (i *index) Render() app.UI {
	// Prevent double rendering components.
	if i.isMounted {
		return app.Main().Body(
			app.Div().Class("shell").Body(func() app.UI {
				if i.err != nil {
					return app.Div().Body(
						&header{titleOnClick: i.NavBackOnClick, hidden: true},
						app.Div().Class("container mt-4 mb-5").Body(i.buildError()),
					)
				}

				if len(i.crds) > 0 {
					return &crdView{crds: i.crds, comment: i.comments, minimal: i.minimal, originalURL: i.lastURL, navigateBackOnClick: i.NavBackOnClick}
				}

				return app.Div().Body(
					&header{titleOnClick: i.NavBackOnClick, hidden: true},
					app.Div().Class("hero").Body(
						app.H1().Body(icon("box"), app.Text("CRD to YAML Generator")),
						app.P().Text("Turn Kubernetes Custom Resource Definitions into sample YAML"),
						app.Div().Class("hero-tags").Body(
							app.Span().Class("badge badge-plain").Body(icon("zap"), app.Text("Fast & Easy")),
							app.Span().Class("badge badge-plain").Body(icon("shield"), app.Text("Secure")),
							app.Span().Class("badge badge-plain").Body(icon("smartphone"), app.Text("Responsive")),
						),
					),
					app.Div().Class("container mb-5").Body(
						&editView{},
						app.Div().Class("middle mt-4 mb-4").Body(
							app.H2().Class("center gap-2").Style("justify-content", "center").Body(
								icon("upload"),
								app.Text("Upload or Fetch CRD"),
							),
							app.P().Class("muted mt-1").Text("Choose how you want to provide your CRD definition"),
						),
						&form{formHandler: i.OnClick, checkHandlerComment: i.OnCheckComment, checkHandlerMinimal: i.OnCheckMinimal},
					),
				)
			}()))
	}

	return app.Main()
}
