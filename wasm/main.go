package main

import (
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var crdContent = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.9.2
  creationTimestamp: null
  name: krokcommands.delivery.krok.app
spec:
  group: delivery.krok.app
  names:
    kind: KrokCommand
    listKind: KrokCommandList
    plural: krokcommands
    singular: krokcommand
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        description: KrokCommand is the Schema for the krokcommands API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: KrokCommandSpec defines the desired state of KrokCommand
            properties:
              commandHasOutputToWrite:
                description: CommandHasOutputToWrite if defined, it signals the underlying
                  Job, to put its output into a generated and created secret.
                type: boolean
              dependencies:
                description: Dependencies defines a list of command names that this
                  command depends on.
                items:
                  type: string
                type: array
              enabled:
                description: Enabled defines if this command can be executed or not.
                type: boolean
              image:
                description: 'Image defines the image name and tag of the command
                  example: krok-hook/slack-notification:v0.0.1'
                type: string
              platforms:
                description: Platforms holds all the platforms which this command
                  supports.
                items:
                  type: string
                type: array
              readInputFromSecret:
                description: ReadInputFromSecret if defined, the command will take
                  a list of key/value pairs in a secret and apply them as arguments
                  to the command.
                properties:
                  name:
                    type: string
                  namespace:
                    type: string
                required:
                - name
                - namespace
                type: object
              schedule:
                description: 'Schedule of the command. example: 0 * * * * // follows
                  cron job syntax.'
                type: string
            required:
            - image
            type: object
          status:
            description: KrokCommandStatus defines the observed state of KrokCommand
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	// The first thing to do is to associate the hello component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", &hello{})
	//app.Route("/index", &index{})

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
	http.Handle("/", &app.Handler{
		Name:   "Preview CRDs",
		Title:  "Preview CRDs",
		Author: "Gergely Brautigam",
		Styles: []string{
			"web/css/alert.css",
			"web/css/halfmoon-variables.min.css",
			"web/css/main.css",
			"web/css/prism.css",
			"web/css/prism-okaidia.css",
			"web/css/root.css",
		},
		Scripts: []string{
			"web/js/prism.js",
		},
		Icon: app.Icon{
			Default: "/web/img/logo.png",
		},
	})

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
