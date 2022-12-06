package queryable

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/storage"
)

func FromGatherer(gatherer prometheus.Gatherer) (storage.QueryableFunc, error) {
	families, err := gatherer.Gather()

	if err != nil {
		return nil, fmt.Errorf("unable to gather metrics: %w", err)
	}

	return func(_ context.Context, _, _ int64) (storage.Querier, error) {
		var instances []storage.Series
		for _, family := range families {
			for _, metric := range family.Metric {
				instances = append(instances, toEntry(family.GetName(), metric))
			}
		}

		return &instantQuerier{instances: instances}, nil
	}, nil
}

func toEntry(name string, metric *io_prometheus_client.Metric) storage.Series {

	value := getValue(metric)
	var labels []string

	labels = append(labels, "__name__", name)

	for _, pair := range metric.Label {
		labels = append(labels, pair.GetName(), pair.GetValue())
	}

	return newEntry(value, labels...)
}

func getValue(metric *io_prometheus_client.Metric) float64 {
	switch {
	case metric.Gauge != nil:
		return metric.GetGauge().GetValue()
	case metric.Counter != nil:
		return metric.GetCounter().GetValue()
	case metric.Untyped != nil:
		return metric.GetUntyped().GetValue()
	}

	return 0
}
