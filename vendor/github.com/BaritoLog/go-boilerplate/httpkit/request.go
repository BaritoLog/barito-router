package httpkit

import "net/http"

func SchemeOfRequest(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}

	return "http"
}
