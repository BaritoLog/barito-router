package httpkit

import (
	"crypto/tls"
	"net/http"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestSchemeOfRequest_https(t *testing.T) {
	req := &http.Request{
		TLS: &tls.ConnectionState{},
	}

	scheme := SchemeOfRequest(req)
	FatalIf(t, scheme != "https", "wrong scheme")
}

func TestSchemeOfRequest_http(t *testing.T) {
	req := &http.Request{}

	scheme := SchemeOfRequest(req)
	FatalIf(t, scheme != "http", "wrong scheme")
}
