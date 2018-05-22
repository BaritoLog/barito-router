package router

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/cssivision/reverseproxy"
)

type ReverseProxyHandler struct {
	targetRawUrl string
	sourceRawUrl string
}

func (h ReverseProxyHandler) Director(req *http.Request) {
	target, _ := url.Parse(h.targetRawUrl)
	targetQuery := target.RawQuery

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
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

func (h ReverseProxyHandler) ModifyResponse(res *http.Response) error {
	location := strings.Replace(res.Header.Get("Location"), h.targetRawUrl, h.sourceRawUrl, -1)
	res.Header.Set("Location", location)

	return nil
}

type Proxy interface {
	ReverseProxy() *reverseproxy.ReverseProxy
	ReverseProxyHandler() *ReverseProxyHandler
}

type proxy struct {
	source string
	target string
}

func NewProxy(source string, target string) Proxy {
	p := new(proxy)
	p.source = source
	p.target = target

	return p
}

func (p *proxy) ReverseProxyHandler() *ReverseProxyHandler {
	return &ReverseProxyHandler{targetRawUrl: p.target, sourceRawUrl: p.source}
}

func (p *proxy) ReverseProxy() *reverseproxy.ReverseProxy {
	proxy := &reverseproxy.ReverseProxy{
		Director:       p.ReverseProxyHandler().Director,
		ModifyResponse: p.ReverseProxyHandler().ModifyResponse,
	}

	return proxy
}
