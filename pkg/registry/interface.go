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

type InitSource func(creds Credential) Source

// Source is something that can collect metrics. It should submit the metrics
// to the channel provided, which the source should not close. A source returns
// once all of its metrics have been collected.
type Source interface {
	Plugin
	Source(c context.Context, collector metric.Collector) error
}

// InitCredentials returns a partially initialised instance of Credentials.
// If configured with "from", the parent credentials are passed as an argument.
type InitCredentials func(cred Credential) Credentials

type Credential interface{}

type Credentials interface {
	Plugin
	Credentials(c context.Context) (Credential, error)
}
