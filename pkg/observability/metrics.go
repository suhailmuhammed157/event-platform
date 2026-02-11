package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ProcessedEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "events_processed_total",
		Help: "Total number of events successfully processed",
	})

	RetryEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "events_retry_total",
		Help: "Total number of retryable events",
	})

	DLQEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "events_dlq_total",
		Help: "Total number of events sent to DLQ",
	})

	ProcessingLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "processing_latency_seconds",
		Help:    "Time to process a single event",
		Buckets: prometheus.LinearBuckets(0.001, 0.01, 10),
	})
)

func Init() {
	prometheus.MustRegister(ProcessedEvents, RetryEvents, DLQEvents, ProcessingLatency)
}

// Expose /metrics HTTP handler on the given port
func ServeMetrics(port string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}
