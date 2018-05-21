package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/hashicorp/consul/api"
)

func TestNew(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewRouter(":8080", trader, consul)

	FatalIf(t, router.Address() != ":8080", "address is return wrong value")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestServeHTTP_TradeError(t *testing.T) {
	want := "some-error"
	r := NewRouter(
		":8080",
		&DummyTrader{err: fmt.Errorf(want)},
		&DummyConsulHandler{},
	)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	// handler := http.HandlerFunc(r.ServeHTTP)
	r.ServeHTTP(rr, req)

	got := rr.Body.String()

	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)
}

func TestServeHTTP_Trade_NoSecret(t *testing.T) {
	r := NewRouter(":8080", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.ServeHTTP, req)
	FatalIfWrongHttpCode(t, rr, http.StatusBadRequest)
}

func TestServeHTTP_Trade_ConsulError(t *testing.T) {
	r := NewRouter(":8080",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{err: fmt.Errorf("some-error")})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ServeHTTP, req)
	FatalIfWrongHttpCode(t, rr, http.StatusFailedDependency)
}

func TestServeHTTP_Trade_NoProfile(t *testing.T) {
	r := NewRouter(":8080", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ServeHTTP, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestServeHTTP_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte("hello"))
	defer ts.Close()

	serverHost, serverPort := httpkit.Host(ts.URL)

	r := NewRouter(":8080",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{
			catalogService: &api.CatalogService{
				ServiceAddress: serverHost,
				ServicePort:    serverPort,
			},
		})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ServeHTTP, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)
}
