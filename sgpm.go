package sgpm

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Config for Middleware
type Config struct {
	Prefix          string
	Source          string
	SourceLabel     string
	HandlerLabel    string
	MethodLabel     string
	StatusCodeLabel string
	DurationBuckets []float64
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

// Middleware collect metrics about requests for prometheus
func Middleware(cfg Config) func(http.Handler) http.Handler {
	durationHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.Prefix,
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "The duration of the HTTP requests.",
		Buckets:   cfg.DurationBuckets,
	}, []string{cfg.HandlerLabel, cfg.SourceLabel, cfg.MethodLabel, cfg.StatusCodeLabel})

	requestsInflight := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: cfg.Prefix,
		Subsystem: "http",
		Name:      "requests_inflight",
		Help:      "The number of inflight requests.",
	}, []string{cfg.HandlerLabel, cfg.SourceLabel})

	prometheus.DefaultRegisterer.MustRegister(
		durationHistogram,
		requestsInflight,
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestsInflight.WithLabelValues(r.URL.Path, cfg.Source).Add(float64(1))

			begin := time.Now()

			wi := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(wi, r)

			durationHistogram.WithLabelValues(r.URL.Path, cfg.Source, r.Method, strconv.Itoa(wi.status)).Observe(time.Now().Sub(begin).Seconds())
			requestsInflight.WithLabelValues(r.URL.Path, cfg.Source).Add(float64(-1))
		})
	}
}
