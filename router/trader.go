package router

import (
	"net"
	"net/http"
	"time"
)

type Trader interface {
	Trade(secret string) (profile Profile, err error)
	Url() string
}

// NewHttpTrader
func NewTrader(url string) Trader {
	return &trader{url: url}
}

type trader struct {
	url string
}

// Address
func (t *trader) Url() string {
	return t.url
}

// Trade
func (t *trader) Trade(secret string) (profile Profile, err error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}

	req, _ := http.NewRequest("GET", t.Url(), nil)
	req.Header.Set("X-App-Secret", secret)

	res, err := netClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		profile = NewProfile("some-consul")
	}

	return
}
