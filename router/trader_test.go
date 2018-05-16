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

func TestTrader_Trade_Ok(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello test")
	}))
	defer ts.Close()

	trader := NewTrader(ts.URL)

	item, err := trader.Trade("secret")
	FatalIfError(t, err)
	FatalIf(t, item == nil, "item can't be nil")
}

func TestTrader_Trade_HttpClientError(t *testing.T) {
	trader := NewTrader("https://wrong-url")

	_, err := trader.Trade("secret")
	FatalIfWrongError(t, err, "Get https://wrong-url: dial tcp: lookup wrong-url: no such host")
}
