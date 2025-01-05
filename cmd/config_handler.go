package cmd

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
)

type ConfigHandler struct {
	configFileLocation string
}

func (h *ConfigHandler) CRDs() ([]*pkg.SchemaType, error) {
	if _, err := os.Stat(h.configFileLocation); os.IsNotExist(err) {
		return nil, fmt.Errorf("file under '%s' does not exist", h.configFileLocation)
	}
	content, err := os.ReadFile(h.configFileLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	configFile := &RenderConfig{}
	if err = yaml.Unmarshal(content, configFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	// for each api group, call file handler and folder handler and gather all
	// the CRDs.
	var result []*pkg.SchemaType

	for _, group := range configFile.APIGroups {
		for _, file := range group.Files {
			handler := FileHandler{location: file, group: group.Name}
			fileResults, err := handler.CRDs()
			if err != nil {
				return nil, fmt.Errorf("failed to process CRDs for files in groups %s: %w", group.Name, err)
			}

			result = append(result, fileResults...)
		}

		for _, folder := range group.Folders {
			handler := FolderHandler{location: folder, group: group.Name}
			folderResults, err := handler.CRDs()
			if err != nil {
				return nil, fmt.Errorf("failed to process CRDs for folders %s: %w", handler.location, err)
			}

			result = append(result, folderResults...)
		}

		for _, url := range group.URLs {
			handler := URLHandler{
				url:      url.URL,
				username: url.Username,
				password: url.Password,
				token:    url.Token,
				group:    group.Name,
			}
			crds, err := handler.CRDs()
			if err != nil {
				return nil, fmt.Errorf("failed to process CRDs for url %s: %w", handler.url, err)
			}

			result = append(result, crds...)
		}

		for _, url := range group.GitURLs {
			handler := GitHandler{
				URL:         url.URL,
				Username:    url.Username,
				Password:    url.Password,
				Token:       url.Token,
				Tag:         url.Tag,
				privSSHKey:  url.PrivateKey,
				useSSHAgent: url.UseSSHAgent,
				group:       group.Name,
			}
			crds, err := handler.CRDs()
			if err != nil {
				return nil, fmt.Errorf("failed to process CRDs for git url %s: %w", handler.URL, err)
			}

			result = append(result, crds...)
		}
	}

	return result, nil
}
