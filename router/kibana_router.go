package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	"github.com/BaritoLog/go-boilerplate/httpkit"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
)

const (
	// KeyKibana is meta service name of kibana
	KeyKibana = "kibana"

	// AppKibanaNoProfilePath is path to register when server returned no profile
	AppKibanaNoProfilePath = "api/kibana_no_profile"
)

type KibanaRouter interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	MustBeAuthorizedMiddleware(http.Handler) http.Handler
}

type kibanaRouter struct {
	addr          string
	marketUrl     string
	accessToken   string
	profilePath   string
	authorizePath string
	casAddr       string
	ssoEnabled    bool
	ssoClient     SSOClient

	cacheBag *cache.Cache

	client *http.Client
	appCtx *appcontext.AppContext
}

// NewKibanaRouter is a function for creating new kibana router
func NewKibanaRouter(addr, marketUrl, accessToken, profilePath, authorizePath string, appCtx *appcontext.AppContext) KibanaRouter {
	return &kibanaRouter{
		addr:          addr,
		marketUrl:     marketUrl,
		accessToken:   accessToken,
		profilePath:   profilePath,
		authorizePath: authorizePath,
		cacheBag:      cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:        createClient(),
		appCtx:        appCtx,
	}
}

func NewKibanaRouterWithSSO(addr, marketUrl, accessToken, profilePath, authorizePath string, appCtx *appcontext.AppContext, ssoClient SSOClient) KibanaRouter {
	return &kibanaRouter{
		addr:          addr,
		marketUrl:     marketUrl,
		accessToken:   accessToken,
		profilePath:   profilePath,
		authorizePath: authorizePath,
		cacheBag:      cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:        createClient(),
		appCtx:        appCtx,
		ssoClient:     ssoClient,
		ssoEnabled:    true,
	}
}

func (r *kibanaRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	span := opentracing.StartSpan("barito_router_viewer.view_kibana")
	defer span.Finish()

	params := mux.Vars(req)
	clusterName := params["cluster_name"]

	span.SetTag("app-group", clusterName)
	profile, err := fetchProfileByClusterName(r.client, span.Context(), r.cacheBag, r.marketUrl, r.accessToken, r.profilePath, clusterName)
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

	sourceUrl := fmt.Sprintf("%s://%s:%s", httpkit.SchemeOfRequest(req), req.Host, r.addr)

	var targetUrl string

	targetUrl = fmt.Sprintf("http://%s", profile.KibanaAddress)

	proxy := NewKibanaProxy(sourceUrl, targetUrl)
	proxy.ReverseProxy().ServeHTTP(w, req)
}

func (r *kibanaRouter) MustBeAuthorizedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// check if authorized

		// get required parameters
		params := mux.Vars(req)
		username := strings.Split(req.Context().Value("email").(string), "@")[0]
		clusterName := params["cluster_name"]

		// check to the barito market
		address := fmt.Sprintf("%s/%s", r.marketUrl, r.authorizePath)
		q := url.Values{}
		q.Add("username", username)
		q.Add("cluster_name", clusterName)

		checkReq, _ := http.NewRequest("GET", address, nil)
		checkReq.URL.RawQuery = q.Encode()
		res, err := r.client.Do(checkReq)
		if err != nil {
			onTradeError(w, err)
			return
		}
		if res.StatusCode != http.StatusOK {
			onAuthorizeError(w)
			return
		}

		next.ServeHTTP(w, req)
	})
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

func (r *kibanaRouter) isUseSSO() bool {
	return r.ssoEnabled
}
