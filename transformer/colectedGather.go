package transformer

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func NewCachedResult(ctx context.Context, c prometheus.Collector) (prometheus.Collector, error) {
	cache := collectedGatherer{
		metrics: []prometheus.Metric{},
		descs:   []*prometheus.Desc{},
	}

	origin := prometheus.NewRegistry()
	origin.MustRegister(c)
	gather, err := origin.Gather()
	if err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	for _, family := range gather {
		for _, metric := range family.GetMetric() {
			desc := prometheus.NewDesc(family.GetName(), family.GetHelp(), labelNames(metric), prometheus.Labels{})
			cache.descs = append(cache.descs, desc)

			var constMetric prometheus.Metric

			switch {
			case metric.Gauge != nil:
				constMetric, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, metric.GetGauge().GetValue(), labelValues(metric)...)
				if err != nil {
					return nil, err
				}
			case metric.Counter != nil:
				constMetric, err = prometheus.NewConstMetric(desc, prometheus.CounterValue, metric.GetCounter().GetValue(), labelValues(metric)...)
				if err != nil {
					return nil, err
				}
			case metric.Untyped != nil:
				constMetric, err = prometheus.NewConstMetric(desc, prometheus.UntypedValue, metric.GetUntyped().GetValue(), labelValues(metric)...)
				if err != nil {
					return nil, err
				}
			}

			cache.metrics = append(cache.metrics, constMetric)
		}
	}

	return cache, nil
}

func labelValues(m *io_prometheus_client.Metric) []string {
	result := []string{}

	for _, l := range m.GetLabel() {
		result = append(result, l.GetValue())
	}

	return result
}

func labelNames(m *io_prometheus_client.Metric) []string {
	result := []string{}

	for _, l := range m.GetLabel() {
		result = append(result, l.GetName())
	}

	return result
}

type collectedGatherer struct {
	descs   []*prometheus.Desc
	metrics []prometheus.Metric
}

func (c collectedGatherer) Describe(descs chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		descs <- desc
	}
}

func (c collectedGatherer) Collect(metrics chan<- prometheus.Metric) {
	for _, metric := range c.metrics {
		metrics <- metric
	}
}
