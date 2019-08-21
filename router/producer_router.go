package router

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/instrumentation"
	log "github.com/sirupsen/logrus"
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

	client *http.Client
	appCtx *appcontext.AppContext

	producerStore ProducerStore
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	return &producerRouter{
		addr:                  addr,
		marketUrl:             marketUrl,
		profilePath:           profilePath,
		profileByAppGroupPath: profileByAppGroupPath,
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

	if appSecret == "" {
		if appGroupSecret != "" && appName != "" {
			profile, err = fetchProfileByAppGroupSecret(p.client, p.marketUrl, p.profileByAppGroupPath, appGroupSecret, appName)
			if profile != nil {
				instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profileByAppGroupPath, w, req)
			}
		} else {
			onNoSecret(w)
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoSecretPath, w, req)
			return
		}
	} else {
		profile, err = fetchProfileByAppSecret(p.client, p.marketUrl, p.profilePath, appSecret)
		if profile != nil {
			instrumentation.RunTransaction(p.appCtx.NewRelicApp(), p.profilePath, w, req)
		}
	}
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		instrumentation.RunTransaction(p.appCtx.NewRelicApp(), AppNoProfilePath, w, req)

		return
	}

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, err := consulService(profile.ConsulHost, srvName)
	if err != nil {
		onConsulError(w, err)
		return
	}

	pAttr := producerAttributes{
		consulAddr:   profile.ConsulHost,
		producerAddr: fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
		producerName: srvName,
		appSecret:    profile.AppSecret,
	}

	producerClient := p.producerStore.GetClient(pAttr)
	ctx := context.Background()
	b, _ := ioutil.ReadAll(req.Body)

	switch req.URL.Path {
	case "/produce":
		timberContext := TimberContextFromProfile(profile)
		timber, err := ConvertBytesToTimber(b, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			return
		}

		result, err := producerClient.Produce(ctx, &timber)
		if err != nil {
			msg := onRpcError(w, err)
			log.Errorf("%s", msg)
			return
		}

		if result != nil {
			onRpcSuccess(w, result.Topic)
		}

	case "/produce_batch":
		timberContext := TimberContextFromProfile(profile)
		timberCollection, err := ConvertBytesToTimberCollection(b, timberContext)
		if err != nil {
			log.Errorf("%s", err.Error())
			return
		}

		result, err := producerClient.ProduceBatch(ctx, &timberCollection)
		if err != nil {
			msg := onRpcError(w, err)
			log.Errorf("%s", msg)
			return
		}

		if result != nil {
			onRpcSuccess(w, result.Topic)
		}
	}
}
