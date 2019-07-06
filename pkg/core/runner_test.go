package core

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"testing"
)

func TestNewRunner(t *testing.T) {
	initRunner := func(configString string) *Runner {
		conf, err := config.FromString(configString)
		require.NoError(t, err)
		runner, err := NewRunner(context.Background(), conf)
		require.NoError(t, err)
		return runner
	}

	t.Run("dependent sessions", func(t *testing.T) {
		runner := initRunner(`
[[credentials.aws]]
name = "root"
shared_config = true

[[credentials.aws]]
from = "root"
profile = "bar"
		`)

		require.Equal(t, 2, len(runner.Sessions))
		require.Equal(t, 0, len(runner.Sources))
		require.Equal(t, "root", runner.Sessions[0].Name)
		require.Equal(t, "", runner.Sessions[1].Name)
	})
}
