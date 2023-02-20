package transformer

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewCachedResult(t *testing.T) {
	originRegistry := newTestingRegistry()
	ctx := context.Background()
	result, err := NewCachedResult(ctx, originRegistry)
	registry := prometheus.NewRegistry()
	registry.MustRegister(result)

	require.NoError(t, err)
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	metricFamilies2, err := registry.Gather()
	require.NoError(t, err)

	inResult(metricFamilies).
		assertTheFamily(t, "test_metric").
		withMetricThatMatches(t, eq("ifIndex", "1"),
			eq("kopytko", "true"),
			eq("potato", "false")).
		hasValue(t, 42)
	inResult(metricFamilies).
		assertTheFamily(t, "test_metric").
		hasMemberCount(t, 5)
	inResult(metricFamilies2).
		assertTheFamily(t, "test_metric").
		withMetricThatMatches(t, eq("ifIndex", "1"),
			eq("kopytko", "true"),
			eq("potato", "false")).
		hasValue(t, 42)
	inResult(metricFamilies2).
		assertTheFamily(t, "test_metric").
		hasMemberCount(t, 5)
}
