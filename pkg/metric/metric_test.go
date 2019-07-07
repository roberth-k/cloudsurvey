package metric

import (
	"github.com/stretchr/testify/require"
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
			s, err := test.d.ToInfluxDBWireProtocol()
			require.NoError(t, err)
			require.Equal(t, test.s, s)
		})
	}
}
