package pizza

type MetricsRegistry interface {
	NewCounterMetric(name string, labels map[string]string) CounterMetric
	NewGaugeMetric(name string, labels map[string]string) GaugeMetric
}

type CounterMetric interface {
	Increment()
}

type GaugeMetric interface {
	Set(float64)
}
