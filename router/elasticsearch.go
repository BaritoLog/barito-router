package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/time/rate"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
)

type ElasticsearchAPI interface {
	Elasticsearch(w http.ResponseWriter, req *http.Request, clusterName, esEndpoint string)
}

type elasticsearchAPI struct {
	marketUrl   string
	accessToken string
	profilePath string
	cacheBag    *cache.Cache
	client      *http.Client
	appCtx      *appcontext.AppContext
	limiter     *rate.Limiter
}

func NewElasticsearchAPI(marketUrl, accessToken, profilePath string, appCtx *appcontext.AppContext) ElasticsearchAPI {
	return &elasticsearchAPI{
		marketUrl:   marketUrl,
		accessToken: accessToken,
		profilePath: profilePath,
		cacheBag:    cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:      createClient(),
		appCtx:      appCtx,
		limiter:     rate.NewLimiter(1, 5),
	}
}

func (api *elasticsearchAPI) Elasticsearch(w http.ResponseWriter, req *http.Request, clusterName, esEndpoint string) {
	span := opentracing.StartSpan("barito_router_viewer.elasticsearch")
	defer span.Finish()

	span.SetTag("app-group", clusterName)

	// Extract App-Secret from the request header
	appSecret := req.Header.Get("App-Secret")
	if appSecret == "" {
		http.Error(w, "App-Secret header is required", http.StatusUnauthorized)
		return
	}

	// Validate the appSecret against Barito Market using fetchProfileByClusterName
	profile, err := fetchProfileByClusterName(api.client, span.Context(), api.cacheBag, api.marketUrl, api.profilePath, clusterName, api.accessToken)
	if err != nil || profile == nil || profile.AppGroupSecret != appSecret {
		http.Error(w, "Invalid app secret or cluster name", http.StatusUnauthorized)
		return
	}

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(w, "Only GET and POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	allowedEndpoints := []string{
		"_search",
		"_search/scroll",
		"_doc",
		"_cat/indices",
		"_eql/search",
		"_mget",
		"_index",
		"_ingest/pipeline",
	}

	isAllowed := false
	for _, endpoint := range allowedEndpoints {
		if strings.HasPrefix(esEndpoint, endpoint) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		http.Error(w, "This endpoint is not allowed", http.StatusForbidden)
		return
	}

	if !api.limiter.Allow() {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	targetUrl := fmt.Sprintf("http://%s/%s", profile.ElasticsearchAddress, esEndpoint)

	esReq, err := http.NewRequest(req.Method, targetUrl, req.Body)
	if err != nil {
		onTradeError(w, err)
		return
	}

	for name, values := range req.Header {
		for _, value := range values {
			esReq.Header.Add(name, value)
		}
	}

	esReq.Header.Add("App-Secret", appSecret)

	esRes, err := api.client.Do(esReq)
	if err != nil {
		onTradeError(w, err)
		return
	}
	defer esRes.Body.Close()

	body, err := ioutil.ReadAll(esRes.Body)
	if err != nil {
		onTradeError(w, err)
		return
	}

	w.WriteHeader(esRes.StatusCode)
	w.Write(body)
}
