package registry

var (
	credentials = make(map[string]InitCredentials)
	plugins     = make(map[string]InitSource)
)

func AddPlugin(name string, f InitSource) {
	plugins[name] = f
}

func AddCredentials(name string, f InitCredentials) {
	credentials[name] = f
}

func GetCredentials(name string) (InitCredentials, error) {
	return credentials[name], nil
}
