package main

import "github.com/maxence-charriere/go-app/v9/pkg/app"

// index is the main page that contains the textarea and the submit button.
// It will also deal with navigation and user submits.
type index struct {
	app.Compo
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
	return app.Textarea()
}

// input is the input button.
type input struct {
	app.Compo
}

func (i *input) Render() app.UI {
	return app.Input()
}

// form is the form in which the user will submit their input.
type form struct {
	app.Compo
}

func (f *form) Render() app.UI {
	return app.Div().Class("container").Body(
		app.Div().Class("mt-md-20").Body(
			app.Form().Method("POST").Body(
				app.Div().Class("mb-3").Body(
					&textarea{},
					&input{},
				),
				app.Button().Class("btn btn-primary").Type("submit").Style("margin-top", "15px;"),
			),
		),
	)
}

func (i *index) Render() app.UI {
	return app.Main().Body(&header{}, &form{})
}
