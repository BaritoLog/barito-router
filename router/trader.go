package router

import (
	"net"
	"net/http"
	"time"
)

type Trader interface {
	Trade(secret string) (profile *Profile, err error)
	Url() string
}

// NewHttpTrader
func NewTrader(url string) Trader {
	client := &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 60 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 60 * time.Second,
		},
	}

	return &trader{url: url, client: client}
}

type trader struct {
	url    string
	client *http.Client
}

// Address
func (t *trader) Url() string {
	return t.url
}

// Trade
func (t *trader) Trade(secret string) (profile *Profile, err error) {

	req, _ := http.NewRequest("GET", t.Url(), nil)
	req.Header.Set("X-App-Secret", secret)

	res, err := t.client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {

		profile, err = NewProfileFromBytes([]byte("some-consul"))
	}

	return
}
