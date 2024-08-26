package router

import (
	"net/http"
	"net/url"
	"strings"
)

type KibanaProxyHandler struct {
	targetRawUrl string
	sourceRawUrl string
}

func (h KibanaProxyHandler) Director(req *http.Request) {
	target, _ := url.Parse(h.targetRawUrl)
	targetQuery := target.RawQuery

	req.URL.Scheme = target.Scheme
	if target.Scheme == "https" {
		req.URL.Host = target.Hostname()
	} else {
		req.URL.Host = target.Host
	}
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

	// If Host is empty, the Request.Write method uses
	// the value of URL.Host.
	// force use URL.Host
	req.Host = req.URL.Host
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}

	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", "")
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func (h KibanaProxyHandler) ModifyResponse(res *http.Response) error {
	location := strings.Replace(res.Header.Get("Location"), h.targetRawUrl, h.sourceRawUrl, -1)
	res.Header.Set("Location", location)

	return nil
}
