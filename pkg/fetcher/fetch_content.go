package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Fetcher wraps an HTTP client.
type Fetcher struct {
	client   *http.Client
	username string
	password string
	token    string
}

// NewFetcher constructs a new client wrapper with a given client.
func NewFetcher(client *http.Client, username, password, token string) *Fetcher {
	return &Fetcher{
		client:   client,
		username: username,
		password: password,
		token:    token,
	}
}

// Fetch constructs a request and does a client.Do with it.
func (f *Fetcher) Fetch(url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request for url '%s': %w", url, err)
	}

	if f.username != "" && f.password != "" {
		req.SetBasicAuth(f.username, f.password)
	}

	if f.token != "" {
		req.Header.Add("Authorization", "Bearer "+f.token)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			err = fmt.Errorf("failed to close with %w after %w", closeErr, err)
		}
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("failed to fetch url content with status code %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return content, nil
}
