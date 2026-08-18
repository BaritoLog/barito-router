package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.com/opentracing/opentracing-go"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/BaritoLog/barito-router/appcontext"
	"github.com/BaritoLog/barito-router/config"
	"github.com/BaritoLog/barito-router/instrumentation"
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

func resetPrometheusMetrics() {
	registry := prometheus.NewRegistry()
	prometheus.DefaultGatherer = registry
	prometheus.DefaultRegisterer = registry

	instrumentation.InitProducerInstrumentation()
}

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

func TestProducerRouter_WithAppSecret_LXC(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		{
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

	router := NewTestSuccessfulProducerLXC(ctrl, marketServer.URL, host, producerPort, consulPort)

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

func TestProducerRouter_WithAppSecret_K8s(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	producerHost, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{},
		ProducerAddress: fmt.Sprintf("%s:%d", producerHost, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8s(ctrl, marketServer.URL, producerHost, producerPort)

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

func TestProducerRouter_WithAppSecret_K8sInvalidConsul(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	producerHost, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{"wrong-consul"},
		ProducerAddress: fmt.Sprintf("%s:%d", producerHost, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8s(ctrl, marketServer.URL, producerHost, producerPort)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusFailedDependency)

	testPayload = sampleRawTimberCollection()
	req, _ = http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp = RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusFailedDependency)
}

func TestProducerRouter_Produce_OnByteIngested(t *testing.T) {
	resetPrometheusMetrics()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	producerHost, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{},
		ProducerAddress: fmt.Sprintf("%s:%d", producerHost, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8s_produce(ctrl, marketServer.URL, producerHost, producerPort)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "")

	expectedByteSize := float64(len(testPayload))

	expected := fmt.Sprintf(`
		# HELP barito_router_produced_total_log_bytes Total log bytes being ingested by the router
		# TYPE barito_router_produced_total_log_bytes counter
		barito_router_produced_total_log_bytes{app_group="", app_name="some-name"} %f
	`, expectedByteSize)

	FatalIfError(t, testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "barito_router_produced_total_log_bytes"))
}

func TestProducerRouter_Produce_batch_OnByteIngested(t *testing.T) {
	resetPrometheusMetrics()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	producerHost, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{},
		ProducerAddress: fmt.Sprintf("%s:%d", producerHost, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8s_produce_batch(ctrl, marketServer.URL, producerHost, producerPort)

	testPayload := sampleRawTimberCollection()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")
	resp1 := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp1, http.StatusOK)
	FatalIfWrongResponseBody(t, resp1, "")

	expectedByteSize := float64(len(testPayload))

	expected := fmt.Sprintf(`
		# HELP barito_router_produced_total_log_bytes Total log bytes being ingested by the router
		# TYPE barito_router_produced_total_log_bytes counter
		barito_router_produced_total_log_bytes{app_group="", app_name="some-name"} %f
	`, expectedByteSize)

	FatalIfError(t, testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "barito_router_produced_total_log_bytes"))

}

func TestProducerRouter_WithAppSecret_DoubleWrite(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		{
			ServiceAddress: host,
			ServicePort:    producerPort,
		},
	})
	defer consulServer.Close()

	host, consulPort := httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{fmt.Sprintf("%s:%d", host, consulPort)},
		ProducerAddress: fmt.Sprintf("%s:%d", host, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerDoubleWrite(ctrl, marketServer.URL, host, producerPort, consulPort)

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
		{
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

func TestProducerRouter_WithAppGroupSecret_LXC(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		{
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

	router := NewTestSuccessfulProducerLXC(ctrl, marketServer.URL, host, producerPort, consulPort)

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

func TestProducerRouter_WithAppGroupSecret_K8s(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	producerHost, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{},
		ProducerAddress: fmt.Sprintf("%s:%d", producerHost, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8s(ctrl, marketServer.URL, producerHost, producerPort)

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

func TestProducerRouter_WithAppGroupSecret_K8s_Forwarding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte("arrived-on-router-b"))
	defer targetServer.Close()

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ProducerLocation: "dc-cluster-b",
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8sWithForwarding(ctrl, marketServer.URL)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")
	config.RouterLocationForwardingMap = map[string]string{
		"dc-cluster-b": targetServer.URL,
	}
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "arrived-on-router-b")

	testPayload = sampleRawTimberCollection()
	req, _ = http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	resp = RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusOK)
	FatalIfWrongResponseBody(t, resp, "arrived-on-router-b")
}
func TestProducerRouter_WithAppGroupSecret_K8s_DoubleForwarding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte("arrived-on-router-b"))
	defer targetServer.Close()

	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ProducerLocation: "dc-cluster-b",
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerK8sWithForwarding(ctrl, marketServer.URL)

	testPayload := sampleRawTimber()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/produce", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Group-Secret", "some-secret")
	req.Header.Add("X-App-Name", "some-name")
	req.Header.Add(RouterForwardingHeaderName, "1")
	config.RouterLocationForwardingMap = map[string]string{
		"dc-cluster-b": targetServer.URL,
	}
	resp := RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusInternalServerError)
	FatalIfWrongResponseBody(t, resp, ErrorDoubleRouterForward)

	testPayload = sampleRawTimberCollection()
	req, _ = http.NewRequest(http.MethodGet, "http://localhost/produce_batch", bytes.NewBuffer(testPayload))
	req.Header.Add("X-App-Secret", "some-secret")
	req.Header.Add(RouterForwardingHeaderName, "1")
	resp = RecordResponse(router.ServeHTTP, req)

	FatalIfWrongResponseStatus(t, resp, http.StatusInternalServerError)
	FatalIfWrongResponseBody(t, resp, ErrorDoubleRouterForward)
}

func TestProducerRouter_WithAppGroupSecret_DoubleWrite(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	targetServer := NewTestServer(http.StatusOK, []byte(""))
	defer targetServer.Close()
	host, producerPort := httpkit.HostOfRawURL(targetServer.URL)

	consulServer := NewJsonTestServer(http.StatusOK, []api.CatalogService{
		{
			ServiceAddress: host,
			ServicePort:    producerPort,
		},
	})
	defer consulServer.Close()

	host, consulPort := httpkit.HostOfRawURL(consulServer.URL)
	marketServer := NewJsonTestServer(http.StatusOK, Profile{
		ConsulHosts:     []string{fmt.Sprintf("%s:%d", host, consulPort)},
		ProducerAddress: fmt.Sprintf("%s:%d", host, producerPort),
	})
	defer marketServer.Close()

	router := NewTestSuccessfulProducerDoubleWrite(ctrl, marketServer.URL, host, producerPort, consulPort)

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

func NewTestSuccessfulProducerLXC(ctrl *gomock.Controller, marketUrl string, host string, producerPort int, consulPort int) ProducerRouter {
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

func NewTestSuccessfulProducerK8s(ctrl *gomock.Controller, marketUrl string, producerHost string, producerPort int) ProducerRouter {
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
		producerAddr: fmt.Sprintf("%s:%d", producerHost, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	return router
}

func NewTestSuccessfulProducerK8sWithForwarding(ctrl *gomock.Controller, marketUrl string) ProducerRouter {
	config := newrelic.NewConfig("barito-router", "")
	config.Enabled = false
	appCtx := appcontext.NewAppContext(config)

	router := &producerRouter{
		addr:                              ":45500",
		marketUrl:                         marketUrl,
		profilePath:                       "profilePath",
		profileByAppGroupPath:             "profileByAppGroupPath",
		client:                            createClient(),
		cacheBag:                          cache.New(1*time.Minute, 10*time.Minute),
		appCtx:                            appCtx,
		producerStore:                     NewProducerStore(),
		isRouterLocationForwardingEnabled: true,
	}

	return router
}

func NewTestSuccessfulProducerK8s_produce(ctrl *gomock.Controller, marketUrl string, producerHost string, producerPort int) ProducerRouter {
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

	pAttr := producerAttributes{
		producerAddr: fmt.Sprintf("%s:%d", producerHost, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	return router
}

func NewTestSuccessfulProducerK8s_produce_batch(ctrl *gomock.Controller, marketUrl string, producerHost string, producerPort int) ProducerRouter {
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
		producerAddr: fmt.Sprintf("%s:%d", producerHost, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	return router
}

func NewTestSuccessfulProducerDoubleWrite(ctrl *gomock.Controller, marketUrl string, host string, producerPort int, consulPort int) ProducerRouter {
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
	pClient.EXPECT().Produce(gomock.Any(), gomock.Any())
	pClient.EXPECT().ProduceBatch(gomock.Any(), gomock.Any())

	pAttr := producerAttributes{
		consulAddr:   fmt.Sprintf("%s:%d", host, consulPort),
		producerAddr: fmt.Sprintf("%s:%d", host, producerPort),
	}

	router.producerStore.producerStoreMap[pAttr] = &grpcParts{
		client: pClient,
	}

	pAttr = producerAttributes{
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

func TestIsK8sLog(t *testing.T) {
	k8sBody := []byte(`{"message":"hello","k8s_metadata":{"pod":"test"}}`)
	if !isK8sLog(k8sBody) {
		t.Fatal("expected k8s log to be detected")
	}

	nonK8sBody := []byte(`{"message":"hello","location":"somewhere"}`)
	if isK8sLog(nonK8sBody) {
		t.Fatal("expected non-k8s log to not be detected")
	}
}

func TestBuildVictoriaLogsBody_Produce(t *testing.T) {
	body := []byte(`{"message":"hello","severity":"INFO"}`)
	result := buildVictoriaLogsBody("/produce", body)
	expected := `{"message":"hello","severity":"INFO"}` + "\n"
	if string(result) != expected {
		t.Fatalf("expected %q, got %q", expected, string(result))
	}
}

func TestBuildVictoriaLogsBody_Produce_K8sFiltered(t *testing.T) {
	body := []byte(`{"message":"hello","k8s_metadata":{"pod":"test"}}`)
	result := buildVictoriaLogsBody("/produce", body)
	if result != nil {
		t.Fatal("expected k8s log to be filtered, got non-nil")
	}
}

func TestBuildVictoriaLogsBody_ProduceBatch(t *testing.T) {
	body := []byte(`{"items":[{"message":"hello"},{"message":"world"}]}`)
	result := buildVictoriaLogsBody("/produce_batch", body)
	expected := "{\"message\":\"hello\"}\n{\"message\":\"world\"}\n"
	if string(result) != expected {
		t.Fatalf("expected %q, got %q", expected, string(result))
	}
}

func TestBuildVictoriaLogsBody_ProduceBatch_FiltersK8sPerItem(t *testing.T) {
	body := []byte(`{"items":[{"message":"hello"},{"message":"k8s","k8s_metadata":{"pod":"test"}},{"message":"world"}]}`)
	result := buildVictoriaLogsBody("/produce_batch", body)
	expected := "{\"message\":\"hello\"}\n{\"message\":\"world\"}\n"
	if string(result) != expected {
		t.Fatalf("expected %q, got %q", expected, string(result))
	}
}

func TestBuildVictoriaLogsBody_ProduceBatch_AllK8s(t *testing.T) {
	body := []byte(`{"items":[{"k8s_metadata":{"pod":"a"}},{"k8s_metadata":{"pod":"b"}}]}`)
	result := buildVictoriaLogsBody("/produce_batch", body)
	if result != nil {
		t.Fatal("expected all-k8s batch to return nil")
	}
}

func TestBuildVictoriaLogsBody_ProduceBatch_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	result := buildVictoriaLogsBody("/produce_batch", body)
	if result != nil {
		t.Fatal("expected invalid JSON batch to return nil")
	}
}

func TestBuildVictoriaLogsBody_ProduceBatch_EmptyItems(t *testing.T) {
	body := []byte(`{"items":[]}`)
	result := buildVictoriaLogsBody("/produce_batch", body)
	if result != nil {
		t.Fatal("expected empty items to return nil")
	}
}

func TestHandleProduceToVictoriaLogs_Success(t *testing.T) {
	resetPrometheusMetrics()

	var receivedBody []byte
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		receivedBody = buf.Bytes()

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}
	body := []byte(`{"message":"hello"}`)

	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "", profile)

	expectedBody := `{"message":"hello"}` + "\n"
	if string(receivedBody) != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, string(receivedBody))
	}
}

func TestHandleProduceToVictoriaLogs_StreamAndExtraFields(t *testing.T) {
	resetPrometheusMetrics()

	var streamFields, extraFields string
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		streamFields = r.Header.Get("VL-Stream-Fields")
		extraFields = r.Header.Get("VL-Extra-Fields")
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{client: createClient(), vlogsClient: newVictoriaLogsClient()}
	profile := &Profile{ClusterName: "test-cluster", Name: "profile-app", AppGroup: "test-group"}
	body := []byte(`{"message":"hello"}`)

	// explicit X-App-Name wins over profile.Name
	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "header-app", profile)
	if streamFields != "cluster_name,application_name,app_group" {
		t.Fatalf("unexpected VL-Stream-Fields: %q", streamFields)
	}
	if extraFields != "cluster_name=test-cluster,application_name=header-app,app_group=test-group" {
		t.Fatalf("unexpected VL-Extra-Fields: %q", extraFields)
	}

	// empty X-App-Name falls back to profile.Name
	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "", profile)
	if extraFields != "cluster_name=test-cluster,application_name=profile-app,app_group=test-group" {
		t.Fatalf("expected fallback to profile.Name, got: %q", extraFields)
	}
}

func TestHandleProduceToVictoriaLogs_Batch_Success(t *testing.T) {
	resetPrometheusMetrics()

	var receivedBody []byte
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		receivedBody = buf.Bytes()
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}
	body := sampleRawTimberCollection()

	router.handleProduceToVictoriaLogs(context.Background(), "/produce_batch", body, "", profile)

	lines := strings.Split(strings.TrimRight(string(receivedBody), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d: %q", len(lines), string(receivedBody))
	}
}

func TestHandleProduceToVictoriaLogs_ServerError(t *testing.T) {
	resetPrometheusMetrics()

	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}
	body := []byte(`{"message": "hello"}`)

	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "", profile)
}

func TestHandleProduceToVictoriaLogs_ConnectionError(t *testing.T) {
	resetPrometheusMetrics()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = "http://localhost:1"
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}
	body := []byte(`{"message": "hello"}`)

	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "", profile)
}

func TestVictoriaLogs_SkipsK8sLogs(t *testing.T) {
	resetPrometheusMetrics()

	var called int32
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}

	k8sPayload := []byte(`{"message":"hello","k8s_metadata":{"pod":"test"}}`)
	router.handleProduceToVictoriaLogs(context.Background(), "/produce", k8sPayload, "", profile)
	if called != 0 {
		t.Fatalf("expected VictoriaLogs to not be called for k8s log, got %d", called)
	}

	nonK8sPayload := []byte(`{"message":"hello","location":"somewhere"}`)
	router.handleProduceToVictoriaLogs(context.Background(), "/produce", nonK8sPayload, "", profile)
	if called != 1 {
		t.Fatalf("expected VictoriaLogs to be called once for non-k8s, got %d", called)
	}
}

// A broken double-write must never crash the router — it runs fire-and-forget
// in its own goroutine, so a panic (e.g. nil client) has to be recovered.
func TestHandleProduceToVictoriaLogs_RecoversFromPanic(t *testing.T) {
	resetPrometheusMetrics()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = "http://localhost:9999"
	defer func() { config.VictoriaLogsUrl = origUrl }()

	// vlogsClient is nil -> Do() dereferences a nil interface -> panic.
	router := &producerRouter{client: createClient()}
	profile := &Profile{ClusterName: "test-cluster"}
	body := []byte(`{"message":"hello"}`)

	// Must return normally instead of propagating the panic.
	router.handleProduceToVictoriaLogs(context.Background(), "/produce", body, "", profile)
}

// The worker pool must drain the queue and forward to VictoriaLogs, including
// decompressing a gzip body off the produce path.
func TestVictoriaLogsWorkerPool_DrainsQueue(t *testing.T) {
	resetPrometheusMetrics()

	received := make(chan []byte, 1)
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		received <- buf.Bytes()
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	// One controlled worker (not the 16-strong pool) so it drains fully and
	// exits before the test returns — no goroutine leaking into the next test.
	router := &producerRouter{vlogsClient: newVictoriaLogsClient(), vlogsJobs: make(chan vlogsJob, 1)}
	workerDone := make(chan struct{})
	go func() { router.vlogsWorker(); close(workerDone) }()

	// gzip the payload so the worker exercises decompression.
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	gz.Write([]byte(`{"message":"hi"}`))
	gz.Close()

	profile := &Profile{ClusterName: "test-cluster"}
	router.enqueueVictoriaLogs(vlogsJob{path: "/produce", body: gzBuf.Bytes(), isGzip: true, profile: profile})

	select {
	case got := <-received:
		if string(got) != `{"message":"hi"}`+"\n" {
			t.Fatalf("expected decompressed body, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker pool did not forward the queued job to VictoriaLogs")
	}

	close(router.vlogsJobs)
	<-workerDone
}

// A full queue must drop the log instead of blocking the produce path.
func TestVictoriaLogsEnqueue_DropsWhenQueueFull(t *testing.T) {
	resetPrometheusMetrics()

	// Unbuffered channel with no workers reading -> every send would block,
	// so enqueue must take the default branch and drop.
	router := &producerRouter{vlogsJobs: make(chan vlogsJob)}
	profile := &Profile{ClusterName: "test-cluster"}

	done := make(chan struct{})
	go func() {
		router.enqueueVictoriaLogs(vlogsJob{path: "/produce", body: []byte(`{}`), profile: profile})
		close(done)
	}()

	select {
	case <-done:
		// returned without blocking -> dropped, as intended
	case <-time.After(1 * time.Second):
		t.Fatal("enqueue blocked on a full queue instead of dropping")
	}
}

func TestVictoriaLogs_ForwardsNonK8sLogs(t *testing.T) {
	resetPrometheusMetrics()

	var called int32
	var receivedContentType string
	vlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer vlServer.Close()

	origUrl := config.VictoriaLogsUrl
	config.VictoriaLogsUrl = vlServer.URL
	defer func() { config.VictoriaLogsUrl = origUrl }()

	router := &producerRouter{
		client:      createClient(),
		vlogsClient: newVictoriaLogsClient(),
	}

	profile := &Profile{ClusterName: "test-cluster"}
	payload := sampleRawTimber()

	router.handleProduceToVictoriaLogs(context.Background(), "/produce", payload, "", profile)

	if called != 1 {
		t.Fatalf("expected VictoriaLogs to be called once, got %d", called)
	}
	if receivedContentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", receivedContentType)
	}
}

func TestVictoriaLogs_NotCalledWhenUrlEmpty(t *testing.T) {
	resetPrometheusMetrics()

	if config.VictoriaLogsUrl != "" {
		t.Fatal("expected VictoriaLogsUrl to be empty by default in tests")
	}
}
