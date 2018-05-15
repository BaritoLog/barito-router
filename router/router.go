package router

import "net/http"

const (
	SecretHeaderName = "X-App-Secret"
)

// Router
type Router interface {
	Server() *http.Server
	Address() string
	Mapper() *Mapper
}

type router struct {
	addr   string
	mapper *Mapper
	server *http.Server
}

// NewRouter
func NewRouter(addr string) Router {
	r := new(router)
	r.addr = addr
	r.mapper = &Mapper{}
	r.server = &http.Server{Addr: addr, Handler: r}

	return r
}

// Address
func (r *router) Address() string {
	return r.addr
}

func (r *router) Mapper() *Mapper {
	return r.mapper
}

// Start
func (r *router) Server() *http.Server {
	return r.server
}

// ServerHTTP
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	w.Write([]byte("Hello Router"))

	w.WriteHeader(http.StatusOK)
}
