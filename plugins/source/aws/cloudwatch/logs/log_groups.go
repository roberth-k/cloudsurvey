package logs

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"log"
	"strconv"
	"time"
)

const (
	LogGroupsPluginName       = "aws_cloudwatch_log_groups"
	LogGroupsPluginMetricName = "aws_cloudwatch_log_group"
)

func init() {
	registry.AddSource(
		LogGroupsPluginName,
		func(sess registry.Session) registry.Source {
			return &LogGroups{
				api: cloudwatchlogs.New(sess.(*session.Session)),
			}
		})
}

type LogGroups struct {
	OmitRetentionTag bool `toml:"omit_retention_tag"`

	api cloudwatchlogsiface.CloudWatchLogsAPI
}

func (plugin *LogGroups) Description() string {
	return "collect stats about cloudwatch log groups"
}

func (plugin *LogGroups) DefaultConfig() string {
	return `[[sources.aws_cloudwatch_log_groups]]
scopes = ["aws_regional"]
omit_retention_tag = false`
}

func (plugin *LogGroups) Source(c context.Context, collector metric.Collector) error {
	input := cloudwatchlogs.DescribeLogGroupsInput{}
	return plugin.api.DescribeLogGroupsPagesWithContext(
		c, &input, func(output *cloudwatchlogs.DescribeLogGroupsOutput, last bool) bool {
			for _, group := range output.LogGroups {
				d, err := plugin.logGroupStats(c, group)
				if err != nil {
					log.Fatal(err) // TODO: XXX
				}
				collector.Record(d)
			}

			return true
		})
}

func (plugin *LogGroups) logGroupStats(
	c context.Context,
	group *cloudwatchlogs.LogGroup,
) (metric.Datum, error) {
	now := util.ContextNowTime(c)

	d := metric.Datum{
		Time:   now,
		Name:   LogGroupsPluginMetricName,
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	d.Tags["name"] = *group.LogGroupName
	d.Fields["age"] = now.Sub(time.Unix(*group.CreationTime/1000, 0))

	hasRetention := group.RetentionInDays != nil && *group.RetentionInDays != 0

	if hasRetention {
		d.Fields["retention_in_days"] = *group.RetentionInDays
	}

	if !plugin.OmitRetentionTag {
		if hasRetention {
			d.Tags["retention"] = strconv.FormatInt(*group.RetentionInDays, 10) + "d"
		} else {
			d.Tags["retention"] = "infinite"
		}
	}

	d.Fields["metric_filter_count"] = 0
	if group.MetricFilterCount != nil {
		d.Fields["metric_filter_count"] = *group.MetricFilterCount
	}

	d.Fields["stored_bytes"] = 0
	if group.StoredBytes != nil {
		d.Fields["stored_bytes"] = *group.StoredBytes
	}

	return d, nil
}
