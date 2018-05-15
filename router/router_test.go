package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestNew(t *testing.T) {
	router := NewRouter(":8080")

	FatalIf(t, router.Address() != ":8080", "address is return wrong value")
	FatalIf(t, router.Mapper() == nil, "mapper can't be nil")
	FatalIf(t, router.Server() == nil, "server can't be nil")
}

func TestServeHTTP_NoSecret(t *testing.T) {
	r := router{}

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusUnauthorized)
}

func TestServeHTTP_Ok(t *testing.T) {
	r := router{}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)

	got := rr.Body.String()
	want := "Hello Router"
	FatalIf(t, got != want, "wrong result: got %v want %v",
		got, want)

}
