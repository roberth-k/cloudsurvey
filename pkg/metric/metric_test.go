package metric

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestDatum_ToInfluxDBLineProtocol(t *testing.T) {
	tm := time.Date(2019, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		s string
		d Datum
	}{
		{
			"test,tag1=a,tag2=b field1=1i,field2=t 1546398245000000000",
			Datum{
				Name:   "test",
				Time:   tm,
				Tags:   map[string]string{"tag1": "a", "tag2": "b"},
				Fields: map[string]interface{}{"field1": 1, "field2": true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			s, err := test.d.ToInfluxDBLineProtocol()
			require.NoError(t, err)
			require.Equal(t, test.s, s)
		})
	}
}

func TestEncodeField(t *testing.T) {
	tests := []struct {
		v interface{}
		s string
	}{
		{int(0), "0i"},
		{int(1), "1i"},
		{int64(0), "0i"},
		{int64(1), "1i"},
		{float64(0.0), "0"},
		{float64(1.0), "1"},
		{float64(1.1), "1.1"},
		{float64(1.5), "1.5"},
		{"test", `"test"`},
		{true, "t"},
		{false, "f"},
		{2 * time.Second, "2000000000i"},
		{time.Date(2019, 1, 2, 3, 4, 5, 0, time.UTC), "1546398245000000000i"},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			w := strings.Builder{}
			err := encodeField(test.v, &w)
			require.NoError(t, err)
			require.Equal(t, test.s, w.String())
		})
	}
}
