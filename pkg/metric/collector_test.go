package metric

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type sliceCollector struct {
	data []Datum
}

func (collector *sliceCollector) Record(datum Datum) {
	collector.data = append(collector.data, datum)
}

func TestMetricTagOverrideCollector(t *testing.T) {
	stash := sliceCollector{}
	collector := MetricTagOverrideCollector{
		Inner: &stash,
		MetricTags: map[string]string{
			"a": "b",
			"c": "d",
		},
	}

	tests := []struct {
		input  Datum
		expect Datum
	}{
		{
			input: Datum{},
			expect: Datum{
				Tags: map[string]string{"a": "b", "c": "d"},
			},
		},
		{
			input: Datum{
				Tags: map[string]string{"a": "x"},
			},
			expect: Datum{
				Tags: map[string]string{"a": "b", "c": "d"},
			},
		},
	}

	for i, test := range tests {
		collector.Record(test.input)
		require.Equal(t, test.expect, stash.data[i])
	}
}
