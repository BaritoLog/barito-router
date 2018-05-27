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

func TestNewNewProduceRouter(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewProduceRouter(":8080", trader, consul)

	FatalIf(t, router.Address() != ":8080", "address is return wrong value")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestNewNewKibanaRouter(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewKibanaRouter(":8081", trader, consul)

	FatalIf(t, router.Address() != ":8081", "address is return wrong value")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestNewNewXtailRouter(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewXtailRouter(":8083", trader, consul)

	FatalIf(t, router.Address() != ":8083", "address is return wrong value")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestProduceRouter_TradeError(t *testing.T) {
	want := "some-error"
	r := NewProduceRouter(
		":8080",
		&DummyTrader{err: fmt.Errorf(want)},
		&DummyConsulHandler{},
	)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")
	rr := httptest.NewRecorder()

	r.ProduceHandler(rr, req)

	got := rr.Body.String()

	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)
}

func TestKibanaRouter_TradeError(t *testing.T) {
	want := "some-error"
	r := NewKibanaRouter(
		":8081",
		&DummyTrader{err: fmt.Errorf(want)},
		&DummyConsulHandler{},
	)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	r.KibanaHandler(rr, req)

	got := rr.Body.String()

	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)
}

func TestXtailRouter_TradeError(t *testing.T) {
	want := "some-error"
	r := NewXtailRouter(
		":8083",
		&DummyTrader{err: fmt.Errorf(want)},
		&DummyConsulHandler{},
	)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	r.KibanaHandler(rr, req)

	got := rr.Body.String()

	FatalIfWrongHttpCode(t, rr, http.StatusBadGateway)
	FatalIf(t, got != want, "wrong body result: %s != %s", got, want)
}

func TestProduceRouter_Trade_NoSecret(t *testing.T) {
	r := NewProduceRouter(":8080", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.ProduceHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusBadRequest)
}

func TestKibanaRouter_Trade_InvalidClusterName(t *testing.T) {
	r := NewKibanaRouter(":8081", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.KibanaHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestXtailRouter_Trade_InvalidClusterName(t *testing.T) {
	r := NewXtailRouter(":8083", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.XtailHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestProduceRouter_Trade_ConsulError(t *testing.T) {
	r := NewProduceRouter(":8080",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{err: fmt.Errorf("some-error")})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ProduceHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusFailedDependency)
}

func TestKibanaRouter_Trade_ConsulError(t *testing.T) {
	r := NewProduceRouter(":8081",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{err: fmt.Errorf("some-error")})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.KibanaHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusFailedDependency)
}

func TestXtailRouter_Trade_ConsulError(t *testing.T) {
	r := NewXtailRouter(":8083",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{err: fmt.Errorf("some-error")})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.XtailHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusFailedDependency)
}

func TestProduceRouter_Trade_NoProfile(t *testing.T) {
	r := NewProduceRouter(":8080", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ProduceHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestKibanaRouter_Trade_NoProfile(t *testing.T) {
	r := NewProduceRouter(":8081", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.KibanaHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestXtailRouter_Trade_NoProfile(t *testing.T) {
	r := NewXtailRouter(":8083", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.KibanaHandler, req)
	FatalIfWrongHttpCode(t, rr, http.StatusNotFound)
}

func TestProduceRouter_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte("hello"))
	defer ts.Close()

	serverHost, serverPort := httpkit.Host(ts.URL)

	r := NewProduceRouter(":8080",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{
			catalogService: &api.CatalogService{
				ServiceAddress: serverHost,
				ServicePort:    serverPort,
			},
		})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "abcdefgh")

	rr := HttpRecord(r.ProduceHandler, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)
}

func TestKibanaRouter_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte("hello"))
	defer ts.Close()

	serverHost, serverPort := httpkit.Host(ts.URL)

	r := NewKibanaRouter(":8081",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{
			catalogService: &api.CatalogService{
				ServiceAddress: serverHost,
				ServicePort:    serverPort,
			},
		})

	req, _ := http.NewRequest("GET", "/", nil)

	rr := HttpRecord(r.KibanaHandler, req)

	FatalIfWrongHttpCode(t, rr, http.StatusOK)
}
