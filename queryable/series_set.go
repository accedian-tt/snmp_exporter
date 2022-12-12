package queryable

import "github.com/prometheus/prometheus/storage"

type instantSeriesSet struct {
	idx    int
	series []storage.Series
}

func (m *instantSeriesSet) At() storage.Series { return m.series[m.idx] }

func newInstantSeriesSet(series ...storage.Series) storage.SeriesSet {
	return &instantSeriesSet{
		idx:    -1,
		series: series,
	}
}

func (m *instantSeriesSet) Next() bool {
	m.idx++
	return m.idx < len(m.series)
}

func (m *instantSeriesSet) Err() error { return nil }

func (m *instantSeriesSet) Warnings() storage.Warnings { return nil }
