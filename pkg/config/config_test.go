package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFromString(t *testing.T) {
	tests := []struct {
		input  string
		expect Config
	}{
		{
			`
			[main]
			verbose = true
			
			[[credentials.aws]]
			from = "foo"
			scopes = ["a", "b"]
			metric_tags.my_tag = "z"
			ignoreme = true
			`,
			Config{
				Main: Main{
					Verbose: true,
				},
				Credentials: map[string][]*Credential{
					"aws": {
						{
							From:       "foo",
							Scopes:     []string{"a", "b"},
							MetricTags: map[string]string{"my_tag": "z"},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			conf, err := FromString(test.input)
			require.NoError(t, err)

			for _, vs := range conf.Credentials {
				for i := range vs {
					vs[i].tree = nil
				}
			}

			require.Equal(t, test.expect, *conf)
		})
	}
}
