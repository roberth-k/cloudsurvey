package core

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"io"
	"log"
	"sync"
)

func NewRunner(ctx context.Context, conf *config.Config) (*Runner, error) {
	var runner Runner

	for pluginName, pluginConfs := range conf.Credentials {
		for _, pluginConf := range pluginConfs {
			if err := runner.loadCredentialPlugin(ctx, pluginName, pluginConf); err != nil {
				return nil, err
			}
		}
	}

	for pluginName, pluginConfs := range conf.Sources {
		for _, pluginConf := range pluginConfs {
			if err := runner.loadSourcePlugin(ctx, pluginName, pluginConf); err != nil {
				return nil, err
			}
		}
	}

	return &runner, nil
}

type Runner struct {
	Sessions []*SessionInstance
	Sources  []*SourceInstance
}

type SessionInstance struct {
	Name       string
	Scopes     []string
	MetricTags map[string]string
	Session    registry.Credential
}

type SourceInstance struct {
	MetricTags map[string]string
	Plugin     registry.Source
}

var (
	newline = []byte{'\n'}
)

// Run configures all plugins and runs the Sources. Any metrics gathered are
// sent to the writer in InfluxDB Wire Protocol format.
func (runner *Runner) Run(ctx context.Context, w io.Writer) error {
	ch := make(chan metric.Datum, 100)
	wg := sync.WaitGroup{}

	for _, source := range runner.Sources {
		source := source
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := source.Plugin.Source(ctx, metric.ChannelCollector(ch))
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

		if _, err := w.Write([]byte(wire)); err != nil {
			log.Fatal(err)
		}

		if _, err := w.Write(newline); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (runner *Runner) getSessionByName(name string) (*SessionInstance, error) {
	for _, session := range runner.Sessions {
		if session.Name == "" {
			continue
		}

		if session.Name == name {
			return session, nil
		}
	}

	return nil, errors.Errorf("session not found: %s", name)
}

func (runner *Runner) getSessionsByScope(scope string) ([]*SessionInstance, error) {
	sessions := map[*SessionInstance]struct{}{}
	var result []*SessionInstance

	for _, session := range runner.Sessions {
		for _, sessionScope := range session.Scopes {
			if sessionScope == scope {
				if _, ok := sessions[session]; !ok {
					sessions[session] = struct{}{}
					result = append(result, session)
				}
			}
		}
	}

	return result, nil
}

func (runner *Runner) loadCredentialPlugin(ctx context.Context, name string, conf *config.Credential) error {
	init, err := registry.GetCredentials(name)
	if err != nil {
		return err
	}

	var cred registry.Credential
	if conf.From != "" {
		sess, err := runner.getSessionByName(conf.From)
		if err != nil {
			return err
		}

		cred = sess.Session
	}

	it := init(cred)
	err = conf.Configure(it)
	if err != nil {
		return err
	}

	if initializer, ok := it.(registry.Initializer); ok {
		if err := initializer.Init(); err != nil {
			return err
		}
	}

	session, err := it.Credentials(ctx)
	if err != nil {
		return err
	}

	runner.Sessions = append(runner.Sessions, &SessionInstance{
		Name:       conf.Name,
		MetricTags: util.MergeStringMaps(conf.MetricTags),
		Session:    session,
	})

	return nil
}

func (runner *Runner) loadSourcePlugin(ctx context.Context, name string, conf *config.Source) error {
	init, err := registry.GetSource(name)
	if err != nil {
		return err
	}

	for _, scope := range conf.Scopes {
		sessions, err := runner.getSessionsByScope(scope)
		if err != nil {
			return err
		}

		for _, session := range sessions {
			it := init(session.Session)
			err = conf.Configure(it)
			if err != nil {
				return err
			}

			if initializer, ok := it.(registry.Initializer); ok {
				if err := initializer.Init(); err != nil {
					return err
				}
			}

			runner.Sources = append(runner.Sources, &SourceInstance{
				MetricTags: util.MergeStringMaps(session.MetricTags, conf.MetricTags),
				Plugin:     it,
			})
		}
	}

	return nil
}
