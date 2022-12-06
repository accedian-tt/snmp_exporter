package queryable

import (
	"sort"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

type stringSet map[string]struct{}
type seriesByLabel []storage.Series

func (a seriesByLabel) Len() int           { return len(a) }
func (a seriesByLabel) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a seriesByLabel) Less(i, j int) bool { return labels.Compare(a[i].Labels(), a[j].Labels()) < 0 }

func (s stringSet) Flat() []string {
	result := make([]string, 0, len(s))

	for k := range s {
		result = append(result, k)
	}

	return result
}

func allMatches(s storage.Series, matchers []*labels.Matcher) bool {
	for _, label := range s.Labels() {
		for _, matcher := range matchers {
			if matcher.Name == label.Name && !matcher.Matches(label.Value) {
				return false
			}
		}
	}

	return true
}

type instantQuerier struct {
	instances []storage.Series
}

func (m *instantQuerier) LabelValues(name string, matchers ...*labels.Matcher) ([]string, storage.Warnings, error) {
	matchingLabelValues := stringSet{}
	for _, matcher := range matchers {
		for _, instance := range m.instances {
			for _, label := range instance.Labels() {
				if label.Name == name && matcher.Matches(label.Value) {
					matchingLabelValues[label.Value] = struct{}{}
				}
			}
		}
	}

	return matchingLabelValues.Flat(), nil, nil
}

func (m *instantQuerier) LabelNames(matchers ...*labels.Matcher) ([]string, storage.Warnings, error) {
	matchingLabelNames := stringSet{}
	for _, matcher := range matchers {
		for _, instance := range m.instances {
			for _, label := range instance.Labels() {
				if matcher.Name == label.Name {
					matchingLabelNames[label.Name] = struct{}{}
				}
			}
		}
	}

	return matchingLabelNames.Flat(), nil, nil
}

func (m *instantQuerier) selectByMatchers(matchers []*labels.Matcher) []storage.Series {
	var selected []storage.Series

	for _, instance := range m.instances {
		if allMatches(instance, matchers) {
			selected = append(selected, instance)
		}
	}

	return selected
}

func (m *instantQuerier) Close() error {
	return nil
}

func (m *instantQuerier) Select(sortSeries bool, _ *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	selected := m.selectByMatchers(matchers)

	if sortSeries {
		sort.Sort(seriesByLabel(selected))
	}

	return newInstantSeriesSet(selected...)
}
