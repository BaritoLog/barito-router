package testkit

import (
	"net/http"
	"net/http/httptest"
)

func NewTestServer(statusCode int, body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(body)
	}))
}

func RecordResponse(handler func(http.ResponseWriter, *http.Request), req *http.Request) (rr *httptest.ResponseRecorder) {
	rr = httptest.NewRecorder()
	http.HandlerFunc(handler).ServeHTTP(rr, req)

	return
}
