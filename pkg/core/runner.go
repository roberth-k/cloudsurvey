package core

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	_ "github.com/tetratom/cloudsurvey/plugins"
	"golang.org/x/sync/errgroup"
	"log"
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
	Session    registry.Session
}

type SourceInstance struct {
	Name       string
	MetricTags map[string]string
	Plugin     registry.Source
}

// Run configures all plugins and runs the Sources. Metrics are sent to the
// given channel. The channel is _not_ closed by Run.
func (runner *Runner) Run(ctx context.Context, ch chan<- metric.Datum) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, source := range runner.Sources {
		source := source
		eg.Go(func() error {
			collector := metric.MetricTagOverrideCollector{
				Inner:      metric.ChannelCollector(ch),
				MetricTags: source.MetricTags,
			}

			if err := source.Plugin.Source(ctx, collector); err != nil {
				log.Printf("error: source %s: %+v", source.Name, err)
			}

			return nil
		})
	}

	return eg.Wait()
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

	var cred registry.Session
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
		Scopes:     conf.Scopes,
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
				Name:       name,
				MetricTags: util.MergeStringMaps(session.MetricTags, conf.MetricTags),
				Plugin:     it,
			})
		}
	}

	return nil
}
