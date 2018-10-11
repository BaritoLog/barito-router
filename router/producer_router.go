package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	AppSecretHeaderName      = "X-App-Secret"
	AppGroupSecretHeaderName = "X-App-Group-Secret"
	AppNameHeaderName        = "App-Name"
	KeyProducer              = "producer"
)

type ProducerRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type producerRouter struct {
	addr        string
	marketUrl   string
	profilePath string

	client *http.Client
}

func NewProducerRouter(addr, marketUrl, profilePath string) ProducerRouter {
	return &producerRouter{
		addr:        addr,
		marketUrl:   marketUrl,
		profilePath: profilePath,
		client:      createClient(),
	}
}

func (p *producerRouter) Server() *http.Server {
	return &http.Server{
		Addr:    p.addr,
		Handler: p,
	}
}

func (p *producerRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ping" {
		onPing(w)
		return
	}

	appSecret := req.Header.Get(AppSecretHeaderName)
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)

	var profile *Profile
	var err error

	if appSecret == "" {
		if appGroupSecret != "" && appName != "" {
			profile, err = fetchProfileByAppGroupSecretKey(p.client, p.marketUrl, p.profilePath, appGroupSecret, appName)
		} else {
			onNoSecret(w)
			return
		}
	} else {
		profile, err = fetchProfileByAppSecretKey(p.client, p.marketUrl, p.profilePath, appSecret)
	}
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		return
	}

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, err := consulService(profile.ConsulHost, srvName)
	if err != nil {
		onConsulError(w, err)
		return
	}

	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}

	h := NewProducerProxyHandler(url, *profile, profile.AppSecret)
	proxy := &httputil.ReverseProxy{
		Director: h.Director,
	}
	proxy.ServeHTTP(w, req)
}
