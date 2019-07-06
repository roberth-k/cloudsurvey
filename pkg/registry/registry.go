package registry

import (
	"github.com/pkg/errors"
)

var (
	credentials = make(map[string]InitCredentials)
	sources     = make(map[string]InitSource)
)

func AddSource(name string, f InitSource) {
	sources[name] = f
}

func GetSource(name string) (InitSource, error) {
	source, ok := sources[name]

	if !ok {
		return nil, errors.Errorf("source plugin not found: %s", name)
	}

	return source, nil
}

func AddCredentials(name string, f InitCredentials) {
	credentials[name] = f
}

func GetCredentials(name string) (InitCredentials, error) {
	cred, ok := credentials[name]

	if !ok {
		return nil, errors.Errorf("credential plugin not found: %s", name)
	}

	return cred, nil
}
