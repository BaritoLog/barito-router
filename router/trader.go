package router

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Trader interface {
	Trade(s string) (profile *Profile, err error)
	Url() string
}

// NewHttpTrader
func NewTraderBySecret(url string) Trader {
	return &trader{
		url:           url,
		client:        createClient(),
		createRequest: profileRequest,
	}
}

func NewTraderByClusterName(url string) Trader {
	return &trader{
		url:           url,
		client:        createClient(),
		createRequest: profileByClusterNameRequest,
	}
}

func createClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 60 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 60 * time.Second,
		},
	}
}

type trader struct {
	url           string
	client        *http.Client
	createRequest func(url, s string) *http.Request
}

// Address
func (t *trader) Url() string {
	return t.url
}

// TradeSecret
func (t *trader) Trade(s string) (profile *Profile, err error) {

	req := t.createRequest(t.Url(), s)
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

func profileRequest(address, s string) (req *http.Request) {
	q := url.Values{}
	q.Add("token", s)

	req, _ = http.NewRequest("GET", address, nil)
	req.URL.RawQuery = q.Encode()
	return
}

func profileByClusterNameRequest(address, s string) (req *http.Request) {
	q := url.Values{}
	q.Add("cluster_name", s)

	req, _ = http.NewRequest("GET", address, nil)
	req.URL.RawQuery = q.Encode()
	return
}
