package config

import (
	"github.com/pelletier/go-toml"
	"io"
	"log"
)

const (
	EnvVarPrefix = "CLOUDSURVEY_"
)

type Config struct {
	Main        Main                     `toml:"main"`
	Credentials map[string][]*Credential `toml:"credentials"`
	Sources     map[string][]*Source     `toml:"sources"`
}

type Main struct {
	Verbose bool `toml:"verbose"`
}

type Credential struct {
	Name       string            `toml:"name"`
	From       string            `toml:"from"`
	Scopes     []string          `toml:"scopes"`
	MetricTags map[string]string `toml:"metric_tags"`

	// full representation of the underlying toml structure for
	// configuring credential plugins
	tree *toml.Tree
}

func (c *Credential) Configure(x interface{}) error {
	return c.tree.Unmarshal(x)
}

type Source struct {
	Name       string            `toml:"name"`
	Scopes     []string          `toml:"scopes"`
	MetricTags map[string]string `toml:"metric_tags"`

	// full representation of the underlying toml structure for
	// configuring source plugins
	tree *toml.Tree
}

func (s *Source) Configure(x interface{}) error {
	return s.tree.Unmarshal(x)
}

func FromTree(tree *toml.Tree) (*Config, error) {
	if err := ApplyEnvironmentVariables(tree); err != nil {
		return nil, err
	}

	var config Config
	if err := tree.Unmarshal(&config); err != nil {
		return nil, err
	}

	for k, vs := range config.Credentials {
		for i := range vs {
			// bit of a workaround, as tree.Get() doesn't appear to
			// support indexed slice access
			slice := tree.Get("credentials." + k).([]*toml.Tree)
			vs[i].tree = slice[i]

			if vs[i].MetricTags == nil {
				vs[i].MetricTags = make(map[string]string)
			}
		}
	}

	for k, vs := range config.Sources {
		for i := range vs {
			// bit of a workaround, as tree.Get() doesn't appear to
			// support indexed slice access
			slice := tree.Get("sources." + k).([]*toml.Tree)
			vs[i].tree = slice[i]

			if vs[i].MetricTags == nil {
				vs[i].MetricTags = make(map[string]string)
			}
		}
	}

	return &config, nil
}

func FromFile(path string) (*Config, error) {
	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}

	return FromTree(tree)
}

func FromReader(r io.Reader) (*Config, error) {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return nil, err
	}

	return FromTree(tree)
}

func FromString(content string) (*Config, error) {
	tree, err := toml.Load(content)
	if err != nil {
		return nil, err
	}

	return FromTree(tree)
}

func FromBytes(bytes []byte) (*Config, error) {
	tree, err := toml.LoadBytes(bytes)
	if err != nil {
		return nil, err
	}

	return FromTree(tree)
}

// ApplyEnvironmentVariables will overwrite any values in the tree for which
// an environmental override of the form CLOUDSURVEY_... is found. Overrides
// follow the same conventions as InfluxDB.
func ApplyEnvironmentVariables(tree *toml.Tree) error {
	log.Print("warning: ApplyEnvironmentVariables is not implemented")
	return nil
}
