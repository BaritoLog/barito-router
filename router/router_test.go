package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestNew(t *testing.T) {
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewRouter(":8080", trader)

	FatalIf(t, router.Address() != ":8080", "address is return wrong value")
	FatalIf(t, router.Mapper() == nil, "mapper can't be nil")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestServeHTTP_NoSecret(t *testing.T) {
	trader := NewTrader("http://some-url")
	r := NewRouter(":8080", trader)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusUnauthorized)
}

func TestServeHTTP_TradeError(t *testing.T) {
	want := "some-error"

	trader := &DummyTrader{err: fmt.Errorf(want)}
	r := NewRouter(":8080", trader)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	got := rr.Body.String()

	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)

}

func TestServeHTTP_Trade_Unauthorized(t *testing.T) {
	trader := &DummyTrader{}
	r := NewRouter(":8080", trader)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusUnauthorized)
}

func TestServeHTTP_Ok(t *testing.T) {

	trader := &DummyTrader{profile: &Profile{}}
	r := NewRouter(":8080", trader)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)
}
