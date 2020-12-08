package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/mock"
	"github.com/BaritoLog/go-boilerplate/httpkit"
	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-lib/metrics"
)

func TestProducerRouter_Ping(t *testing.T) {
	marketServer := NewTestServer(http.StatusOK, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/ping", nil)

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath", "profileByAppGroupPath", appCtx)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
}

func TestProducerRouter_FetchError(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":65500", "http://wrong-market", "profilePath", "profileByAppGroupPath", appCtx)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadGateway)
}

func TestProducerRouter_NoSecret(t *testing.T) {
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":65500", "http://wrong-market", "profilePath", "profileByAppGroupPath", appCtx)

	req, _ := http.NewRequest("GET", "/", nil)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusBadRequest)
}

func TestProducerRouter_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath", "profileByAppGroupPath", appCtx)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestProducerRouter_WithAppGroupSecret_NoProfile(t *testing.T) {
	marketServer := NewTestServer(http.StatusNotFound, []byte(``))
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":45500", marketServer.URL, "profileByAppGroupPath", "profileByAppGroupPath", appCtx)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusNotFound)
}

func TestProducerRouter_ConsulError(t *testing.T) {
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts: []string{"wrong-consul"},
	})
	defer marketServer.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-App-Secret", "some-secret")

	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := NewProducerRouter(":45500", marketServer.URL, "profilePath", "profileByAppGroupPath", appCtx)
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusFailedDependency)
}

func TestProducerRouter_WithAppSecret(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		api.CatalogService{
			ServiceAddress: host,
			ServicePort:    producerPort,
		},
	})
	defer consulServer.Close()

	host, consulPort := httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts: []string{fmt.Sprintf("%s:%d", host, consulPort)},
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducer(ctrl, marketServer.URL, host, producerPort, consulPort)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")

	testPayload = sampleRawTimberCollection()
	req, _ = http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp = RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")
}

func TestProducerRouter_WithTrace(t *testing.T) {

	traceHeaders := []string{
		"X-B3-Sampled",
		"X-B3-Spanid",
		"X-B3-Traceid",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		api.CatalogService{
			ServiceAddress: host,
			ServicePort:    producerPort,
		},
	})
	defer consulServer.Close()

	host, consulPort := httpkit.HostOfRawURL(consulServer.URL)
	marketServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		// make sure the trace is there
		for _, name := range traceHeaders {
			FatalIf(t, r.Header.Get(name) == "", fmt.Sprintf("Header %q is not exists", name))
			fmt.Println(r.Header.Get(name))
		}

		p := Profile{
			ConsulHosts: []string{fmt.Sprintf("%s:%d", host, consulPort)},
		}
		body, _ := json.Marshal(p)
		w.Write(body)
	}))
	defer marketServer.Close()

	router := NewTestSuccessfulProducerWithTrace(ctrl, marketServer.URL, host, producerPort, consulPort)

	testPayload := sampleRawTimberCollection()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-app")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")
}

func TestProducerRouter_WithAppGroupSecret(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		api.CatalogService{
			ServiceAddress: host,
			ServicePort:    producerPort,
		},
	})
	defer consulServer.Close()

	host, consulPort := httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts: []string{fmt.Sprintf("%s:%d", host, consulPort)},
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducer(ctrl, marketServer.URL, host, producerPort, consulPort)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")

	testPayload = sampleRawTimberCollection()
	req, _ = http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp = RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")
}

func NewTestSuccessfulProducer(ctrl *gomock.Controller, marketUrl string, host string, producerPort int, consulPort int) ProducerRouter {
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := &producerRouter{
		addr:                  ":45500",
		marketUrl:             marketUrl,
		profilePath:           "profilePath",
		profileByAppGroupPath: "profileByAppGroupPath",
		client:                createClient(),
		cacheBag:              cache.New(1*time.Minute, 10*time.Minute),
		appCtx:                appCtx,
		producerStore:         NewProducerStore(),
	}

	pClient := mock.NewMockProducerClient(ctrl)
	pClient.EXPECT().Produce(gomock.Any(), gomock.Any())
	pClient.EXPECT().ProduceBatch(gomock.Any(), gomock.Any())

	pAttr := producerAttributes{
		consulAddr:   fmt.Sprintf("%s:%d", host, consulPort),
		producerAddr: fmt.Sprintf("%s:%d", host, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	return router
}

func NewTestSuccessfulProducerWithTrace(ctrl *gomock.Controller, marketUrl string, host string, producerPort int, consulPort int) ProducerRouter {
	initTracer()
	config.EnableTracing = true
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := &producerRouter{
		addr:                  ":45500",
		marketUrl:             marketUrl,
		profilePath:           "profilePath",
		profileByAppGroupPath: "profileByAppGroupPath",
		client:                createClient(),
		cacheBag:              cache.New(1*time.Minute, 10*time.Minute),
		appCtx:                appCtx,
		producerStore:         NewProducerStore(),
	}

	pClient := mock.NewMockProducerClient(ctrl)
	pClient.EXPECT().ProduceBatch(gomock.Any(), gomock.Any())

	pAttr := producerAttributes{
		consulAddr:   fmt.Sprintf("%s:%d", host, consulPort),
		producerAddr: fmt.Sprintf("%s:%d", host, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	return router
}

func initTracer() {
	// config from environment variable
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	// Create tracer and then initialize global tracer
	closer, err := cfg.InitGlobalTracer(
		config.JaegerServiceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)

	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()
}
