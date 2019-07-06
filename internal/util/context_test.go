package util

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestContextWithNowTime(t *testing.T) {
	t.Run("when time is present", func(t *testing.T) {
		now := time.Now()
		c := ContextWithNowTime(context.Background(), now)
		tm := ContextNowTime(c)
		require.Equal(t, now, tm)
	})

	t.Run("when time is not present", func(t *testing.T) {
		before := time.Now()
		tm := ContextNowTime(context.Background())
		after := time.Now()
		require.True(t, tm.After(before))
		require.True(t, tm.Before(after))
	})
}
