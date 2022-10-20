package pkg

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type ViewPage struct {
	CRD string
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
	view := &ViewPage{
		CRD: crdContent,
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
