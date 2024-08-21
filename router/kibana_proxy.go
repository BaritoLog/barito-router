package router

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"

	"github.com/cssivision/reverseproxy"
)

type KibanaProxy interface {
	ReverseProxy() *reverseproxy.ReverseProxy
	KibanaProxyHandler() *KibanaProxyHandler
}

type kibanaProxy struct {
	source      string
	target      string
	mtlsEnabled bool
}

func NewKibanaProxy(source, target string, mtlsEnabled bool) KibanaProxy {
	return &kibanaProxy{
		source:      source,
		target:      target,
		mtlsEnabled: mtlsEnabled,
	}
}

func (p *kibanaProxy) KibanaProxyHandler() *KibanaProxyHandler {
	return &KibanaProxyHandler{targetRawUrl: p.target, sourceRawUrl: p.source}
}

func (p *kibanaProxy) getTransport() (transport *http.Transport) {
	transport = &http.Transport{}
	if !p.mtlsEnabled {
		return
	}

	mtlsCertPathPrefix := os.Getenv("MTLS_CERTS_PATH_PREFIX")
	caCert, err := os.ReadFile(mtlsCertPathPrefix + "/ca.crt")
	if err != nil {
		return
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return
	}

	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair(mtlsCertPathPrefix+"/tls.crt", mtlsCertPathPrefix+"/tls.key")
	if err != nil {
		return
	}

	parsedTargetURL, err := url.Parse(p.target)
	if err != nil {
		return
	}
	tlsConfig := &tls.Config{
		ServerName:         parsedTargetURL.Hostname(),
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}

	transport.TLSClientConfig = tlsConfig
	return
}

func (p *kibanaProxy) ReverseProxy() *reverseproxy.ReverseProxy {
	proxy := &reverseproxy.ReverseProxy{
		Director:       p.KibanaProxyHandler().Director,
		ModifyResponse: p.KibanaProxyHandler().ModifyResponse,
		Transport:      p.getTransport(),
	}

	return proxy
}
