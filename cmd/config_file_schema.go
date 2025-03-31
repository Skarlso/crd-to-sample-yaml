package cmd

// URLs contains url configuration.
type URLs struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

// GITUrls contains git url configuration.
type GITUrls struct {
	URL         string `json:"url"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Token       string `json:"token,omitempty"`
	Tag         string `json:"tag,omitempty"`
	PrivateKey  string `json:"privateKey,omitempty"`
	UseSSHAgent bool   `json:"useSSHAgent,omitempty"`
}

// APIGroups defines groups by which grouping will happen in the resulting HTML output.
type APIGroups struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Files       []string  `json:"files,omitempty"`
	Folders     []string  `json:"folders,omitempty"`
	URLs        []URLs    `json:"urls,omitempty"`
	GitURLs     []GITUrls `json:"gitUrls,omitempty"`
}

// RenderConfig defines a configuration for the resulting rendered HTML content.
type RenderConfig struct {
	APIGroups []APIGroups `json:"apiGroups"`
}
