package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BaritoLog/go-boilerplate/httpkit"
)

type KibanaRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type kibanaRouter struct {
	addr        string
	marketUrl   string
	profilePath string

	client *http.Client
}

// NewKibanaRouter
func NewKibanaRouter(addr, marketUrl, profilePath string) KibanaRouter {
	return &kibanaRouter{
		addr:        addr,
		marketUrl:   marketUrl,
		profilePath: profilePath,
		client:      createClient(),
	}
}

func (r *kibanaRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// TODO: validate if no clustername
	clusterName := KibanaGetClustername(req)

	profile, err := fetchProfileByClusterName(r.client, r.marketUrl, r.profilePath, clusterName)
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		return
	}

	srvName, _ := profile.MetaServiceName(KeyKibana)
	srv, err := consulService(profile.ConsulHost, srvName)
	if err != nil {
		onConsulError(w, err)
		return
	}

	scheme := httpkit.SchemeOfRequest(req)

	// TODO: change to fmt.Sprintf
	sourceUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", req.Host, r.addr),
	}

	// TODO: change to fmt.Sprintf
	targetUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}

	proxy := NewProxy(sourceUrl.String(), targetUrl.String())
	proxy.ReverseProxy().ServeHTTP(w, req)
}

func (r *kibanaRouter) Server() *http.Server {
	return &http.Server{
		Addr:    r.addr,
		Handler: r,
	}
}

func KibanaGetClustername(req *http.Request) string {
	host := strings.Split(req.Host, ".")
	return host[0]
}
