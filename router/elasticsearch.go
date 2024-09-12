package router

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
)

type ElasticsearchAPI interface {
	ExecuteAPI(w http.ResponseWriter, req *http.Request, clusterName, esEndpoint string)
}

type elasticsearchAPI struct {
	marketUrl   string
	accessToken string
	profilePath string
	cacheBag    *cache.Cache
	client      *http.Client
	appCtx      *appcontext.AppContext
}

func NewElasticsearchAPI(marketUrl, accessToken, profilePath string, appCtx *appcontext.AppContext) ElasticsearchAPI {
	return &elasticsearchAPI{
		marketUrl:   marketUrl,
		accessToken: accessToken,
		profilePath: profilePath,
		cacheBag:    cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:      createClient(),
		appCtx:      appCtx,
	}
}

func (api *elasticsearchAPI) ExecuteAPI(w http.ResponseWriter, req *http.Request, clusterName, esEndpoint string) {
	span := opentracing.StartSpan("barito_router_viewer.view_elasticsearch")
	defer span.Finish()

	span.SetTag("app-group", clusterName)

	profile, err := fetchProfileByClusterName(api.client, span.Context(), api.cacheBag, api.marketUrl, api.accessToken, api.profilePath, clusterName)
	if err != nil {
		onTradeError(w, err)
		return
	}
	if profile == nil {
		onNoProfile(w)
		return
	}

	if req.Method == http.MethodPost || req.Method == http.MethodDelete {
		http.Error(w, "POST and DELETE requests are not allowed", http.StatusMethodNotAllowed)
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
