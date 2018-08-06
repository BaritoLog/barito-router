package router

import (
	"net/url"

	"github.com/cssivision/reverseproxy"
)

type ProducerProxy interface {
	ProducerProxyHandler() ProducerProxyHandler
	ReverseProxy() *reverseproxy.ReverseProxy
}

type producerProxy struct {
	target  *url.URL
	profile Profile
}

func NewProducerProxy(target *url.URL, profile Profile) ProducerProxy {
	return &producerProxy{
		target:  target,
		profile: profile,
	}
}

func (p *producerProxy) ProducerProxyHandler() ProducerProxyHandler {
	return producerProxyHandler{
		target:  p.target,
		profile: p.profile,
	}
}

func (p *producerProxy) ReverseProxy() *reverseproxy.ReverseProxy {
	proxy := &reverseproxy.ReverseProxy{
		Director: p.ProducerProxyHandler().Director,
	}

	return proxy
}
