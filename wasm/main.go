package main

import (
	"log"
	"net/http"
	"os"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	args := os.Args
	var static bool
	if len(args) > 1 && args[1] == "--static" {
		static = true
	}
	// The first thing to do is to associate the crdView component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", func() app.Composer {
		return &index{}
	})
	app.Route("/share", func() app.Composer {
		return &crdView{}
	})

	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	handler := &app.Handler{
		Name:    "Preview CRDs",
		Title:   "Preview CRDs",
		Author:  "Gergely Brautigam",
		Version: "v0.6.4",
		HTML:    func() app.HTMLHtml { return app.Html().DataSet("bs-core", "modern").DataSet("bs-theme", "dark") },
		Styles: []string{
			"web/css/alert.css",
			"https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-twilight.min.css",
			"https://cdn.jsdelivr.net/npm/halfmoon@2.0.1/css/halfmoon.min.css",
			"https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css",
		},
		RawHeaders: []string{
			`
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<style>
				header{
					margin: 0px;
					padding: 20px 20px 0px  ;
					border-bottom: 1px solid black;
				}
				nav{
					display: flex;
				}
				.title{
					position: relative;
					top: -5px;
					margin-right: auto;
					font-size: 25px;
				}
				.title:hover{
					color: rgb(0, 0, 164);
					cursor: pointer;
				}
				li{
					width: 50px;
					margin-left: 20px;
					display: inline-block;
					list-style: none;
				}
			</style>`,
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js",
			"https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
		},
		Icon: app.Icon{
			Default: "/web/img/logo.png",
		},
	}
	http.Handle("/", handler)

	if static {
		generateGitHubPages(handler)
		os.Exit(0)
	}

	//nolint: gosec // it's fine
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

func generateGitHubPages(h *app.Handler) {
	if err := app.GenerateStaticWebsite(".", h); err != nil {
		panic(err)
	}
}
