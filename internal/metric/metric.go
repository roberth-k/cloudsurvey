package metric

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Datum is a data point extracted by a source plugin. Its shape follows the
// InfluxDB Line Protocol.
type Datum struct {
	Name   string
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

var (
	stringBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
)

func (m Datum) ToInfluxDBLineProtocol() (string, error) {
	// keys is a slice used for sorting tags and fields
	keys := make([]string, 0, maxInt(len(m.Tags), len(m.Fields)))

	w := stringBuilderPool.Get().(*strings.Builder)

	w.WriteString(m.Name)

	// ensure tags are ordered lexically
	keys = keys[:0]
	for k := range m.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m.Tags[k]
		w.WriteByte(',')
		w.WriteString(k)
		w.WriteByte('=')
		w.WriteString(v)
	}

	w.WriteByte(' ')

	// ensure fields are ordered lexically
	keys = keys[:0]
	for k := range m.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			w.WriteByte(',')
		}

		v := m.Fields[k]
		w.WriteString(k)
		w.WriteByte('=')
		if err := encodeField(v, w); err != nil {
			return "", err
		}
	}

	w.WriteByte(' ')
	w.WriteString(strconv.FormatInt(m.Time.UnixNano(), 10))

	s := w.String()
	w.Reset()
	stringBuilderPool.Put(w)

	return s, nil
}

func encodeField(value interface{}, b *strings.Builder) error {
	switch v := value.(type) {
	case int:
		b.WriteString(strconv.FormatInt(int64(v), 10))
		b.WriteByte('i')
	case int32:
		b.WriteString(strconv.FormatInt(int64(v), 10))
		b.WriteByte('i')
	case int64:
		b.WriteString(strconv.FormatInt(int64(v), 10))
		b.WriteByte('i')
	case uint:
		b.WriteString(strconv.FormatUint(uint64(v), 10))
		b.WriteByte('i')
	case uint32:
		b.WriteString(strconv.FormatUint(uint64(v), 10))
		b.WriteByte('i')
	case uint64:
		// TODO: What if the number is out of bounds?
		b.WriteString(strconv.FormatUint(uint64(v), 10))
		b.WriteByte('i')
	case string:
		b.WriteByte('"')
		b.WriteString(v) // TODO: quote the string.
		b.WriteByte('"')
	case time.Time:
		b.WriteString(strconv.FormatInt(v.UnixNano(), 10))
		b.WriteByte('i')
	case time.Duration:
		b.WriteString(strconv.FormatInt(v.Nanoseconds(), 10))
		b.WriteByte('i')
	case float64:
		b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			b.WriteByte('t')
		} else {
			b.WriteByte('f')
		}
	case fmt.Stringer:
		// Stringer should come as late as possible, as types like time.Duration
		// require special handling.
		b.WriteByte('"')
		b.WriteString(v.String()) // TODO: quote the string
		b.WriteByte('"')
	case nil:
		return errors.New("field value is nil")
	default:
		name := "(unknown)"
		if t := reflect.TypeOf(value); t != nil {
			name = t.String()
		}

		return errors.Errorf("unknown field type: %s", name)
	}

	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
