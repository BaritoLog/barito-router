package router

import (
	"fmt"
	"net/http"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestNewNewXtailRouter(t *testing.T) {
	consul := &DummyConsulHandler{}
	trader := &DummyTrader{url: "http://some-url"}
	FatalIf(t, trader.Url() != "http://some-url", "trader.Url() return wrong")

	router := NewXtailRouter(":8083", trader, consul)

	FatalIf(t, router.Address() != ":8083", "address is return wrong value")
	FatalIf(t, router.Server() == nil, "server can't be nil")
	FatalIf(t, router.Trader() != trader, "trader is wrong")
}

func TestXtailRouter_Trade_InvalidClusterName(t *testing.T) {
	r := NewXtailRouter(":8083", &DummyTrader{}, &DummyConsulHandler{})

	req, _ := http.NewRequest("GET", "/", nil)

	resp := RecordResponse(r.XtailHandler, req)
	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestXtailRouter_Trade_ConsulError(t *testing.T) {
	r := NewXtailRouter(":8083",
		&DummyTrader{profile: &Profile{}},
		&DummyConsulHandler{err: fmt.Errorf("some-error")})

	req, _ := http.NewRequest("GET", "/", nil)

	resp := RecordResponse(r.XtailHandler, req)
	FatalIfWrongResponseStatus(t, resp, http.StatusFailedDependency)
}
