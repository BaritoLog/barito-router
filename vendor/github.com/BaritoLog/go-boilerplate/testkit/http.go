package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func NewTestServer(statusCode int, body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(body)
	}))
}

func NewJsonTestServer(statusCode int, val interface{}) *httptest.Server {
	body, _ := json.Marshal(val)
	return NewTestServer(statusCode, body)
}

func RecordResponse(handler func(http.ResponseWriter, *http.Request), req *http.Request) *http.Response {
	rr := httptest.NewRecorder()
	http.HandlerFunc(handler).ServeHTTP(rr, req)

	return rr.Result()
}
