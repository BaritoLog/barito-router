package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestTrader_TradeSecret_Ok(t *testing.T) {
	ts := NewHttpTestServer(http.StatusOK, []byte(`{
			"id": 1,
			"name": "some-name",
			"consul": "some-consul"
		}`))
	defer ts.Close()

	trader := NewTraderBySecret(ts.URL)
	profile, err := trader.Trade("secret")

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

	trader := NewTraderByClusterName(ts.URL)
	profile, err := trader.Trade(name)

	FatalIfError(t, err)
	FatalIf(t, profile == nil, "profile can't be nil")

	profile, err = trader.Trade("wrong-name")
	FatalIfError(t, err)
	FatalIf(t, profile != nil, "profile must be nil")
}

func TestTrader_HttpClientError(t *testing.T) {
	trader := NewTraderBySecret("https://wrong-url")

	_, err := trader.Trade("things")
	FatalIfWrongError(t, err, "no such host")

}
