package router

import (
	"github.com/cssivision/reverseproxy"
)

type KibanaProxy interface {
	ReverseProxy() *reverseproxy.ReverseProxy
	KibanaProxyHandler() *KibanaProxyHandler
}

type kibanaProxy struct {
	source string
	target string
}

func NewKibanaProxy(source string, target string) KibanaProxy {
	return &kibanaProxy{
		source: source,
		target: target,
	}
}

func (p *kibanaProxy) KibanaProxyHandler() *KibanaProxyHandler {
	return &KibanaProxyHandler{targetRawUrl: p.target, sourceRawUrl: p.source}
}

func (p *kibanaProxy) ReverseProxy() *reverseproxy.ReverseProxy {
	proxy := &reverseproxy.ReverseProxy{
		Director:       p.KibanaProxyHandler().Director,
		ModifyResponse: p.KibanaProxyHandler().ModifyResponse,
	}

	return proxy
}
