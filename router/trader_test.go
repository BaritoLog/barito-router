package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestTrader_TradeName(t *testing.T) {
	name := "foobar"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qName := r.URL.Query().Get("cluster_name")

		if name != qName {
			w.WriteHeader(http.StatusNotFound)
		} else {
			fmt.Fprintln(w, `{
					"id": 1,
					"name": "some-name",
					"consul": "some-consul"
				}`)

		}

	}))
	defer ts.Close()

	trader := NewTrader(ts.URL)
	profile, err := trader.TradeName(name)

	FatalIfError(t, err)
	FatalIf(t, profile == nil, "profile can't be nil")

	profile, err = trader.TradeName("wrong-name")
	FatalIfError(t, err)
	FatalIf(t, profile != nil, "profile must be nil")
}

func TestTrader_HttpClientError(t *testing.T) {
	trader := NewTrader("https://wrong-url")

	_, err := trader.TradeSecret("secret")
	FatalIfWrongError(t, err, "dial tcp: lookup wrong-url: no such host")

	_, err = trader.TradeName("secret")
	FatalIfWrongError(t, err, "dial tcp: lookup wrong-url: no such host")
}
