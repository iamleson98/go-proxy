package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true // Enable logging

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
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(`(?i).*\.gif:443$`))).HandleConnect(goproxy.AlwaysReject)

	// Start the proxy server
	log.Println("Starting proxy on :8080")
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
