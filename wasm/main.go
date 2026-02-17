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
		HTML: func() app.HTMLHtml {
			return app.Html().DataSet("bs-theme", "auto")
		},
		Styles: []string{
			"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
			"https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css",
			"web/css/modern-style.css",
		},
		RawHeaders: []string{
			`
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<meta name="description" content="Generate sample YAML files from Kubernetes CRD definitions with an intuitive web interface">
			<meta name="keywords" content="Kubernetes, CRD, YAML, generator, CustomResourceDefinition">
			<style>
				body {
					font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
					line-height: 1.6;
					background: #f9fafb;
					min-height: 100vh;
					color: #1f2937;
				}

				.main-container {
					background: #fff;
					border: 1px solid #e5e7eb;
					border-radius: 0.5rem;
					margin: 2rem auto;
					max-width: 1200px;
					overflow: hidden;
					color: #1f2937;
				}

				@media (prefers-color-scheme: dark) {
					body {
						background: #1e1f22;
						color: #f2f3f5;
					}
					.main-container {
						background: #2b2d31;
						border-color: #3f4147;
						color: #f2f3f5;
					}
				}

				.btn {
					border-radius: 0.375rem;
					font-weight: 500;
					transition: background 0.15s ease;
					border: none;
					padding: 0.625rem 1.25rem;
				}

				.btn-primary {
					background: #2563eb;
					color: #fff;
				}

				.btn-primary:hover {
					background: #1d4ed8;
				}

				@media (prefers-color-scheme: dark) {
					.btn-primary {
						background: #5865f2;
						color: #fff;
					}
					.btn-primary:hover {
						background: #7983f5;
					}
					.form-control, .form-select {
						background: #383a40;
						border-color: #3f4147;
						color: #f2f3f5;
					}
					.card {
						background: #2b2d31;
						border-color: #3f4147;
					}
				}

				.form-control, .form-select {
					border-radius: 0.375rem;
					border: 1px solid #e5e7eb;
					transition: border-color 0.15s ease;
					padding: 0.625rem 0.75rem;
				}

				.form-control:focus, .form-select:focus {
					border-color: #2563eb;
					box-shadow: none;
				}

				.card {
					border: 1px solid #e5e7eb;
					border-radius: 0.375rem;
				}
			</style>`,
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",
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
