package codebuild

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codebuild/codebuildiface"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"golang.org/x/sync/errgroup"
	"time"
)

const (
	BuildsPluginName       = "aws_codebuild_builds"
	BuildsPluginMetricName = "aws_codebuild_build"
)

func init() {
	registry.AddSource(
		BuildsPluginName,
		func(sess registry.Session) registry.Source {
			return &Builds{
				api: codebuild.New(sess.(*session.Session)),
			}
		})
}

type Builds struct {
	Since time.Duration `toml:"since"`

	api codebuildiface.CodeBuildAPI
}

func (plugin *Builds) Description() string {
	return "produces codebuild build stats"
}

func (plugin *Builds) DefaultConfig() string {
	return `
[[sources.aws_codebuild_builds]]
scopes = ["aws_regional"]
since = "1h"`
}

func (plugin *Builds) Source(c context.Context, collector metric.Collector) error {
	now := time.Now()
	eg, c := errgroup.WithContext(c)
	c = util.ContextWithNowTime(c, now)
	ch := make(chan *codebuild.Build, 10)

	eg.Go(func() error {
		defer close(ch)
		return plugin.listBuilds(c, now.Add(-plugin.Since), ch)
	})

	eg.Go(func() error {
		for build := range ch {
			datum, err := plugin.buildStats(c, build)
			if err != nil {
				return err
			}

			collector.Record(datum)
		}

		return nil
	})

	return eg.Wait()
}

func (plugin *Builds) listBuilds(c context.Context, since time.Time, ch chan<- *codebuild.Build) error {
	listBuildsInput := codebuild.ListBuildsInput{
		SortOrder: aws.String("DESCENDING"),
	}

	for {
		listBuilds, err := plugin.api.ListBuildsWithContext(c, &listBuildsInput)
		if err != nil {
			return err
		}
		listBuildsInput.NextToken = listBuilds.NextToken

		getBuildsInput := codebuild.BatchGetBuildsInput{Ids: listBuilds.Ids}
		getBuilds, err := plugin.api.BatchGetBuildsWithContext(c, &getBuildsInput)
		if err != nil {
			return err
		}

		for _, build := range getBuilds.Builds {
			if build.EndTime != nil {
				if build.EndTime.Before(since) {
					goto done
				}

				if !*build.BuildComplete {
					// TODO: Handle non-terminal builds as a separate metric.
					continue
				}

				ch <- build
			}
		}
	}

done:
	return nil
}

func (plugin *Builds) buildStats(c context.Context, build *codebuild.Build) (metric.Datum, error) {
	now := util.ContextNowTime(c)
	d := metric.Datum{
		Time:   now,
		Name:   BuildsPluginMetricName,
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	d.Tags["project_name"] = *build.ProjectName
	d.Tags["status"] = *build.BuildStatus
	d.Fields["duration"] = build.EndTime.Sub(*build.StartTime)

	return d, nil
}
