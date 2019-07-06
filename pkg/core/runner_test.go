package core

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"github.com/tetratom/cloudsurvey/plugins/source/aws/iam"
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

	t.Run("load source plugin for all sessions in scope", func(t *testing.T) {
		runner := initRunner(`
[[credentials.aws]]
name = "root1"
shared_config = true
scopes = ["all"]
metric_tags.foo = "a"

[[credentials.aws]]
name = "root2"
shared_config = true
scopes = ["all"]
metric_tags.foo = "c"

[[sources.aws_iam_users]]
scopes = ["all"]
metric_tags.bar = "b"
omit_user_tags = true
		`)

		require.Equal(t, 2, len(runner.Sources))
		require.Equal(t, map[string]string{"foo": "a", "bar": "b"}, runner.Sources[0].MetricTags)
		require.IsType(t, (*iam.Users)(nil), runner.Sources[0].Plugin)
		require.Equal(t, map[string]string{"foo": "c", "bar": "b"}, runner.Sources[1].MetricTags)
		require.IsType(t, (*iam.Users)(nil), runner.Sources[1].Plugin)
		require.Equal(t, true, runner.Sources[1].Plugin.(*iam.Users).OmitUserTags)
	})
}
