package queryable

import (
	"time"

	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

func newEntry(value float64, labelPairs ...string) storage.Series {
	metricLabels := labels.FromStrings(labelPairs...)
	return &storage.SeriesEntry{
		Lset: metricLabels,
		SampleIteratorFn: func() chunkenc.Iterator {
			return &SingleValueIterator{value: value, time: time.Now()}
		},
	}
}

type SingleValueIterator struct {
	value  float64
	time   time.Time
	isUsed bool
}

func (s *SingleValueIterator) Next() chunkenc.ValueType {
	if s.isUsed {
		return chunkenc.ValNone
	}
	s.isUsed = true
	return chunkenc.ValFloat
}

func (s *SingleValueIterator) Seek(_ int64) chunkenc.ValueType {
	return chunkenc.ValFloat
}

func (s *SingleValueIterator) At() (int64, float64) {
	return timestamp.FromTime(s.time), s.value
}

func (s *SingleValueIterator) AtHistogram() (int64, *histogram.Histogram) {
	return 0, nil
}

func (s *SingleValueIterator) AtFloatHistogram() (int64, *histogram.FloatHistogram) {
	return 0, nil
}

func (s *SingleValueIterator) AtT() int64 {
	return 0
}

func (s *SingleValueIterator) Err() error {
	return nil
}
