package metric

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	labelReplacer       = strings.NewReplacer(",", "\\,", "=", "\\=", " ", "\\ ")
	measurementReplacer = strings.NewReplacer(",", "\\,", " ", "\\ ")
	stringReplacer      = strings.NewReplacer("\"", "\\\"")
)

type wireProtocolEncoder struct {
	bytes.Buffer
	fieldCount int
}

func (enc *wireProtocolEncoder) Measurement(name string) {
	_, _ = measurementReplacer.WriteString(enc, name)
}

func (enc *wireProtocolEncoder) Tag(name, value string) {
	_ = enc.WriteByte(',')
	_, _ = labelReplacer.WriteString(enc, name)
	_ = enc.WriteByte('=')
	_, _ = labelReplacer.WriteString(enc, value)
}

func (enc *wireProtocolEncoder) string(s string) {
	enc.WriteByte('"')
	stringReplacer.WriteString(enc, s)
	enc.WriteByte('"')
}

func (enc *wireProtocolEncoder) Field(name string, value interface{}) error {
	if enc.fieldCount == 0 {
		_ = enc.WriteByte(' ')
	} else {
		_ = enc.WriteByte(',')
	}

	enc.fieldCount++

	_, _ = labelReplacer.WriteString(enc, name)
	_ = enc.WriteByte('=')

	switch v := value.(type) {
	case int:
		enc.WriteString(strconv.FormatInt(int64(v), 10))
		enc.WriteByte('i')
	case int32:
		enc.WriteString(strconv.FormatInt(int64(v), 10))
		enc.WriteByte('i')
	case int64:
		enc.WriteString(strconv.FormatInt(int64(v), 10))
		enc.WriteByte('i')
	case uint:
		enc.WriteString(strconv.FormatUint(uint64(v), 10))
		enc.WriteByte('i')
	case uint32:
		enc.WriteString(strconv.FormatUint(uint64(v), 10))
		enc.WriteByte('i')
	case uint64:
		// TODO: What if the number is out of bounds?
		enc.WriteString(strconv.FormatUint(uint64(v), 10))
		enc.WriteByte('i')
	case string:
		enc.string(v)
	case time.Time:
		enc.WriteString(strconv.FormatInt(v.UnixNano(), 10))
		enc.WriteByte('i')
	case time.Duration:
		enc.WriteString(strconv.FormatInt(v.Nanoseconds(), 10))
		enc.WriteByte('i')
	case float64:
		enc.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			enc.WriteByte('t')
		} else {
			enc.WriteByte('f')
		}
	case fmt.Stringer:
		// Stringer should come as late as possible, as types like time.Duration
		// require special handling.
		enc.string(v.String())
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

func (enc *wireProtocolEncoder) Timestamp(t time.Time) {
	enc.WriteByte(' ')
	enc.WriteString(strconv.FormatInt(t.UnixNano(), 10))
}

func (enc *wireProtocolEncoder) Reset() {
	enc.Buffer.Reset()
	enc.fieldCount = 0
}
