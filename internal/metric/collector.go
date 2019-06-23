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
