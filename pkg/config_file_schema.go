package pkg

// APIGroups defines groups by which grouping will happen in the resulting HTML output.
type APIGroups struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files,omitempty"`
	Folders     []string `json:"folders,omitempty"`
}

// RenderConfig defines a configuration for the resulting rendered HTML content.
type RenderConfig struct {
	APIGroups []APIGroups `json:"apiGroups"`
}
