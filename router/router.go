package router

import "net/http"

// Router
type Router interface {
	Start()
	Address() string
}

type router struct {
	addr string
}

// NewRouter
func NewRouter(addr string) Router {
	return &router{addr: addr}
}

// Address
func (r *router) Address() string {
	return r.addr
}

// Start
func (r *router) Start() {
	server := &http.Server{
		Addr:    r.addr,
		Handler: r,
	}

	server.ListenAndServe()
}

// ServerHTTP
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello Router"))
	w.WriteHeader(http.StatusOK)
}
