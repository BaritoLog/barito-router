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
	producerRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_producer_request_total",
		Help: "Number request to producer",
	}, []string{"app_group", "app_name"})
	producerRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_producer_request_error_total",
		Help: "Number error request to producer  ",
	}, []string{"app_group", "app_name", "batch", "error"})
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
	}, []string{"app_group", "app_name"})
	totalLogBytesIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_produced_total_log_bytes",
		Help: "Total log bytes being ingested by the router",
	}, []string{"app_group", "app_name"})
}

func IncreaseProducerRequestCount(appGroup, appName string) {
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	producerRequestCount.WithLabelValues(appGroup, appName).Inc()
}

func IncreaseProducerRequestError(appGroup, appName string, r *http.Request, errorMsg string) {
	batch := "false"
	if r.URL.Path == "/produce_batch" {
		batch = "true"
	}
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	producerRequestError.WithLabelValues(appGroup, appName, batch, errorMsg).Inc()
}

func ObserveBaritoMarketLatency(timeDuration time.Duration) {
	latencyToMarket.Observe(timeDuration.Seconds())
}

func ObserveConsulLatency(appGroup, host string, timeDuration time.Duration) {
	latencyToConsul.WithLabelValues(appGroup, host).Observe(timeDuration.Seconds())
}

func ObserveProducerLatency(appGroup, appName string, timeDuration time.Duration) {
	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}
	latencyToProducer.WithLabelValues(appGroup, appName).Observe(timeDuration.Seconds())
}

func ObserveByteIngestion(appGroup, appName string, receivedByte []byte) {
	totalLogBytesIngested.WithLabelValues(appGroup, appName).Add(math.Round(float64(len(receivedByte))))
}
