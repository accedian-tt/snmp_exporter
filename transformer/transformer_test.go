package transformer

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/snmp_exporter/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformations(t *testing.T) {
	ctx := context.Background()
	registry := newTestingRegistry()
	m := config.Module{Transform: []config.TransformRule{
		{
			Name:       "simple_scalar",
			Expression: "77",
		},
		{
			Name:       "multiply_scalars",
			Expression: "11 * 3",
		},
		{
			Name:       "simple_sum",
			Expression: "sum(test_metric)",
		},
		{
			Name:       "sum_by_label",
			Expression: "sum(test_metric) by (ifIndex)",
		},
		{
			Name:       "sum_without_label",
			Expression: "sum(test_metric) without (ifIndex, kopytko)",
		},
		{
			Name:       "comparison",
			Expression: "test_metric <= 42",
		},
		{
			Name:       "alternative",
			Expression: "min(test_metric) by (ifIndex) or test_metric2",
		},
		{
			Name:       "select",
			Expression: `test_metric{ifIndex="3"}`,
		},
		{
			Name:       "multiply",
			Expression: "sum(test_metric) by (ifIndex) * test_metric2",
		},
		{
			Name:       "divide",
			Expression: `test_metric{ifIndex="1"}/test_metric{ifIndex="1"}`,
		},
		{
			Name:       "rate",
			Expression: `rate(test_metric{ifIndex="1"}[1h])`,
		},
		{
			Name:       "offset",
			Expression: `test_metric{ifIndex="1"} offset 1d`,
		},
	}}
	transformer, err := New(ctx, &m, registry)
	require.NoError(t, err)

	registry.MustRegister(transformer)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	inResult(metricFamilies).
		assertTheFamily(t, "simple_scalar").
		withSingleMetric(t).
		hasValue(t, 77)

	inResult(metricFamilies).
		assertTheFamily(t, "multiply_scalars").
		withSingleMetric(t).
		hasValue(t, 33)

	inResult(metricFamilies).
		assertTheFamily(t, "simple_sum").
		withSingleMetric(t).
		hasValue(t, 220.1)

	inResult(metricFamilies).
		assertTheFamily(t, "sum_by_label").
		hasMemberCount(t, 3)
	inResult(metricFamilies).
		assertTheFamily(t, "sum_by_label").
		withMetricThatMatches(t, eq("ifIndex", "1")).
		hasValue(t, 42)
	inResult(metricFamilies).
		assertTheFamily(t, "sum_by_label").
		withMetricThatMatches(t, eq("ifIndex", "2")).
		hasValue(t, 43)
	inResult(metricFamilies).
		assertTheFamily(t, "sum_by_label").
		withMetricThatMatches(t, eq("ifIndex", "3")).
		hasValue(t, 135.1)

	inResult(metricFamilies).
		assertTheFamily(t, "sum_without_label").
		hasMemberCount(t, 2)
	inResult(metricFamilies).
		assertTheFamily(t, "sum_without_label").
		withMetricThatMatches(t, eq("potato", "false")).
		hasValue(t, 129)
	inResult(metricFamilies).
		assertTheFamily(t, "sum_without_label").
		withMetricThatMatches(t, eq("potato", "true")).
		hasValue(t, 91.1)

	inResult(metricFamilies).
		assertTheFamily(t, "comparison").
		withMetricThatMatches(t, eq("ifIndex", "1"), eq("kopytko", "true"), eq("potato", "false")).
		hasValue(t, 42)

	inResult(metricFamilies).
		assertTheFamily(t, "alternative").
		hasMemberCount(t, 4)
	inResult(metricFamilies).
		assertTheFamily(t, "alternative").
		withMetricThatMatches(t, eq("ifIndex", "1")).
		hasValue(t, 42)
	inResult(metricFamilies).
		assertTheFamily(t, "alternative").
		withMetricThatMatches(t, eq("ifIndex", "2")).
		hasValue(t, 43)
	inResult(metricFamilies).
		assertTheFamily(t, "alternative").
		withMetricThatMatches(t, eq("ifIndex", "3")).
		hasValue(t, 44)
	inResult(metricFamilies).
		assertTheFamily(t, "alternative").
		withMetricThatMatches(t, eq("ifIndex", "4")).
		hasValue(t, 1)

	inResult(metricFamilies).
		assertTheFamily(t, "select").
		hasMemberCount(t, 3)

	inResult(metricFamilies).
		assertTheFamily(t, "multiply").
		hasMemberCount(t, 3)
	inResult(metricFamilies).
		assertTheFamily(t, "multiply").
		withMetricThatMatches(t, eq("ifIndex", "1")).
		hasValue(t, 42)
	inResult(metricFamilies).
		assertTheFamily(t, "multiply").
		withMetricThatMatches(t, eq("ifIndex", "2")).
		hasValue(t, 43)
	inResult(metricFamilies).
		assertTheFamily(t, "multiply").
		withMetricThatMatches(t, eq("ifIndex", "3")).
		hasValue(t, 0)

	inResult(metricFamilies).
		assertTheFamily(t, "divide").
		withSingleMetric(t).
		hasValue(t, 1)

	inResult(metricFamilies).
		assertTheFamily(t, "rate").
		withSingleMetric(t).
		hasValue(t, 0)

	inResult(metricFamilies).
		assertTheFamily(t, "offset").
		withSingleMetric(t).
		hasValue(t, 42)
}

func TestCienaCes(t *testing.T) {
	ctx := context.Background()
	registry := cienaCesCfmSyntheticLossSessionAvgFrameLoss()
	m := config.Module{Transform: []config.TransformRule{
		{
			Name:       "cienaCesCfmSyntheticLossSessionAvgFrameLossMax",
			Expression: "cienaCesCfmSyntheticLossSessionAvgFrameLossFar > cienaCesCfmSyntheticLossSessionAvgFrameLossNear or cienaCesCfmSyntheticLossSessionAvgFrameLossNear",
		},
		{
			Name:       "cienaCesCfmSyntheticLossSessionAvgFrameLossMax2",
			Expression: `max({__name__=~"cienaCesCfmSyntheticLossSessionAvgFrameLossFar|cienaCesCfmSyntheticLossSessionAvgFrameLossNear"}) without (__name__)`,
		},
	}}
	transformer, err := New(ctx, &m, registry)
	require.NoError(t, err)

	registry.MustRegister(transformer)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	inResult(metricFamilies).
		assertTheFamily(t, "cienaCesCfmSyntheticLossSessionAvgFrameLossMax").
		withSingleMetric(t).
		hasValue(t, 8)

	inResult(metricFamilies).
		assertTheFamily(t, "cienaCesCfmSyntheticLossSessionAvgFrameLossMax2").
		withMetricThatMatches(t,
			eq("cienaCesCfmSyntheticLossSessionLocalMEPId", "1"),
			eq("cienaCesCfmSyntheticLossSessionServiceIndex", "13"),
			eq("cienaCesCfmSyntheticLossSessionTargetMEPId", "2"),
			eq("cienaCesCfmSyntheticLossSessionTestId", "1"),
		).
		hasValue(t, 8)
}

func newTestingRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	testMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "test_metric",
		Help: "This metric will help test transformer",
	}, []string{"ifIndex", "kopytko", "potato"})

	testMetric.With(prometheus.Labels{"ifIndex": "1", "kopytko": "true", "potato": "false"}).Set(42)
	testMetric.With(prometheus.Labels{"ifIndex": "2", "kopytko": "false", "potato": "false"}).Set(43)
	testMetric.With(prometheus.Labels{"ifIndex": "3", "kopytko": "true", "potato": "false"}).Set(44)
	testMetric.With(prometheus.Labels{"ifIndex": "3", "kopytko": "false", "potato": "true"}).Set(45)
	testMetric.With(prometheus.Labels{"ifIndex": "3", "kopytko": "true", "potato": "true"}).Set(46.1)

	registry.MustRegister(testMetric)

	testMetric2 := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "test_metric2",
		Help: "This metric will help test transformer",
	}, []string{"ifIndex"})

	testMetric2.With(prometheus.Labels{"ifIndex": "1"}).Set(1)
	testMetric2.With(prometheus.Labels{"ifIndex": "2"}).Set(1)
	testMetric2.With(prometheus.Labels{"ifIndex": "3"}).Set(0)
	testMetric2.With(prometheus.Labels{"ifIndex": "4"}).Set(1)

	registry.MustRegister(testMetric2)

	return registry
}

func cienaCesCfmSyntheticLossSessionAvgFrameLoss() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	far := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cienaCesCfmSyntheticLossSessionAvgFrameLossFar",
		Help: "far",
	}, []string{
		"cienaCesCfmSyntheticLossSessionLocalMEPId",
		"cienaCesCfmSyntheticLossSessionServiceIndex",
		"cienaCesCfmSyntheticLossSessionTargetMEPId",
		"cienaCesCfmSyntheticLossSessionTestId",
		"cienaCesPmInstanceName",
	})

	far.With(prometheus.Labels{
		"cienaCesCfmSyntheticLossSessionLocalMEPId":   "1",
		"cienaCesCfmSyntheticLossSessionServiceIndex": "13",
		"cienaCesCfmSyntheticLossSessionTargetMEPId":  "2",
		"cienaCesCfmSyntheticLossSessionTestId":       "1",
		"cienaCesPmInstanceName":                      "25_CEL21-0001631_1",
	}).Set(1)

	registry.MustRegister(far)

	Near := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cienaCesCfmSyntheticLossSessionAvgFrameLossNear",
		Help: "near",
	}, []string{
		"cienaCesCfmSyntheticLossSessionLocalMEPId",
		"cienaCesCfmSyntheticLossSessionServiceIndex",
		"cienaCesCfmSyntheticLossSessionTargetMEPId",
		"cienaCesCfmSyntheticLossSessionTestId",
		"cienaCesPmInstanceName",
	})

	Near.With(prometheus.Labels{
		"cienaCesCfmSyntheticLossSessionLocalMEPId":   "1",
		"cienaCesCfmSyntheticLossSessionServiceIndex": "13",
		"cienaCesCfmSyntheticLossSessionTargetMEPId":  "2",
		"cienaCesCfmSyntheticLossSessionTestId":       "1",
		"cienaCesPmInstanceName":                      "25_CEL21-0001631_1",
	}).Set(8)

	registry.MustRegister(Near)

	return registry
}

type familiesConditions []*io_prometheus_client.MetricFamily
type familyConditions io_prometheus_client.MetricFamily
type metricConditions io_prometheus_client.Metric

func inResult(t []*io_prometheus_client.MetricFamily) familiesConditions {
	return t
}

func (f familiesConditions) assertTheFamily(t *testing.T, name string) *familyConditions {
	if f == nil {
		return nil
	}
	for _, family := range f {
		if family.GetName() == name {
			return (*familyConditions)(family)
		}
	}

	assert.Failf(t, "missing metric family", "Missing family: %v", name)
	return nil
}

func (f *familyConditions) withMetricThatMatches(t *testing.T, matchers ...*labels.Matcher) *metricConditions {
	if f == nil {
		return nil
	}

	var result *metricConditions

	for _, metric := range f.Metric {
		if matches(metric, matchers) {
			if result != nil {
				assert.Failf(t, "duplicated metric", "In family %s there are multiple metrics: %v", *f.Name, matchers)
				return nil
			}
			result = (*metricConditions)(metric)
		}
	}

	if result == nil {
		assert.Failf(t, "missing metric", "In family %s there is no metric that matches: %v", *f.Name, matchers)
	}
	return result
}

func matches(metric *io_prometheus_client.Metric, matchers []*labels.Matcher) bool {
matcherLoop:
	for _, matcher := range matchers {
		for _, label := range metric.Label {
			if matcher.Name == label.GetName() && matcher.Matches(label.GetValue()) {
				continue matcherLoop
			}
		}
		return false
	}

	return true
}

func (f *familyConditions) hasMemberCount(t *testing.T, count int) {
	if f == nil {
		return
	}
	assert.Len(t, f.Metric, count)
}

func (f *familyConditions) withSingleMetric(t *testing.T) *metricConditions {
	if f == nil {
		return nil
	}
	if len(f.Metric) != 1 {
		assert.Failf(t, "unexpected metric family", "Expected family %v has a single metric but it has: %v", *f.Name, len(f.Metric))
		return nil
	}
	return (*metricConditions)(f.Metric[0])
}

func (m *metricConditions) hasValue(t *testing.T, expectedValue float64) {
	if m == nil {
		return
	}
	assert.Equal(t, expectedValue, m.Gauge.GetValue(), m.Label)
}

func eq(name, value string) *labels.Matcher {
	return labels.MustNewMatcher(labels.MatchEqual, name, value)
}
