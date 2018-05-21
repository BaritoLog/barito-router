package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	SecretHeaderName = "X-App-Secret"
)

// Router
type Router interface {
	Server() *http.Server
	Address() string
	Trader() Trader
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	KibanaRouter(w http.ResponseWriter, req *http.Request)
}

type router struct {
	addr   string
	trader Trader
	consul ConsulHandler
	server *http.Server
}

// NewRouter
func NewRouter(addr string, trader Trader, consul ConsulHandler) Router {
	r := new(router)
	r.addr = addr
	r.trader = trader
	r.consul = consul
	r.server = &http.Server{Addr: addr}

	return r
}

// Address
func (r *router) Address() string {
	return r.addr
}

func (r *router) Trader() Trader {
	return r.trader
}

// Start
func (r *router) Server() *http.Server {
	return r.server
}

func (r *router) KibanaRouter(w http.ResponseWriter, req *http.Request) {
	clusterName := req.Host
	profile, err := r.Trader().TradeName(clusterName)
	if err != nil {
		r.OnTradeError(w, err)
		return
	}

	if profile == nil {
		r.OnNoProfile(w)
		return
	}

	srv, err := r.consul.Service(profile.Consul, "kibana")
	if err != nil {
		r.OnConsulError(w, err)
		return
	}

	url := &url.URL{
		Scheme: req.URL.Scheme,
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, req)
}

// ServerHTTP
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	secret := req.Header.Get(SecretHeaderName)
	if secret == "" {
		r.OnNoSecret(w)
		return
	}

	profile, err := r.Trader().TradeSecret(secret)
	if err != nil {
		r.OnTradeError(w, err)
		return
	}

	if profile == nil {
		r.OnNoProfile(w)
		return
	}

	srv, err := r.consul.Service(profile.Consul, "barito-receiver")
	if err != nil {
		r.OnConsulError(w, err)
		return
	}

	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, req)
}

func (r *router) OnTradeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(err.Error()))
}

func (r *router) OnNoProfile(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (r *router) OnNoSecret(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func (r *router) OnConsulError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusFailedDependency)
}
