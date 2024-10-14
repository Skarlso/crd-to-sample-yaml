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

	content   []byte
	isMounted bool
	err       error
	comments  bool
	minimal   bool
}

func (i *index) buildError() app.UI {
	return app.Div().Class("alert alert-danger").Role("alert").Body(
		app.Span().Class("closebtn").OnClick(i.dismissError).Body(app.Text("×")),
		app.H4().Class("alert-heading").Text("Oops!"),
		app.Text(i.err.Error()),
	)
}

func (i *index) dismissError(_ app.Context, _ app.Event) {
	i.err = nil
}

type title struct {
	app.Compo
}

func (b *title) Render() app.UI {
	return app.Div().Class("title").Text("CRD Parser")
}

// header is the site header.
type header struct {
	app.Compo

	titleOnClick func(ctx app.Context, _ app.Event)
	hidden       bool
}

type backButton struct {
	app.Compo

	hidden  bool
	onClick func(ctx app.Context, _ app.Event)
}

func (b *backButton) Render() app.UI {
	return app.Ul().Body(
		app.Li().Body(
			app.A().Href("#").Body(
				app.I().Class("fa fa-arrow-left fa-2x")),
		)).Hidden(b.hidden).OnClick(b.onClick)
}

func (h *header) Render() app.UI {
	return app.Header().Body(app.Nav().Body(
		&title{},
		&backButton{onClick: h.titleOnClick, hidden: h.hidden},
		app.Ul().Body(
			app.Li().Body(
				app.A().Href("https://github.com/Skarlso/crd-to-sample-yaml").Target("_blank").Body(
					app.I().Class("fa fa-github fa-2x")),
			)),
	))
}

// textarea is the textarea component that is used to supply the CRD content.
type textarea struct {
	app.Compo
}

func (t *textarea) Render() app.UI {
	return app.Div().Class("input-group mb-3").Body(
		app.Span().Class("input-group-text").Body(app.Text("CRD")),
		app.Textarea().
			Class("form-control").
			ID("crd_data").
			Name("crd_data").
			Placeholder("Place CRD here...").
			AutoFocus(true),
	)
}

// input is the input button.
type input struct {
	app.Compo
}

func (i *input) Render() app.UI {
	return app.Div().Class("input-group mb-3").Body(
		app.Span().Class("input-group-text").Body(app.Text("URL")),
		app.Input().
			Class("url_to_crd").Class("form-control").Placeholder("Paste URL to CRD here...").
			ID("url_to_crd").
			Name("url_to_crd"),
		app.Input().Class("url_username").Class("form-control").Placeholder("Optional username here...").ID("url_username"),
		app.Input().Class("url_password").Class("form-control").Placeholder("Optional password here...").ID("url_password").Type("password"),
		app.Input().Class("url_token").Class("form-control").Placeholder("Optional token here...").ID("url_token").Type("password"),
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
	return app.Div().Body(
		app.Div().Class("row mb-3").Body(
			&textarea{},
			&input{},
			&checkBox{checkHandlerComment: f.checkHandlerComment, checkHandlerMinimal: f.checkHandlerMinimal},
		),
		app.Div().Class("text-end").Body(app.Button().Class("btn btn-primary").Type("submit").Style("margin-top", "15px").Text("Submit").OnClick(f.formHandler)),
	)
}

func (i *index) OnClick(_ app.Context, _ app.Event) {
	ta := app.Window().GetElementByID("crd_data").Get("value")
	if v := ta.String(); v != "" {
		if len(v) > maximumBytes {
			i.err = errors.New("content exceeds maximum length of 200KB")

			return
		}

		i.content = []byte(v)

		return
	}

	inp := app.Window().GetElementByID("url_to_crd").Get("value")
	if inp.String() == "" {
		return
	}

	username := app.Window().GetElementByID("url_username").Get("value")
	password := app.Window().GetElementByID("url_password").Get("value")
	token := app.Window().GetElementByID("url_token").Get("value")

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

	content, err = sanitize.Sanitize(content)
	if err != nil {
		i.err = fmt.Errorf("failed to sanitize content: %w", err)

		return
	}

	i.content = content
}

// checkBox defines if comments should be generated for the sample YAML output.
type checkBox struct {
	app.Compo

	checkHandlerComment app.EventHandler
	checkHandlerMinimal app.EventHandler
}

func (c *checkBox) Render() app.UI {
	return app.Div().Body(
		app.Div().Class("form-check").Body(
			app.Label().Class("form-check-label").For("enable-comments").Body(app.Text("Enable comments on YAML output")),
			app.Input().Class("form-check-input").Type("checkbox").ID("enable-comments").OnClick(c.checkHandlerComment),
		),
		app.Div().Class("form-check").Body(
			app.Label().Class("form-check-label").For("enable-minimal").Body(app.Text("Enable minimal required YAML output")),
			app.Input().Class("form-check-input").Type("checkbox").ID("enable-minimal").OnClick(c.checkHandlerMinimal),
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
	i.content = nil
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
		if err := parser.ParseProperties(version.Name, buf, version.Schema.Properties); err != nil {
			e.content = []byte(err.Error())

			return
		}
		e.content = append(e.content, buf.Bytes()...)
	}
}

func (e *editView) Render() app.UI {
	return app.Div().Body(
		app.H6().Text("Dynamically render CRD content"),
		app.Div().Class("input-group input-group-lg").Body(
			app.Div().Class("container").Body(
				app.Div().Class("row justify-content-around").Body(
					app.Textarea().Class("col form-control").Style("height", "350px").Style("max-height", "800px").Placeholder("Start typing...").ID("input-area").OnInput(e.OnInput),
					// app.Pre().Class("col").Body(
					//	app.Code().Style("height", "250px").Style("max-height", "800px").Class("language-yaml").ID("output-area").Body(
					//		app.Text(string(e.content)),
					//	)),
					app.Textarea().Class("col form-control").Style("height", "350px").Style("max-height", "800px").ID("output-area").Text(string(e.content)),
				),
			),
		))
}

func (i *index) Render() app.UI {
	// Prevent double rendering components.
	if i.isMounted {
		return app.Main().Body(
			app.Script().Src("https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js"),
			app.Script().Src("https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js"),
			app.Div().Class("container").Body(func() app.UI {
				if i.err != nil {
					return app.Div().Class("container").Body(&header{titleOnClick: i.NavBackOnClick, hidden: true}, i.buildError())
				}

				if i.content != nil {
					return &crdView{content: i.content, comment: i.comments, minimal: i.minimal, navigateBackOnClick: i.NavBackOnClick}
				}

				return app.Div().Class("container").Body(
					&header{titleOnClick: i.NavBackOnClick, hidden: true},
					&editView{},
					app.Div().Body(app.H6().Text("Render web view of CRD content")),
					&form{formHandler: i.OnClick, checkHandlerComment: i.OnCheckComment, checkHandlerMinimal: i.OnCheckMinimal})
			}()))
	}

	return app.Main()
}
