package metric

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once    sync.Once
	histo   *prometheus.HistogramVec
	counter *prometheus.CounterVec
)

func init() {
	once.Do(func() {
		histo = newHistogramVec()
		prometheus.MustRegister(histo)

		counter = newCounterVec()
		prometheus.MustRegister(counter)
	})
}

// Handler is an http.HandlerFunc to serve metrics endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

// TraceRequestTime is method to write metrics to prometheus
func TraceRequestTime(method, action, status string, elapsedTime float64) {
	histo.WithLabelValues(method, action, status).Observe(elapsedTime)
}

func IncrementByOne(entity, status string) {
	counter.WithLabelValues(entity, status).Inc()
}

func newHistogramVec() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "service_latency_seconds",
		Help: "service events response in miliseconds",
	}, []string{"method", "action", "status"})
}

func newCounterVec() *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "entity_total",
		Help: "number of entity recorded based on its status",
	}, []string{"status", "entity"})
}
