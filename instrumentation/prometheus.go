package instrumentation

import (
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	pb "github.com/vwidjaya/barito-proto/producer"
)

var disableAppNameLabelMetrics bool = false

var producerRequestCount *prometheus.CounterVec
var producerRequestError *prometheus.CounterVec
var producerNumberMessagePerBatch *prometheus.SummaryVec
var producerLengthPerMessage *prometheus.SummaryVec
var producerLatencyToMarket prometheus.Summary
var producerLatencyToConsul *prometheus.SummaryVec
var producerLatencyToProducer *prometheus.SummaryVec

// list error message for producerRequestError metrics
const (
	ErrorFetchProfile  = "fetch_profile"
	ErrorConsulCall    = "consul_call"
	ErrorNoProducer    = "no_producer"
	ErrorTimberConvert = "timber_convert"
	ErrorProducerCall  = "producer_call"
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
	producerNumberMessagePerBatch = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_producer_message_per_batch",
		Help:       "Number message per batch",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group", "app_name"})
	producerLengthPerMessage = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_producer_length_per_message",
		Help:       "Number of length per message",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group", "app_name"})
	producerLatencyToMarket = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "barito_producer_latency_to_market",
		Help:       "Latency to barito market",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	})
	producerLatencyToConsul = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_producer_latency_to_consul",
		Help:       "Latency to consul",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group", "host"})
	producerLatencyToProducer = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "barito_producer_latency_to_producer",
		Help:       "Latency to producer",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     1 * time.Minute,
	}, []string{"app_group"})
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
	producerLatencyToMarket.Observe(timeDuration.Seconds())
}

func ObserveConsulLatency(appGroup, host string, timeDuration time.Duration) {
	producerLatencyToConsul.WithLabelValues(appGroup, host).Observe(timeDuration.Seconds())
}

func ObserveProducerLatency(appGroup string, timeDuration time.Duration) {
	producerLatencyToProducer.WithLabelValues(appGroup).Observe(timeDuration.Seconds())
}

func ObserveTimberCollection(appGroup, appName string, timberCollection *pb.TimberCollection) {
	length := len(timberCollection.Items)

	if disableAppNameLabelMetrics {
		appName = "GLOBAL"
	}

	producerNumberMessagePerBatch.WithLabelValues(appGroup, appName).Observe(float64(length))

	for i := 0; i < length; i++ {
		producerLengthPerMessage.
			WithLabelValues(appGroup, appName).
			Observe(float64(len(timberCollection.Items[i].Content.String())))
	}
}
