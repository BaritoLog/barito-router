package instrumentation

import (
	"math"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var disableAppNameLabelMetrics bool = false

var forwardToOtherRouterSuccess *prometheus.CounterVec
var forwardToOtherRouterFailed *prometheus.CounterVec
var doubleRouterForward *prometheus.CounterVec
var producerRequestCount *prometheus.CounterVec
var producerRequestError *prometheus.CounterVec
var producerNumberMessagePerBatch *prometheus.SummaryVec
var producerLengthPerMessage *prometheus.SummaryVec
var latencyToMarket prometheus.Summary
var latencyToConsul *prometheus.SummaryVec
var latencyToProducer *prometheus.SummaryVec
var totalLogBytesIngested *prometheus.CounterVec

// list error message for producerRequestError metrics
const (
	ErrorFetchProfile      = "fetch_profile"
	ErrorConsulCall        = "consul_call"
	ErrorNoProducer        = "no_producer"
	ErrorTimberConvert     = "timber_convert"
	ErrorProducerCall      = "producer_call"
	ErrorGzipDecompression = "gzip_decompression"
)

func init() {
	InitProducerInstrumentation()
	disable, exists := os.LookupEnv("DISABLE_APP_NAME_LABEL_METRICS")
	if exists && disable == "true" {
		disableAppNameLabelMetrics = true
	}
}

func InitProducerInstrumentation() {
	doubleRouterForward = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_double_router_forward_total",
		Help: "Number of request forwarded to other router",
	}, []string{"app_group", "app_name"})
	forwardToOtherRouterSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_forward_to_other_router_success_total",
		Help: "Number of success request forwarded to other router",
	}, []string{"app_group", "app_name", "router_address"})
	forwardToOtherRouterFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_forward_to_other_router_failed_total",
		Help: "Number of failed request forwarded to other router",
	}, []string{"app_group", "app_name", "router_address"})
	producerRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_producer_request_total",
		Help: "Number request to producer",
	}, []string{"app_group", "app_name", "producer_address"})
	producerRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_producer_request_error_total",
		Help: "Number error request to producer  ",
	}, []string{"app_group", "app_name", "batch", "error", "producer_address"})
	latencyToMarket = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "barito_router_latency_to_market",
		Help:       "Latency to barito market",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	})
	latencyToConsul = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_routerr_latency_to_consul",
		Help:       "Latency to consul",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group", "host"})
	latencyToProducer = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_router_latency_to_producer",
		Help:       "Latency to producer",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group", "app_name", "producer_address"})
	totalLogBytesIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_produced_total_log_bytes",
		Help: "Total log bytes being ingested by the router",
	}, []string{"app_group", "app_name", "producer_address"})
}

func IncreaseProducerRequestCount(appGroup, appName, producerAddress string) {
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	producerRequestCount.WithLabelValues(appGroup, appName, producerAddress).Inc()
}

func IncreaseProducerRequestError(appGroup, appName, producerAddress string, r *http.Request, errorMsg string) {
	batch := "false"
	if r.URL.Path == "/produce_batch" {
		batch = "true"
	}
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	producerRequestError.WithLabelValues(appGroup, appName, batch, errorMsg, producerAddress).Inc()
}

func IncreaseForwardToOtherRouterSuccess(appGroup, appName, routerAddress string) {
	forwardToOtherRouterSuccess.WithLabelValues(appGroup, appName, routerAddress).Inc()
}

func IncreaseForwardToOtherRouterFailed(appGroup, appName, routerAddress string) {
	forwardToOtherRouterFailed.WithLabelValues(appGroup, appName, routerAddress).Inc()
}

func IncreaseDoubleRouterForward(appGroup, appName string) {
	doubleRouterForward.WithLabelValues(appGroup, appName).Inc()
}

func ObserveBaritoMarketLatency(timeDuration time.Duration) {
	latencyToMarket.Observe(timeDuration.Seconds())
}

func ObserveConsulLatency(appGroup, host string, timeDuration time.Duration) {
	latencyToConsul.WithLabelValues(appGroup, host).Observe(timeDuration.Seconds())
}

func ObserveProducerLatency(appGroup, appName, producerAddress string, timeDuration time.Duration) {
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	latencyToProducer.WithLabelValues(appGroup, appName, producerAddress).Observe(timeDuration.Seconds())
}

func ObserveByteIngestion(appGroup, appName, producerAddress string, receivedByte []byte) {
	totalLogBytesIngested.WithLabelValues(appGroup, appName, producerAddress).Add(math.Round(float64(len(receivedByte))))
}
