package router

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, req, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		logProduceError(instrumentation.ErrorFetchProfile, "", appGroupSecret, appName, req, err)
		instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoProfilePath, w, req)

		return
	}

	instrumentation.IncreaseProducerRequestCount(
		profile.ClusterName,
		appName,
	)
	span.SetTag("app-group", profile.ClusterName)

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, consulAddr, err := consulService(profile.ConsulHosts, srvName, profile.ClusterName, p.cacheBag)
	if err != nil {
		onConsulError(w, err)
		logProduceError(instrumentation.ErrorConsulCall, profile.ClusterName, appGroupSecret, appName, req, err)
		return
	}
	if srv == nil {
		logProduceError(instrumentation.ErrorNoProducer, profile.ClusterName, appGroupSecret, appName, req, err)
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
		instrumentation.ObserveTimberCollection(profile.ClusterName, appName, &timberCollection)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, req, err)
			return
		}

		startTime := time.Now()
		result, err = producerClient.ProduceBatch(ctx, &timberCollection)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, time.Since(startTime))

		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, req, err)
		}
		checkProduceResult(w, result, err)

	} else {
		timber, err := ConvertBytesToTimber(b, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			logProduceError(instrumentation.ErrorTimberConvert, profile.ClusterName, appGroupSecret, appName, req, err)
			return
		}

		startTime := time.Now()
		result, err = producerClient.Produce(ctx, &timber)
		instrumentation.ObserveProducerLatency(profile.ClusterName, appName, time.Since(startTime))
		if err != nil {
			logProduceError(instrumentation.ErrorProducerCall, profile.ClusterName, appGroupSecret, appName, req, err)
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
