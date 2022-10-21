package pkg

import (
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
)

const htmlPaddingLength = 2

type Version struct {
	Version     string
	Kind        string
	Group       string
	Properties  []*Property
	Description string
}

type ViewPage struct {
	Versions []Version
}

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

func NewServer(address string) (*Server, error) {
	if err := loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}
	return &Server{
		address: address,
	}, nil
}

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

func (s *Server) Run() error {
	// read all files from location and create links for them.
	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
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

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("received request on index handler: method: %s; origin: %s; User-Agent: %s; ", r.Method, r.Header.Get("Origin"), r.Header.Get("User-Agent")))
	webSite, err := fs.ReadFile(files, "templates/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read index page: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(webSite))
}

func (s *Server) FormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("received request on form handler: method: %s; origin: %s; User-Agent: %s; ", r.Method, r.Header.Get("Origin"), r.Header.Get("User-Agent")))
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "value to parse form: %s", err)
		return
	}
	value := r.Form["crd_data"]

	if len(value) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "form value is empty")
		return
	}
	crdContent := value[0]
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal([]byte(crdContent), crd); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to unmarshal into custom resource definition: %s", err)
		return
	}
	versions := make([]Version, 0)
	for _, version := range crd.Spec.Versions {
		out, err := parseCRD(version.Schema.OpenAPIV3Schema.Properties, version.Name, version.Schema.OpenAPIV3Schema.Required)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to parse properties: %s", err)
			return
		}
		versions = append(versions, Version{
			Version:     version.Name,
			Properties:  out,
			Kind:        crd.Spec.Names.Kind,
			Group:       crd.Spec.Group,
			Description: version.Schema.OpenAPIV3Schema.Description,
		})
	}
	view := ViewPage{
		Versions: versions,
	}
	t, err := template.ParseFS(files, "templates/view.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to load view page: %s", err)
		return
	}
	if err := t.Execute(w, view); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to execute template: %s", err)
		return
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
		if len(properties[k].Properties) > 0 {
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
		}
		output = append(output, p)
	}
	return output, nil
}
