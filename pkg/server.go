package pkg

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
)

// Error contains an HTML error message.
type Error struct {
	Msg string
}

// Version wraps a top level version resource which contains the underlying openAPIV3Schema.
type Version struct {
	Version     string
	Kind        string
	Group       string
	Properties  []*Property
	Description string
	YAML        string
}

// ViewPage is the template for view.html.
type ViewPage struct {
	Versions []Version
}

// Server wraps an address and creates a server around it.
type Server struct {
	address string
}

var (
	//go:embed templates
	files embed.FS
	//go:embed static
	static    embed.FS
	templates map[string]*template.Template
)

// NewServer creates a new address wrapper.
func NewServer(address string) (*Server, error) {
	if err := loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}
	return &Server{
		address: address,
	}, nil
}

// loadTemplates creates a map of loaded templates that are primed and ready to be rendered.
func loadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, "templates")
	if err != nil {
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}
		pt, err := template.ParseFS(files, "templates/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}
	return nil
}

// Run starts a simple http server.
func (s *Server) Run() error {
	// read all files from location and create links for them.
	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/render", s.URLHandler).Methods("GET").Queries("url", "{url}")
	r.HandleFunc("/submit", s.FormHandler).Methods("POST")
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(static)))
	srv := &http.Server{
		Handler:      r,
		Addr:         s.address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

// URLHandler handles sharing urls.
func (s *Server) URLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url, ok := vars["url"]
	if !ok {
		parseError("url was not defined in the parameters", w)
		return
	}

	f := fetcher.NewFetcher(http.DefaultClient)
	content, err := f.Fetch(url)
	if err != nil {
		parseError("failed to fetch url data", w)
		return
	}

	s.renderContent(w, string(content))
}

// IndexHandler handles all requests to / by rendering index.html.
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request on index handler: method: %s; origin: %s; User-Agent: %s; ", r.Method, r.Header.Get("Origin"), r.Header.Get("User-Agent"))
	t := templates["index.html"]
	e := Error{}
	if err := t.Execute(w, e); err != nil {
		fmt.Fprintf(w, "failed to parse index template: %s", err)
	}
}

// FormHandler handles submits through the provided form and renders view.html.
func (s *Server) FormHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request on form handler: method: %s; origin: %s; User-Agent: %s; ", r.Method, r.Header.Get("Origin"), r.Header.Get("User-Agent"))
	if err := r.ParseForm(); err != nil {
		parseError(fmt.Sprintf("value to parse form: %s", err), w)
		return
	}
	value := r.Form["crd_data"]
	if len(value) == 0 || value[0] == "" {
		url := r.Form["url_to_crd"]
		if len(url) == 0 || url[0] == "" {
			parseError("either form or url needs to be provided", w)
			return
		}

		f := fetcher.NewFetcher(http.DefaultClient)
		content, err := f.Fetch(url[0])
		if err != nil {
			parseError("failed to fetch url data", w)
			return
		}
		value = []string{string(content)}
	}
	crdContent := value[0]
	s.renderContent(w, crdContent)
}

func (s *Server) renderContent(w http.ResponseWriter, crdContent string) {
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal([]byte(crdContent), crd); err != nil {
		parseError(fmt.Sprintf("failed to unmarshal into custom resource definition: %s", err), w)
		return
	}
	versions := make([]Version, 0)
	for _, version := range crd.Spec.Versions {
		out, err := parseCRD(version.Schema.OpenAPIV3Schema.Properties, version.Name, version.Schema.OpenAPIV3Schema.Required)
		if err != nil {
			parseError(fmt.Sprintf("failed to parse properties: %s", err), w)
			return
		}
		var buffer []byte
		buf := bytes.NewBuffer(buffer)
		if err := parseProperties(crd.Spec.Group, version.Name, crd.Spec.Names.Kind, version.Schema.OpenAPIV3Schema.Properties, buf, 0, false); err != nil {
			parseError(fmt.Sprintf("failed to generate yaml sample: %s", err), w)
			return
		}
		versions = append(versions, Version{
			Version:     version.Name,
			Properties:  out,
			Kind:        crd.Spec.Names.Kind,
			Group:       crd.Spec.Group,
			Description: version.Schema.OpenAPIV3Schema.Description,
			YAML:        buf.String(),
		})
	}
	view := ViewPage{
		Versions: versions,
	}
	t := templates["view.html"]
	if err := t.Execute(w, view); err != nil {
		parseError(fmt.Sprintf("failed to execute template: %s", err), w)
		return
	}
}

// parseError creates an error on the index.html page.
func parseError(msg string, w http.ResponseWriter) {
	t := templates["index.html"]
	e := Error{
		Msg: msg,
	}
	if err := t.Execute(w, e); err != nil {
		fmt.Fprintf(w, "failed to execute template: %s", err)
	}
}

// Property builds up a Tree structure of embedded things.
type Property struct {
	Name        string
	Description string
	Type        string
	Nullable    bool
	Patterns    string
	Format      string
	Indent      int
	Version     string
	Required    bool
	Properties  []*Property
}

// parseCRD takes the properties and constructs a linked list out of the embedded properties that the recursive
// template can call and construct linked divs.
func parseCRD(properties map[string]v1beta1.JSONSchemaProps, version string, requiredList []string) ([]*Property, error) {
	var (
		sortedKeys []string
		output     []*Property
	)
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		// Create the Property with the values necessary.
		// Check if there are properties for it in Properties or in Array -> Properties.
		// If yes, call parseCRD and add the result to the created properties Properties list.
		// If not, or if we are done, add this new property to the list of properties and return it.
		v := properties[k]
		required := false
		for _, item := range requiredList {
			if item == k {
				required = true
				break
			}
		}
		p := &Property{
			Name:        k,
			Type:        v.Type,
			Description: v.Description,
			Patterns:    v.Pattern,
			Format:      v.Format,
			Nullable:    v.Nullable,
			Version:     version,
			Required:    required,
		}
		if len(properties[k].Properties) > 0 && properties[k].AdditionalProperties == nil {
			requiredList = v.Required
			out, err := parseCRD(properties[k].Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		} else if properties[k].Type == "array" && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0 {
			requiredList = v.Required
			out, err := parseCRD(properties[k].Items.Schema.Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		} else if properties[k].AdditionalProperties != nil {
			requiredList = v.Required
			out, err := parseCRD(properties[k].AdditionalProperties.Schema.Properties, version, requiredList)
			if err != nil {
				return nil, err
			}
			p.Properties = out
		}
		output = append(output, p)
	}
	return output, nil
}
