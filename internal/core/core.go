package core

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tetratom/cloudsurvey/internal/config"
	"github.com/tetratom/cloudsurvey/internal/metric"
	"github.com/tetratom/cloudsurvey/internal/registry"
	"log"
	"sync"
)

type core struct {
	credentials []registry.Credentials
	sources     []registry.Source
}

// Run configures all plugins and runs the sources. Any metrics gathered are
// sent to the writer in InfluxDB Wire Protocol format.
func Run(ctx context.Context, w StringLineWriter, conf *config.Config) error {
	var c core
	if err := c.init(conf); err != nil {
		return errors.Wrap(err, "init core")
	}

	ch := metric.ChannelCollector(make(chan metric.Datum, 100))
	wg := sync.WaitGroup{}

	for _, source := range c.sources {
		source := source
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := source.Source(ctx, ch)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for m := range ch {
		wire, err := m.ToInfluxDBLineProtocol()
		if err != nil {
			log.Fatal(err)
		}

		_, err = w.WriteStringLine(wire)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (c *core) init(conf *config.Config) error {
	for pluginName, pluginConfs := range conf.Credentials {
		for _, pluginConf := range pluginConfs {
			if err := c.loadCredentialPlugin(pluginName, pluginConf); err != nil {
				return err
			}
		}
	}

	// TODO: initialize sources

	return nil
}

func (c *core) loadCredentialPlugin(name string, conf *config.Credential) error {
	init, err := registry.GetCredentials(name)
	if err != nil {
		return err
	}

	if conf.From != "" {
		return errors.New("not implemented")
	}

	it := init(nil)
	err = conf.Configure(it)
	if err != nil {
		return errors.Wrap(err, "configure")
	}

	if initializer, ok := it.(registry.Initializer); ok {
		if err := initializer.Init(); err != nil {
			return errors.Wrap(err, "initializer")
		}
	}

	c.credentials = append(c.credentials, it)
	return nil
}