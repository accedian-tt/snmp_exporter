package transformer

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/snmp_exporter/config"
	"github.com/prometheus/snmp_exporter/queryable"
)

type Transformer struct {
	engine *promql.Engine
	source storage.Queryable
	module *config.Module
	ctx    context.Context
}

func New(ctx context.Context, module *config.Module, gatherer prometheus.Gatherer) (Transformer, error) {
	engine := promql.NewEngine(promql.EngineOpts{
		Timeout:    1 * time.Second,
		MaxSamples: 10000,
	})
	source, err := queryable.FromGatherer(gatherer)
	if err != nil {
		return Transformer{}, fmt.Errorf("failed to create source from gatherer: %w", err)
	}
	return Transformer{
		ctx:    ctx,
		engine: engine,
		source: source,
		module: module,
	}, nil
}

// Describe implements Prometheus.Collector.
func (t Transformer) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("transformer", "transformer", nil, nil)
}

// Collect implements Prometheus.Collector.
func (t Transformer) Collect(ch chan<- prometheus.Metric) {
	for _, rule := range t.module.Transform {
		metricName := rule.Name
		expression := rule.Expression

		println(metricName, expression)
		query, _ := t.engine.NewInstantQuery(t.source, nil, expression, time.Now())
		result := query.Exec(t.ctx)
		query.Close()
		toGauge(rule.Name, result).Collect(ch)
	}
}

func toGauge(name string, result *promql.Result) *prometheus.GaugeVec {
	labelNames := collectLabelNames(result)
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
	}, labelNames)

	vector, _ := result.Vector()
	for _, sample := range vector {
		labels := prometheus.Labels{}

		for _, labelName := range labelNames {
			labels[labelName] = ""
		}

		for _, label := range sample.Metric {
			if label.Name == "__name__" {
				continue
			}
			labels[label.Name] = label.Value
		}

		metric.With(labels).Set(sample.V)
	}

	if len(vector) == 0 {
		scalar, _ := result.Scalar()
		labels := prometheus.Labels{}

		metric.With(labels).Set(scalar.V)
	}

	return metric
}

func collectLabelNames(result *promql.Result) []string {
	vector, _ := result.Vector()
	labels := map[string]struct{}{}
	for _, v := range vector {
		for _, label := range v.Metric {
			if label.Name == "__name__" {
				continue
			}
			labels[label.Name] = struct{}{}
		}
	}

	ret := make([]string, len(labels))
	i := 0
	for k := range labels {
		ret[i] = k
		i++
	}

	return ret
}
