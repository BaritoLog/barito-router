package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

const (
	SecretHeaderName = "X-App-Secret"
)

// Router
type Router interface {
	Server() *http.Server
	Address() string
	Trader() Trader
	ProduceHandler(w http.ResponseWriter, req *http.Request)
	KibanaHandler(w http.ResponseWriter, req *http.Request)
}

type router struct {
	addr   string
	trader Trader
	consul ConsulHandler
	server *http.Server
}

// NewProduceRouter
func NewProduceRouter(addr string, trader Trader, consul ConsulHandler) Router {

	r := new(router)
	r.addr = addr
	r.trader = trader
	r.consul = consul

	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/produce", r.ProduceHandler)

	r.server = &http.Server{
		Addr:    addr,
		Handler: muxRouter,
	}

	return r
}

// NewKibanaRouter
func NewKibanaRouter(addr string, trader Trader, consul ConsulHandler) Router {

	r := new(router)
	r.addr = addr
	r.trader = trader
	r.consul = consul

	r.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(r.KibanaHandler),
	}

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

func (r *router) KibanaHandler(w http.ResponseWriter, req *http.Request) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	host := strings.Split(req.Host, ".")
	clusterName := host[0]

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

	sourceUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", req.Host, r.Address()),
	}

	targetUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}

	proxy := NewProxy(sourceUrl.String(), targetUrl.String())
	proxy.ReverseProxy().ServeHTTP(w, req)
}

// ProduceHandler
func (r *router) ProduceHandler(w http.ResponseWriter, req *http.Request) {
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
	w.Write([]byte("No Profile"))
}

func (r *router) OnNoSecret(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("No Secret"))
}

func (r *router) OnConsulError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusFailedDependency)
	w.Write([]byte(err.Error()))
}
