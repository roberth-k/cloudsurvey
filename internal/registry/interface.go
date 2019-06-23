package registry

import (
	"context"
	"github.com/tetratom/cloudsurvey/internal/metric"
)

// Initializer, when implemented by a plugin, is invoked once after reading
// its configuration.
type Initializer interface {
	Init() error
}

type Plugin interface {
	Description() string
}

type InitSource func(creds Credential) Source

// Source is something that can collect metrics. It should submit the metrics
// to the channel provided, which the source should not close. A source returns
// once all of its metrics have been collected.
type Source interface {
	Plugin
	Source(c context.Context, collector metric.Collector) error
}

type InitCredentials func() Credentials

type Credential interface{}

type Credentials interface {
	Plugin
	Credentials(c context.Context) (Credential, error)
}
