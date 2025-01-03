// Why does this exist? You can't use git clone through libraries because
// it will be blocked by CORS. A tiny proxy service will take care of that.
package main

import (
	"io"
	"net/http"
	"time"
)

const (
	headerTimeout = 5 * time.Second
	readTimeout   = 20 * time.Second
	clientTimeout = time.Second * 30
)

type CorsProxy struct{}

func NewCorsProxy() *CorsProxy {
	return &CorsProxy{}
}

func (p *CorsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)

		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)

		return
	}

	// create the request to server
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	// add ALL headers to the connection
	for n, h := range r.Header {
		for _, h := range h {
			req.Header.Add(n, h)
		}
	}

	// create a basic client to send the request
	client := http.Client{
		Timeout: clientTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	for h, v := range resp.Header {
		for _, v := range v {
			w.Header().Add(h, v)
		}
	}
	// copy the response from the server to the connected client request
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))

		return
	}
}

func (p *CorsProxy) Serve() *http.Server {
	return &http.Server{
		Addr:              ":8999",
		Handler:           p,
		ReadHeaderTimeout: headerTimeout,
		ReadTimeout:       readTimeout,
	}
}
