package metrics

import appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"

// NoopCollector is a MetricsCollector that does nothing.
// Useful for testing or when metrics are disabled.
type NoopCollector struct{}

// NewNoopCollector creates a new no-op metrics collector.
func NewNoopCollector() *NoopCollector {
	return &NoopCollector{}
}

func (c *NoopCollector) IncValidate()              {}
func (c *NoopCollector) IncCompile()               {}
func (c *NoopCollector) IncCompileFail()            {}
func (c *NoopCollector) IncCacheHit()               {}
func (c *NoopCollector) IncCacheMiss()              {}
func (c *NoopCollector) ObserveValidateLatency(ms float64) {}
func (c *NoopCollector) ObserveCompileLatency(ms float64)  {}

// Ensure NoopCollector implements MetricsCollector.
var _ appFormula.MetricsCollector = (*NoopCollector)(nil)
