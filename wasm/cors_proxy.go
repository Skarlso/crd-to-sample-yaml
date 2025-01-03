// Why does this exist? You can't use git clone through libraries because
// it will be blocked by CORS. A tiny proxy service will take care of that.
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	headerTimeout = 5 * time.Second
	readTimeout   = 20 * time.Second
)

func createReverseProxyServer(baseURL string) *http.Server {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetURL := req.URL.Query().Get("url")
			if targetURL == "" {
				return
			}

			target, err := url.Parse(targetURL)
			if err != nil {
				return
			}

			req.URL = target
			req.Host = target.Host
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", baseURL)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              ":8999",
		Handler:           handler,
		ReadHeaderTimeout: headerTimeout,
		ReadTimeout:       readTimeout,
	}

	return server
}
