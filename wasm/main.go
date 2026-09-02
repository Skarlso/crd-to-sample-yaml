package main

import (
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
		Name:    "CRD to YAML Generator",
		Title:   "CRD to YAML Generator",
		Author:  "Gergely Brautigam",
		Version: "v1.0.0",
		Styles: []string{
			"/web/css/cty.css",
		},
		Scripts: []string{
			"/web/js/cty.js",
		},
		RawHeaders: []string{
			`
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<meta name="description" content="Generate sample YAML files from Kubernetes CRD definitions with an intuitive web interface">
			<meta name="keywords" content="Kubernetes, CRD, YAML, generator, CustomResourceDefinition">
			<script>
				// Resolved before first paint so the page never flashes the wrong theme.
				(function () {
					var stored = null;
					try { stored = localStorage.getItem('cty-theme'); } catch (e) {}
					var dark = stored ? stored === 'dark'
						: window.matchMedia('(prefers-color-scheme: dark)').matches;
					document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light');
				})();
			</script>`,
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

	err := http.ListenAndServe(":8000", nil) //nolint:gosec // it's fine
	if err != nil {
		panic(err)
	}
}

func generateGitHubPages(h *app.Handler) {
	err := app.GenerateStaticWebsite(".", h)
	if err != nil {
		panic(err)
	}
}
