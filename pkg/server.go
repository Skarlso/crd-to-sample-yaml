package pkg

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const htmlPaddingLength = 10

type Version struct {
	Version    string
	Properties []Property
}

type ViewPage struct {
	Versions []Version
}

type Server struct {
	address string
}

var (
	//go:embed templates/*
	files     embed.FS
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
	srv := &http.Server{
		Handler:      r,
		Addr:         s.address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *Server) IndexHandler(w http.ResponseWriter, request *http.Request) {
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
		//properties := make([]Property, 0)
		properties, err := parseCRD(version.Schema.OpenAPIV3Schema.Properties, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to parse properties: %s", err)
			return
		}
		versions = append(versions, Version{
			Version:    version.Name,
			Properties: properties,
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
	w.WriteHeader(http.StatusOK)
	if err := t.Execute(w, view); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to execute template: %s", err)
		return
	}
}

type Property struct {
	Name        string
	Description string
	Type        string
	Nullable    bool
	Patterns    string
	Format      string
	Indent      int
}

func parseCRD(properties map[string]v1beta1.JSONSchemaProps, indent int) ([]Property, error) {
	var (
		sortedKeys []string
		output     []Property
	)
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		if len(properties[k].Properties) == 0 {
			if properties[k].Type == "array" && properties[k].Items.Schema != nil && len(properties[k].Items.Schema.Properties) > 0 {
				out, err := parseCRD(properties[k].Items.Schema.Properties, indent+htmlPaddingLength)
				if err != nil {
					return nil, err
				}
				output = append(output, out...)
			} else {
				v := properties[k]
				t := v.Type
				if t == "array" {
					of := v.Items.Schema.Type
					t = fmt.Sprintf("%s of %ss", t, of)
				}
				output = append(output, Property{
					Name:        k,
					Type:        t,
					Description: v.Description,
					Patterns:    v.Pattern,
					Format:      v.Format,
					Nullable:    v.Nullable,
					Indent:      indent,
				})
			}
		} else if len(properties[k].Properties) > 0 {
			v := properties[k]
			output = append(output, Property{
				Name:        k,
				Type:        v.Type,
				Description: v.Description,
				Patterns:    v.Pattern,
				Format:      v.Format,
				Nullable:    v.Nullable,
				Indent:      indent,
			})
			// recursively parse all sub-properties
			out, err := parseCRD(properties[k].Properties, indent+htmlPaddingLength)
			if err != nil {
				return nil, err
			}
			output = append(output, out...)
		}
	}
	return output, nil
}
