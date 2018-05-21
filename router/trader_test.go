package router

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestTrader_New(t *testing.T) {
	want := "http://host:port"
	trader := NewTrader(want)

	got := trader.Url()
	FatalIf(t, got != want, "%s != %s", got, want)
}

func TestTrader_TradeSecret_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte(`{
			"id": 1,
			"name": "some-name",
			"consul": "some-consul"
		}`))
	defer ts.Close()

	trader := NewTrader(ts.URL)
	profile, err := trader.TradeSecret("secret")

	FatalIfError(t, err)
	FatalIf(t, profile == nil, "profile can't be nil")
}

func TestTrader_TradeName_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte(`{
			"id": 1,
			"name": "some-name",
			"consul": "some-consul"
		}`))
	defer ts.Close()

	trader := NewTrader(ts.URL)
	profile, err := trader.TradeName("foobar")

	FatalIfError(t, err)
	FatalIf(t, profile == nil, "profile can't be nil")
}

func TestTrader_Trade_HttpClientError(t *testing.T) {
	trader := NewTrader("https://wrong-url")

	_, err := trader.TradeSecret("secret")
	FatalIf(t, !strings.Contains(err.Error(), "no such host"), "wrong error")
}
