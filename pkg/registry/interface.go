package registry

import (
	"context"
	"github.com/tetratom/cloudsurvey/pkg/metric"
)

// Initializer, when implemented by a plugin, is invoked once after reading
// its configuration.
type Initializer interface {
	Init() error
}

type Plugin interface {
	Description() string
}

type InitSource func(cred Session) Source

// Plugin is something that can collect metrics. It should submit the metrics
// to the channel provided, which the source should not close. A source returns
// once all of its metrics have been collected.
//
// A source may be long-lived, and partake in several executions with
// significant time-spacing.
type Source interface {
	Plugin
	Source(c context.Context, collector metric.Collector) error
}

// InitCredentials returns a partially initialised instance of Plugin.
// If configured with "from", the parent credentials are passed as an argument.
type InitCredentials func(cred Session) Credentials

type Session interface{}

type Credentials interface {
	Plugin
	Credentials(c context.Context) (Session, error)
}
