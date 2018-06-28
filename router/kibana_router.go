package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BaritoLog/go-boilerplate/httpkit"
	"github.com/hashicorp/consul/api"
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
	if isKibanaPath(clusterName) {
		cookie := getCookie(req)
		if cookie != "" {
			clusterName = cookie
		}
	}

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

	sourceUrl := fmt.Sprintf("%s://%s:%s", httpkit.SchemeOfRequest(req), req.Host, r.addr)
	targetUrl := fmt.Sprintf("%s://%s:%d", getTargetScheme(srv), srv.ServiceAddress, srv.ServicePort)

	urlPath := strings.Split(req.URL.Path, "/")

	for i, v := range urlPath {
		if v == clusterName {
			urlPath = append(urlPath[:i], urlPath[i+1:]...)
			break
		}
	}

	if getCookie(req) != clusterName {
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "clusterName", Value: clusterName, Expires: expiration}
		http.SetCookie(w, &cookie)
	}

	req.URL.Path = strings.Join(urlPath, "/")
	proxy := NewProxy(sourceUrl, targetUrl)
	proxy.ReverseProxy().ServeHTTP(w, req)
}

func (r *kibanaRouter) Server() *http.Server {
	return &http.Server{
		Addr:    r.addr,
		Handler: r,
	}
}

func KibanaGetClustername(req *http.Request) string {
	urlPath := strings.Split(req.URL.Path, "/")
	if len(urlPath) > 1 {
		return urlPath[1]
	}

	return ""
}

func getTargetScheme(srv *api.CatalogService) (scheme string) {
	scheme = srv.NodeMeta["kibana_scheme"]

	if scheme == "" {
		scheme = "http"
	}

	return scheme
}

func getCookie(req *http.Request) string {
	cookie, _ := req.Cookie("clusterName")

	if cookie != nil {
		return cookie.Value
	}

	return ""
}

func isKibanaPath(path string) bool {
	switch path {
	case
		"bundles",
		"api",
		"ui",
		"app",
		"es_admin",
		"elasticsearch",
		"plugins":
		return true
	}
	return false
}
