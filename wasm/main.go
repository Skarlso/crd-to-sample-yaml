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
		Version: "v0.8.0",
		HTML: func() app.HTMLHtml {
			return app.Html().DataSet("bs-theme", "auto")
		},
		Styles: []string{
			"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",
			"https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css",
			"https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css",
			"web/css/modern-style.css",
		},
		RawHeaders: []string{
			`
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<meta name="description" content="Generate sample YAML files from Kubernetes CRD definitions with an intuitive web interface">
			<meta name="keywords" content="Kubernetes, CRD, YAML, generator, CustomResourceDefinition">
			<style>
				/* Prevent automatic scrolling behavior */
				* {
					scroll-behavior: auto !important;
				}

				*:focus {
					scroll-behavior: auto !important;
				}

				:root {
					--primary-color: #0d6efd;
					--secondary-color: #6c757d;
					--success-color: #198754;
					--info-color: #0dcaf0;
					--warning-color: #ffc107;
					--danger-color: #dc3545;
					--light-color: #f8f9fa;
					--dark-color: #212529;
					--border-radius: 0.5rem;
					--box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
					--transition: all 0.2s ease-in-out;
				}

				body {
					font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
					line-height: 1.6;
					background: linear-gradient(135deg, #8b9dc3 0%, #9ca3af 100%);
					min-height: 100vh;
					color: #111827;
				}

				.main-container {
					background: rgba(255, 255, 255, 0.98);
					backdrop-filter: blur(10px);
					border-radius: var(--border-radius);
					box-shadow: var(--box-shadow);
					margin: 2rem auto;
					max-width: 1200px;
					overflow: hidden;
					color: #111827;
				}

				@media (prefers-color-scheme: dark) {
					body {
						background: linear-gradient(135deg, #1a202c 0%, #2d3748 100%);
						color: #f7fafc;
					}
					.main-container {
						background: rgba(26, 32, 44, 0.98);
						color: #f7fafc;
					}
				}

				.navbar-brand {
					font-weight: 700;
					font-size: 1.5rem;
					color: #1f2937;
				}

				.btn {
					border-radius: var(--border-radius);
					font-weight: 500;
					transition: var(--transition);
					border: none;
					padding: 0.75rem 1.5rem;
				}

				.btn-primary {
					background: #4b5563;
					box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
				}

				.btn-primary:hover {
					background: #374151;
					box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
				}

				.form-control, .form-select {
					border-radius: var(--border-radius);
					border: 2px solid #e9ecef;
					transition: var(--transition);
					padding: 0.75rem 1rem;
				}

				.form-control:focus, .form-select:focus {
					border-color: #6b7280;
					box-shadow: 0 0 0 0.2rem rgba(107, 114, 128, 0.25);
				}

				.card {
					border: none;
					border-radius: var(--border-radius);
					box-shadow: 0 0.25rem 0.75rem rgba(0, 0, 0, 0.1);
					transition: var(--transition);
				}

				.card:hover {
					box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.12);
				}

				/* Custom accordion styling is now in modern-style.css */

				.loading-spinner {
					display: inline-block;
					width: 20px;
					height: 20px;
					border: 3px solid rgba(255, 255, 255, 0.3);
					border-radius: 50%;
					border-top-color: #fff;
					animation: spin 1s ease-in-out infinite;
				}

				@keyframes spin {
					to { transform: rotate(360deg); }
				}

				.fade-in {
					animation: fadeIn 0.5s ease-in;
				}

				@keyframes fadeIn {
					from { opacity: 0; }
					to { opacity: 1; }
				}
			</style>`,
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",
			"https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
			"https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js",
			"https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js",
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

	if err := http.ListenAndServe(":8000", nil); err != nil { //nolint:gosec // it's fine
		panic(err)
	}
}

func generateGitHubPages(h *app.Handler) {
	if err := app.GenerateStaticWebsite(".", h); err != nil {
		panic(err)
	}
}
