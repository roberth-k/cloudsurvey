package metric

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWireProtocolEncoder_Measurement(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"test", "test"},
		{"tes t", "tes\\ t"},
		{"tes,t", "tes\\,t"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var enc wireProtocolEncoder
			enc.Measurement(test.input)
			require.Equal(t, test.expect, enc.String())
		})
	}
}

func TestWireProtocolEncoder_Tag(t *testing.T) {
	tests := []struct {
		k string
		v string
		s string
	}{
		{`f`, `a`, `,f=a`},
		{`f,`, `a,`, `,f\,=a\,`},
		{`f=`, `a=`, `,f\==a\=`},
		{`f `, `a `, `,f\ =a\ `},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			var enc wireProtocolEncoder
			enc.Tag(test.k, test.v)
			require.Equal(t, test.s, enc.String())
		})
	}
}

func TestWireProtocolEncoder_Field(t *testing.T) {
	tests := []struct {
		k string
		v interface{}
		s string
	}{
		{"f=", int(0), " f\\==0i"},
		{"f ", int(0), " f\\ =0i"},
		{"f,", int(0), " f\\,=0i"},
		{"f", int(0), " f=0i"},
		{"f", int(1), " f=1i"},
		{"f", int64(0), " f=0i"},
		{"f", int64(1), " f=1i"},
		{"f", float64(0.0), " f=0"},
		{"f", float64(1.0), " f=1"},
		{"f", float64(1.1), " f=1.1"},
		{"f", float64(1.5), " f=1.5"},
		{"f", `test`, ` f="test"`},
		{"f", `"test"`, ` f="\"test\""`},
		{"f", true, " f=t"},
		{"f", false, " f=f"},
		{"f", 2 * time.Second, " f=2000000000i"},
		{"f", time.Date(2019, 1, 2, 3, 4, 5, 0, time.UTC), " f=1546398245000000000i"},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			var enc wireProtocolEncoder
			err := enc.Field(test.k, test.v)
			require.NoError(t, err)
			require.Equal(t, test.s, enc.String())
		})
	}
}
