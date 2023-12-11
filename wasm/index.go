package main

import (
	"errors"
	"net/http"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
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
}

func (i *index) buildError() app.UI {
	return app.Div().Class("alert alert-danger").Role("alert").Body(
		app.Span().Class("closebtn").OnClick(i.dismissError).Body(app.Text("×")),
		app.H4().Class("alert-heading").Text("Oops!"),
		app.Text(i.err.Error()),
	)
}

func (i *index) dismissError(ctx app.Context, e app.Event) {
	i.err = nil
}

// header is the site header.
type header struct {
	app.Compo
}

func (h *header) Render() app.UI {
	return app.Header().Body(app.Nav().Body(
		app.Div().Class("title").Text("CRD Parser"),
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
	return app.Textarea().
		Class("form-control").
		Placeholder("Paste CRD here...").
		Style("height", "200px").
		ID("crd_data").
		Name("crd_data").
		AutoFocus(true)
}

// input is the input button.
type input struct {
	app.Compo
}

func (i *input) Render() app.UI {
	return app.Input().
		Class("url_to_crd").
		ID("url_to_crd").
		Name("url_to_crd").
		Placeholder("Paste URL to CRD here...")
}

// form is the form in which the user will submit their input.
type form struct {
	app.Compo

	formHandler  app.EventHandler
	checkHandler app.EventHandler
}

func (f *form) Render() app.UI {
	return app.Div().Class("mt-md-20").Body(
		app.Div().Body(
			app.Div().Class("mb-3").Body(
				&textarea{},
				&input{},
				&checkBox{checkHandler: f.checkHandler},
			),
			app.Button().Class("btn btn-primary").Type("submit").Style("margin-top", "15px").Text("Submit").OnClick(f.formHandler),
		),
	)
}

func (i *index) OnClick(ctx app.Context, e app.Event) {
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

	f := fetcher.NewFetcher(http.DefaultClient)
	content, err := f.Fetch(inp.String())
	if err != nil {
		i.err = err

		return
	}
	if len(content) > maximumBytes {
		i.err = errors.New("content exceeds maximum length of 200KB")

		return
	}

	i.content = content
}

// checkBox defines if comments should be generated for the sample YAML output.
type checkBox struct {
	app.Compo

	checkHandler app.EventHandler
}

func (c *checkBox) Render() app.UI {
	// https://halfmoonui.pythonanywhere.com/docs/checkbox/ v1.1.1
	return app.P().Body(app.Div().Class("custom-checkbox").Body(
		app.Input().Type("checkbox").ID("enable-comments").OnClick(c.checkHandler),
		app.Label().For("enable-comments").Body(app.Text("enable comments")),
	))

}

func (i *index) OnCheck(ctx app.Context, e app.Event) {
	i.comments = !i.comments
}

func (i *index) OnMount(ctx app.Context) {
	i.isMounted = true
}

func (i *index) Render() app.UI {
	// Prevent double rendering components.
	if i.isMounted {
		return app.Main().Body(app.Div().Class("container").Body(func() app.UI {
			if i.err != nil {
				return app.Div().Class("container").Body(&header{}, i.buildError())
			}

			if i.content != nil {
				return &crdView{content: i.content, comment: i.comments}
			}

			return app.Div().Class("container").Body(&header{}, &form{formHandler: i.OnClick, checkHandler: i.OnCheck})
		}()))
	}

	return app.Main()
}
