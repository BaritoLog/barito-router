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
	Mapper() Mapper
	Trader() Trader
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type router struct {
	addr   string
	trader Trader
	consul ConsulHandler
	mapper Mapper
	server *http.Server
}

// NewRouter
func NewRouter(addr string, trader Trader, consul ConsulHandler) Router {
	r := new(router)
	r.addr = addr
	r.mapper = NewMapper()
	r.trader = trader
	r.consul = consul
	r.server = &http.Server{Addr: addr, Handler: r}

	return r
}

// Address
func (r *router) Address() string {
	return r.addr
}

// Mapper
func (r *router) Mapper() Mapper {
	return r.mapper
}

func (r *router) Trader() Trader {
	return r.trader
}

// Start
func (r *router) Server() *http.Server {
	return r.server
}

// ServerHTTP
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	secret := req.Header.Get(SecretHeaderName)
	if secret == "" {
		r.OnUnauthorized(w)
		return
	}

	profile, err := r.Trader().Trade(secret)
	if err != nil {
		r.OnTradeError(w, err)
		return
	}

	if profile == nil {
		r.OnUnauthorized(w)
		return
	}

	srv, err := r.consul.Service(profile.Consul, "barito-receiver")
	if err != nil {
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

func (r *router) OnUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}
