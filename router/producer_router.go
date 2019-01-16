package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/instrumentation"
)

const (
	AppSecretHeaderName      = "X-App-Secret"
	AppGroupSecretHeaderName = "X-App-Group-Secret"
	AppNameHeaderName        = "X-App-Name"
	KeyProducer              = "producer"
	AppNoProfilePath		 = "api/no_profile"
	AppNoSecretPath			 = "api/no_secret"
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
}

func NewProducerRouter(addr, marketUrl, profilePath string, profileByAppGroupPath string, appCtx *appcontext.AppContext) ProducerRouter {
	return &producerRouter{
		addr:                  addr,
		marketUrl:             marketUrl,
		profilePath:           profilePath,
		profileByAppGroupPath: profileByAppGroupPath,
		client:                createClient(),
		appCtx: 			   appCtx,
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
				instrumentation.RunTransaction(p.appCtx.NewrelicApp(), p.profileByAppGroupPath, w, req)
			}
		} else {
			onNoSecret(w)
			instrumentation.RunTransaction(p.appCtx.NewrelicApp(), AppNoSecretPath, w, req)
			return
		}
	} else {
		profile, err = fetchProfileByAppSecret(p.client, p.marketUrl, p.profilePath, appSecret)
		if profile != nil {
			instrumentation.RunTransaction(p.appCtx.NewrelicApp(), p.profilePath, w, req)
		}
	}	
	if err != nil {
		onTradeError(w, err)
		return
	}

	if profile == nil {
		onNoProfile(w)
		instrumentation.RunTransaction(p.appCtx.NewrelicApp(), AppNoProfilePath, w, req)

		return
	}

	srvName, _ := profile.MetaServiceName(KeyProducer)
	srv, err := consulService(profile.ConsulHost, srvName)
	if err != nil {
		onConsulError(w, err)
		return
	}

	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", srv.ServiceAddress, srv.ServicePort),
	}

	h := NewProducerProxyHandler(url, *profile, profile.AppSecret)
	proxy := &httputil.ReverseProxy{
		Director:     h.Director,
		ErrorHandler: h.ErrorHandler,
	}
	proxy.ServeHTTP(w, req)
}
