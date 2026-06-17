package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
)

const (
	namespace = "quantlab"
	subsystem = "formula_engine"
)

// PrometheusCollector implements MetricsCollector using Prometheus counters and histograms.
type PrometheusCollector struct {
	validateTotal    prometheus.Counter
	compileTotal     prometheus.Counter
	compileFailTotal prometheus.Counter
	cacheHitTotal    prometheus.Counter
	cacheMissTotal   prometheus.Counter
	validateLatency  prometheus.Histogram
	compileLatency   prometheus.Histogram
}

// NewPrometheusCollector creates and registers Prometheus metrics.
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		validateTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "validate_total",
			Help:      "Total number of formula validations.",
		}),
		compileTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "compile_total",
			Help:      "Total number of formula compilations.",
		}),
		compileFailTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "compile_fail_total",
			Help:      "Total number of failed formula compilations.",
		}),
		cacheHitTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_hit_total",
			Help:      "Total number of cache hits.",
		}),
		cacheMissTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_miss_total",
			Help:      "Total number of cache misses.",
		}),
		validateLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "validate_latency_ms",
			Help:      "Latency of formula validation in milliseconds.",
			Buckets:   prometheus.DefBuckets,
		}),
		compileLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "compile_latency_ms",
			Help:      "Latency of formula compilation in milliseconds.",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}

func (c *PrometheusCollector) IncValidate()              { c.validateTotal.Inc() }
func (c *PrometheusCollector) IncCompile()               { c.compileTotal.Inc() }
func (c *PrometheusCollector) IncCompileFail()            { c.compileFailTotal.Inc() }
func (c *PrometheusCollector) IncCacheHit()               { c.cacheHitTotal.Inc() }
func (c *PrometheusCollector) IncCacheMiss()              { c.cacheMissTotal.Inc() }
func (c *PrometheusCollector) ObserveValidateLatency(ms float64) { c.validateLatency.Observe(ms) }
func (c *PrometheusCollector) ObserveCompileLatency(ms float64)  { c.compileLatency.Observe(ms) }

// Ensure PrometheusCollector implements MetricsCollector.
var _ appFormula.MetricsCollector = (*PrometheusCollector)(nil)
