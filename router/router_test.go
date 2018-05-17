package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/hashicorp/consul/api"
)

func TestNew(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewRouter(":8080", trader, consul)

	FatalIf(t, router.Address() != ":8080", "address is return wrong value")
	FatalIf(t, router.Mapper() == nil, "mapper can't be nil")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

// func TestServeHTTP_NoSecret(t *testing.T) {
// 	trader := NewTrader("http://some-url")
// 	r := NewRouter(":8080", trader)
//
// 	req, _ := http.NewRequest("GET", "/", nil)
// 	rr := httptest.NewRecorder()
//
// 	handler := http.HandlerFunc(r.ServeHTTP)
// 	handler.ServeHTTP(rr, req)
//
// 	FatalIfWrongHttpCode(t, rr, http.StatusUnauthorized)
// }

// func TestServeHTTP_TradeError(t *testing.T) {
// 	want := "some-error"
//
// 	trader := &DummyTrader{err: fmt.Errorf(want)}
// 	r := NewRouter(":8080", trader)
//
// 	req, _ := http.NewRequest("GET", "/", nil)
// 	req.Header.Add("X-App-Secret", "abcdefgh")
// 	rr := httptest.NewRecorder()
//
// 	handler := http.HandlerFunc(r.ServeHTTP)
// 	handler.ServeHTTP(rr, req)
//
// 	got := rr.Body.String()
//
// 	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
// 	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)
// }

func TestServeHTTP_Trade_Unauthorized(t *testing.T) {
	trader := &DummyTrader{}
	consul := &DummyConsulHandler{}
	r := NewRouter(":8080", trader, consul)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusUnauthorized)
}

func TestServeHTTP_Ok(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	urls, _ := url.Parse(ts.URL)
	urlsDetail := strings.Split(urls.Host, ":")
	serverHost := urlsDetail[0]
	serverPort, _ := strconv.Atoi(urls.Port())

	trader := &DummyTrader{profile: &Profile{}}
	consul := &DummyConsulHandler{
		catalogService: &api.CatalogService{
			ServiceAddress: serverHost,
			ServicePort:    serverPort,
		},
	}
	r := NewRouter(":8080", trader, consul)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(r.ServeHTTP)
	handler.ServeHTTP(rr, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)
}
