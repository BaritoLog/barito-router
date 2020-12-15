package router

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	pb "github.com/vwidjaya/barito-proto/producer"
)

const (
	AppSecretHeaderName      = "X-App-Secret"
	AppGroupSecretHeaderName = "X-App-Group-Secret"
	AppNameHeaderName        = "X-App-Name"
	KeyProducer              = "producer"
	AppNoProfilePath         = "api/producer_no_profile"
	AppNoSecretPath          = "api/no_secret"
)

type ProducerRouter interface {
	Server() *http.Server
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type producerRouter struct {
	addr                  string
	marketUrl             string
	profilePath           string
	profileByAppGroupPath string

	cacheBag *cache.Cache

	client *http.Client
	appCtx *appcontext.AppContext

	producerStore *ProducerStore
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	return &producerRouter{
		addr:                  addr,
		marketUrl:             marketUrl,
		profilePath:           profilePath,
		profileByAppGroupPath: profileByAppGroupPath,
		cacheBag:              cache.New(config.CacheExpirationTimeSeconds, 2*config.CacheExpirationTimeSeconds),
		client:                createClient(),
		appCtx:                appCtx,
		producerStore:         NewProducerStore(),
	}
}

func (p *producerRouter) Server() *http.Server {
	return &http.Server{
		Addr:    p.addr,
		Handler: p,
	}
}

func (p *producerRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ping" {
		onPing(w)
		return
	}

	appSecret := req.Header.Get(AppSecretHeaderName)
	appGroupSecret := req.Header.Get(AppGroupSecretHeaderName)
	appName := req.Header.Get(AppNameHeaderName)

	var profile *Profile
	var err error

	span := opentracing.StartSpan("barito_router_producer.produce_log")
	defer span.Finish()

	if appSecret == "" {
		if appGroupSecret != "" && appName != "" {
			profile, err = fetchProfileByAppGroupSecret(p.client, span.Context(), p.cacheBag, p.marketUrl, p.profileByAppGroupPath, appGroupSecret, appName)
			if profile != nil {
				instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profileByAppGroupPath, w, req)
			}
		} else {
			onNoSecret(w)
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoSecretPath, w, req)
			return
		}
	} else {
		profile, err = fetchProfileByAppSecret(p.client, span.Context(), p.cacheBag, p.marketUrl, p.profilePath, appSecret)
		if profile != nil {
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profilePath, w, req)
		}
	}
	if err != nil {
		onTradeError(w, err)
		logProduceError("Failed when fetch Profile", "", appGroupSecret, appName, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		logProduceError("Failed when fetch Profile", "", appGroupSecret, appName, err)
		instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoProfilePath, w, req)

		return
	}

	span.SetTag("app-group", profile.ClusterName)

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, consulAddr, err := consulService(profile.ConsulHosts, srvName, p.cacheBag)
	if err != nil {
		onConsulError(w, err)
		logProduceError("Failed when call consul", profile.ClusterName, appGroupSecret, appName, err)
		return
	}
	if srv == nil {
		logProduceError("Failed when fetch producer service from consul", profile.ClusterName, appGroupSecret, appName, err)
		onConsulError(w, fmt.Errorf("Can't find service from consul: %s", KeyProducer))
		return
	}

	if config.ProducerPort != "" {
		port, err := strconv.Atoi(config.ProducerPort)
		if err == nil {
			srv.ServicePort = port
		}
	}

	pAttr := producerAttributes{
		consulAddr:   consulAddr,
		producerAddr: fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
		producerName: srvName,
		appSecret:    profile.AppSecret,
	}

	producerClient := p.producerStore.GetClient(pAttr)
	ctx := context.Background()
	b, _ := ioutil.ReadAll(req.Body)

	timberContext := TimberContextFromProfile(profile)
	var result *pb.ProduceResult

	if req.URL.Path == "/produce_batch" {
		timberCollection, err := ConvertBytesToTimberCollection(b, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError("Failed when convert message to timberCollection", profile.ClusterName, appGroupSecret, appName, err)
			return
		}

		result, err = producerClient.ProduceBatch(ctx, &timberCollection)
		if err != nil {
			logProduceError("Failed when do RPC to producer on /produce_batch", profile.ClusterName, appGroupSecret, appName, err)
		}
		checkProduceResult(w, result, err)

	} else {
		timber, err := ConvertBytesToTimber(b, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError("Failed when convert message to timber", profile.ClusterName, appGroupSecret, appName, err)
			return
		}

		result, err = producerClient.Produce(ctx, &timber)
		if err != nil {
			logProduceError("Failed when do RPC to producer on /produce", profile.ClusterName, appGroupSecret, appName, err)
		}
		checkProduceResult(w, result, err)
	}
}

func checkProduceResult(w http.ResponseWriter, result *pb.ProduceResult, err error) {
	if err != nil {
		msg := onRpcError(w, err)
		log.Errorf("%s", msg)
		return
	}

	if result != nil {
		onRpcSuccess(w, result.Topic)
	}
}
