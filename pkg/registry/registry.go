package registry

var (
	credentials = make(map[string]InitCredentials)
	sources     = make(map[string]InitSource)
)

func AddSource(name string, f InitSource) {
	sources[name] = f
}

func AddCredentials(name string, f InitCredentials) {
	credentials[name] = f
}

func GetCredentials(name string) (InitCredentials, error) {
	return credentials[name], nil
}
