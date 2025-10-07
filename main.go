package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
)

var whitelist = map[string]bool{
	"10.52.0.121":   true,
	"10.52.198.146": true,
	"10.52.196.141": true,
}

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false // Enable logging

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		var clientIP string

		// Try X-Forwarded-For first
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			clientIP = strings.TrimSpace(strings.Split(xff, ",")[0])
		} else {
			// Fallback to RemoteAddr
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				clientIP = ip
			}
		}

		if !whitelist[clientIP] {
			resp := goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden, "Your IP is not whitelisted.")
			return r, resp
		}

		return r, nil
	})

	// Inject custom header into all requests
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("X-GoProxy", "Powered-by-Goproxy")
		return r, nil
	})

	// Start the proxy server
	log.Println("Starting proxy on :8000")
	log.Fatal(http.ListenAndServe(":8000", proxy))
}
