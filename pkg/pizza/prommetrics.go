package pizza

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusRegistry struct {
	reg *prometheus.Registry
}

func NewPrometheusRegistry() *PrometheusRegistry {
	return &PrometheusRegistry{
		reg: prometheus.NewRegistry(),
	}
}

func (reg *PrometheusRegistry) Serve(port int) {
	slog.Info("serving metrics", "port", port)
	http.Handle("/metrics", promhttp.HandlerFor(reg.reg, promhttp.HandlerOpts{Registry: reg.reg}))
	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("prometheus serve failure", "error", err)
	}
}

func (reg *PrometheusRegistry) NewCounterMetric(name string, labels map[string]string) CounterMetric {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: name, ConstLabels: labels})
	reg.reg.MustRegister(counter)
	return &PrometheusCounterMetric{counter: counter}
}

func (reg *PrometheusRegistry) NewGaugeMetric(name string, labels map[string]string) GaugeMetric {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, ConstLabels: labels})
	reg.reg.MustRegister(gauge)
	return &PrometheusGaugeMetric{gauge: gauge}
}

type PrometheusCounterMetric struct {
	counter prometheus.Counter
}

func (m *PrometheusCounterMetric) Increment() {
	m.counter.Inc()
}

type PrometheusGaugeMetric struct {
	gauge prometheus.Gauge
}

func (m *PrometheusGaugeMetric) Set(value float64) {
	m.gauge.Set(value)
}
