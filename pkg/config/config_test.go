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
			tags = ["a", "b"]
			yotta = true
			`,
			Config{
				Main: Main{
					Verbose: true,
				},
				Credentials: map[string][]Credential{
					"aws": {
						{
							From: "foo",
							Tags: []string{"a", "b"},
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

			conf.tree = nil
			for _, vs := range conf.Credentials {
				for i := range vs {
					vs[i].tree = nil
				}
			}

			require.Equal(t, test.expect, *conf)
		})
	}
}
