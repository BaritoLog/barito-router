package router

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Trader interface {
	TradeSecret(secret string) (profile *Profile, err error)
	TradeName(name string) (profile *Profile, err error)
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

// TradeSecret
func (t *trader) TradeSecret(secret string) (profile *Profile, err error) {

	req, _ := http.NewRequest("GET", t.Url(), nil)
	req.Header.Set("X-App-Secret", secret)

	q := url.Values{}
	q.Add("token", secret)

	req.URL.RawQuery = q.Encode()

	res, err := t.client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		profile, err = NewProfileFromBytes(body)
	}

	return
}

// TradeName
func (t *trader) TradeName(name string) (profile *Profile, err error) {

	req, _ := http.NewRequest("GET", t.Url(), nil)

	q := url.Values{}
	q.Add("cluster_name", name)

	req.URL.RawQuery = q.Encode()

	res, err := t.client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		profile, err = NewProfileFromBytes(body)
	}

	return
}
