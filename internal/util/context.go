package util

import (
	"context"
	"time"
)

type contextNowTimeKey struct{}

var contextNowTimeKeyValue contextNowTimeKey

func ContextWithNowTime(c context.Context, t time.Time) context.Context {
	return context.WithValue(c, contextNowTimeKeyValue, &t)
}

func ContextNowTime(c context.Context) time.Time {
	if t, ok := c.Value(contextNowTimeKeyValue).(*time.Time); ok {
		return *t
	} else {
		return time.Now()
	}
}
