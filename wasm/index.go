package main

import (
	"net/http"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// index is the main page that contains the textarea and the submit button.
// It will also deal with navigation and user submits.
type index struct {
	app.Compo

	content   []byte
	isMounted bool
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

	formHandler app.EventHandler
}

func (f *form) Render() app.UI {
	return app.Div().Class("mt-md-20").Body(
		app.Div().Body(
			app.Div().Class("mb-3").Body(
				&textarea{},
				&input{},
			),
			app.Button().Class("btn btn-primary").Type("submit").Style("margin-top", "15px").Text("Submit").OnClick(f.formHandler),
		),
	)
}

func (i *index) OnClick(ctx app.Context, e app.Event) {
	ta := app.Window().GetElementByID("crd_data").Get("value")
	if v := ta.String(); v != "" {
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
		app.Log("failed to fetch url: ", err)
		return
	}

	i.content = content
}

func (i *index) OnMount(ctx app.Context) {
	i.isMounted = true
}

func (i *index) Render() app.UI {
	// Prevent double rendering components.
	if i.isMounted {
		return app.Main().Body(app.Div().Class("container").Body(func() app.UI {
			if i.content != nil {
				return &crdView{content: i.content}
			}

			return app.Div().Class("container").Body(&header{}, &form{formHandler: i.OnClick})
		}()))
	}

	return app.Main()
}
