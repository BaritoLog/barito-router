package router

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
)

const (
	KeyKibana   = "kibana"
	KeyProducer = "producer"
)

// Router
type Router interface {
	Server() *http.Server
	Address() string
	Trader() Trader
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

// XtailHandler
func (r *router) XtailHandler(w http.ResponseWriter, req *http.Request) {
	host := strings.Split(req.Host, ".")
	clusterName := host[0]

	profile, err := r.Trader().Trade(clusterName)
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		return
	}

	srv, err := r.consul.Service(profile.ConsulHost, "kafka-pixy")
	if err != nil {
		onConsulError(w, err)
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
			onKafkaPixyError(w, err)
			break
		}

		if message != nil {
			fmt.Fprintf(w, "%s\n", string(message))
			w.(http.Flusher).Flush()
		}
	}

}
