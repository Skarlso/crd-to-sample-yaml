package cmd

import (
	"fmt"
	"io"
	"net/http"
)

// Fetcher wraps an http client.
type Fetcher struct {
	client *http.Client
}

// NewFetcher constructs a new client wrapper with a given client.
func NewFetcher(client *http.Client) *Fetcher {
	return &Fetcher{
		client: client,
	}
}

// Fetch constructs a request and does a client.Do with it.
func (f *Fetcher) Fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request for url '%s': %w", url, err)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close with %s after %w", closeErr, err)
		}
	}()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return content, nil
}
