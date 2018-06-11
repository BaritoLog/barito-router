package router

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
)

const (
	SecretHeaderName = "X-App-Secret"
	KeyKibana        = "kibana"
	KeyProducer      = "producer"
)

// Router
type Router interface {
	Server() *http.Server
	Address() string
	Trader() Trader
	ProduceHandler(w http.ResponseWriter, req *http.Request)
	KibanaHandler(w http.ResponseWriter, req *http.Request)
	XtailHandler(w http.ResponseWriter, req *http.Request)
}

type router struct {
	addr   string
	trader Trader
	consul ConsulHandler
	server *http.Server
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//randSeq
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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

// NewXtailRouter
func NewXtailRouter(addr string, trader Trader, consul ConsulHandler) Router {
	r := new(router)
	r.addr = addr
	r.trader = trader
	r.consul = consul

	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/xtail", r.XtailHandler)

	r.server = &http.Server{
		Addr:    addr,
		Handler: muxRouter,
	}

	http2.ConfigureServer(r.server, &http2.Server{})

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

	srvName, _ := profile.MetaServiceName(KeyKibana)
	srv, err := r.consul.Service(profile.ConsulHost, srvName)
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

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, err := r.consul.Service(profile.ConsulHost, srvName)
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

// XtailHandler
func (r *router) XtailHandler(w http.ResponseWriter, req *http.Request) {
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

	srv, err := r.consul.Service(profile.ConsulHost, "kafka-pixy")
	if err != nil {
		r.OnConsulError(w, err)
		return
	}

	kafkaPixyHost := fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort)
	kafkaTopic := srv.NodeMeta["kafka_topic"]
	if kafkaTopic == "" {
		kafkaTopic = "barito-log"
	}

	rand.Seed(time.Now().UnixNano())

	k := NewKafkaPixy(kafkaPixyHost, kafkaTopic, randSeq(10))

	for {
		message, err := k.Consume()
		if err != nil {
			r.OnKafkaPixyError(w, err)
			break
		}

		if message != nil {
			fmt.Fprintf(w, "%s\n", string(message))
			w.(http.Flusher).Flush()
		}
	}

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

func (r *router) OnKafkaPixyError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusFailedDependency)
	w.Write([]byte(err.Error()))
}
