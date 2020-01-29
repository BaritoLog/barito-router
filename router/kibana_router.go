package router

import (
	"fmt"
	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	"github.com/BaritoLog/go-boilerplate/httpkit"
	"github.com/hashicorp/consul/api"
	"github.com/patrickmn/go-cache"
	"net/http"
	"net/url"
	"strings"

	"github.com/BaritoLog/cas"
)

const (
	// KeyKibana is meta service name of kibana
	KeyKibana = "kibana"

	// AppKibanaNoProfilePath is path to register when server returned no profile
	AppKibanaNoProfilePath = "api/kibana_no_profile"
)

type KibanaRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type kibanaRouter struct {
	addr          string
	marketUrl     string
	accessToken   string
	profilePath   string
	authorizePath string
	casAddr       string

	cacheBag *cache.Cache

	client *http.Client
	appCtx *appcontext.AppContext
}

// NewKibanaRouter is a function for creating new kibana router
func NewKibanaRouter(addr, marketUrl, accessToken, profilePath, authorizePath, casAddr string, appCtx *appcontext.AppContext) KibanaRouter {
	return &kibanaRouter{
		addr:          addr,
		marketUrl:     marketUrl,
		accessToken:   accessToken,
		profilePath:   profilePath,
		authorizePath: authorizePath,
		casAddr:       casAddr,
		cacheBag:      cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:        createClient(),
		appCtx:        appCtx,
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
	profile, err := fetchProfileByClusterName(r.client, r.cacheBag, r.marketUrl, r.accessToken, r.profilePath, clusterName)
	if profile != nil {
		instrumentation.RunTransaction(r.appCtx.NewRelicApp(), r.profilePath, w, req)
	}
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		instrumentation.RunTransaction(r.appCtx.NewRelicApp(), AppKibanaNoProfilePath, w, req)
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
	srv, _, err := consulService(profile.ConsulHosts, srvName, r.cacheBag)
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

		cookie := &http.Cookie{
			MaxAge:   86400,
			HttpOnly: false,
			Secure:   false,
			Path:     "/",
		}

		client := cas.NewClient(&cas.Options{
			URL:    url,
			Cookie: cookie,
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
