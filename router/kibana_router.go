package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BaritoLog/go-boilerplate/httpkit"
	"github.com/hashicorp/consul/api"

	cas "github.com/BaritoLog/cas"
)

const (
	KeyKibana = "kibana"
)

type KibanaRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type kibanaRouter struct {
	addr          string
	marketUrl     string
	profilePath   string
	authorizePath string
	casAddr       string

	client *http.Client
}

// NewKibanaRouter
func NewKibanaRouter(addr, marketUrl, profilePath, authorizePath, casAddr string) KibanaRouter {
	return &kibanaRouter{
		addr:          addr,
		marketUrl:     marketUrl,
		profilePath:   profilePath,
		authorizePath: authorizePath,
		casAddr:       casAddr,
		client:        createClient(),
	}
}

func (r *kibanaRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ping" {
		onPing(w)
		return
	}

	if r.isUseCAS() {
		if !cas.IsAuthenticated(req) {
			cas.RedirectToLogin(w, req)
			return
		}

		if req.URL.Path == "/logout" {
			cas.RedirectToLogout(w, req)
			return
		}
	}

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

	if r.isUseCAS() {
		username := cas.Username(req)
		success, err := r.isUserAuthorized(username, clusterName)
		if err != nil {
			return
		}
		if !success {
			onAuthorizeError(w)
			return
		}
	}

	srvName, _ := profile.MetaServiceName(KeyKibana)
	srv, err := consulService(profile.ConsulHost, srvName)
	if err != nil {
		onConsulError(w, err)
		return
	}

	sourceUrl := fmt.Sprintf("%s://%s:%s", httpkit.SchemeOfRequest(req), req.Host, r.addr)
	targetUrl := fmt.Sprintf("%s://%s:%d", getTargetScheme(srv), srv.ServiceAddress, srv.ServicePort)

	proxy := NewKibanaProxy(sourceUrl, targetUrl)
	proxy.ReverseProxy().ServeHTTP(w, req)
}

func (r *kibanaRouter) Server() *http.Server {
	if r.isUseCAS() {
		casURL := r.casAddr
		url, _ := url.Parse(casURL)
		client := cas.NewClient(&cas.Options{
			URL:        url,
			CookiePath: "/",
		})
		return &http.Server{
			Addr:    r.addr,
			Handler: client.Handle(r),
		}
	}

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

func (r *kibanaRouter) isUseCAS() bool {
	return r.casAddr != ""
}

func (r *kibanaRouter) isUserAuthorized(username string, clusterName string) (success bool, err error) {
	address := fmt.Sprintf("%s/%s", r.marketUrl, r.authorizePath)
	q := url.Values{}
	q.Add("username", username)
	q.Add("cluster_name", clusterName)
	success = false

	req, _ := http.NewRequest("GET", address, nil)
	req.URL.RawQuery = q.Encode()

	res, err := r.client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		success = true
	}

	return
}
