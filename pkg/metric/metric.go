package metric

import (
	"github.com/tetratom/cloudsurvey/internal/util"
	"sort"
	"sync"
	"time"
)

// Datum is a data point extracted by a source plugin. Its shape follows the
// InfluxDB Wire Protocol.
type Datum struct {
	Name   string
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

var (
	encoderPool = sync.Pool{
		New: func() interface{} {
			return &wireProtocolEncoder{}
		},
	}
)

func (m Datum) ToInfluxDBWireProtocol() (string, error) {
	// keys is a slice used for sorting tags and fields
	keys := make([]string, 0, util.MaxInt(len(m.Tags), len(m.Fields)))

	enc := encoderPool.Get().(*wireProtocolEncoder)

	enc.Measurement(m.Name)

	// ensure tags are ordered lexically
	keys = keys[:0]
	for k := range m.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		enc.Tag(k, m.Tags[k])
	}

	// ensure fields are ordered lexically
	keys = keys[:0]
	for k := range m.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := enc.Field(k, m.Fields[k]); err != nil {
			return "", err
		}
	}

	enc.Timestamp(m.Time)

	s := enc.String()
	enc.Reset()
	encoderPool.Put(enc)

	return s, nil
}
