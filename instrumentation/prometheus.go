package instrumentation

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var disableAppNameLabelMetrics bool = false
var producerLatencyToMarket prometheus.Summary
var producerLatencyToConsul *prometheus.SummaryVec
var producerTotalLogBytesIngested *prometheus.CounterVec

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
	producerTotalLogBytesIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "barito_router_produced_total_log_bytes",
		Help: "Total log bytes being ingested by the router",
	}, []string{"app_group", "app_name"})
}

func ObserveBaritoMarketLatency(timeDuration time.Duration) {
	producerLatencyToMarket.Observe(timeDuration.Seconds())
}

func ObserveConsulLatency(appGroup, host string, timeDuration time.Duration) {
	producerLatencyToConsul.WithLabelValues(appGroup, host).Observe(timeDuration.Seconds())
}

func ObserveByteIngestion(appGroup, appName string, receivedByte []byte) {
	producerTotalLogBytesIngested.WithLabelValues(appGroup, appName).Add(float64(len(receivedByte)))
}
