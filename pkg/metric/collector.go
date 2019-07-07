package metric

// Collector is something that a source invokes to send metrics.
type Collector interface {
	Record(datum Datum)
}

// ChannelCollector is a collector that sends data directly to a chan. If the
// chan is full, the collector will block.
type ChannelCollector chan<- Datum

func (cc ChannelCollector) Record(datum Datum) {
	cc <- datum
}

// MetricTagOverrideCollector wraps another Collector, but modifies the data
// passing through by appending or overriding the given MetricTags.
type MetricTagOverrideCollector struct {
	Inner      Collector
	MetricTags map[string]string
}

func (collector MetricTagOverrideCollector) Record(datum Datum) {

	if datum.Tags == nil {
		datum.Tags = make(map[string]string)
	}

	for k, v := range collector.MetricTags {
		datum.Tags[k] = v
	}

	collector.Inner.Record(datum)
}
