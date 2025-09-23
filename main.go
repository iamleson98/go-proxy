package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

var whitelist = map[string]bool{
	"10.52.0.121":   true,
	"10.52.198.146": true,
}

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true // Enable logging

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

	// Block access to Reddit during work hours
	proxy.OnRequest(goproxy.DstHostIs("www.reddit.com")).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		hour := time.Now().Hour()
		if hour >= 8 && hour <= 17 {
			resp := goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, "Access to Reddit is blocked during work hours.")
			return req, resp
		}

		return req, nil
	})

	// Reject HTTPS connections to *.gif files
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// May contain multiple IPs, take the first one
			parts := strings.Split(ip, ",")
			first := strings.TrimSpace(parts[0])
			return r, nil
		}

		// Fallback to RemoteAddr
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)

		return r, nil
	})

	// Start the proxy server
	log.Println("Starting proxy on :8000")
	log.Fatal(http.ListenAndServe(":8000", proxy))
}
