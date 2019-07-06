package config

import (
	"github.com/pelletier/go-toml"
	"io"
)

type Config struct {
	Main        Main                     `toml:"main"`
	Credentials map[string][]*Credential `toml:"credentials"`
}

type Main struct {
	Verbose bool `toml:"verbose"`
}

type Credential struct {
	Name string   `toml:"name"`
	From string   `toml:"from"`
	Tags []string `toml:"tags"`

	// full representation of the underlying toml structure for
	// configuring credential plugins
	tree *toml.Tree
}

func (c *Credential) Configure(x interface{}) error {
	return c.tree.Unmarshal(x)
}

func FromTree(tree *toml.Tree) (*Config, error) {
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
