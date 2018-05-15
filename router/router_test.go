package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestNew(t *testing.T) {
	router := NewRouter(":8080")
	addr := router.Address()

	FatalIf(t, addr != ":8080", "address is return wrong value")
}

func TestServeHTTP(t *testing.T) {
	r := router{}

	req, _ := http.NewRequest("GET", "/health-check", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)

	got := rr.Body.String()
	want := "Hello Router"
	FatalIf(t, got != want, "wrong result: got %v want %v",
		got, want)

}
