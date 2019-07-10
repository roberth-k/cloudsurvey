package costexplorer

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/costexplorer/costexploreriface"
	"github.com/pkg/errors"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"strconv"
	"time"
)

const (
	DailyPluginName                     = "aws_ce_daily"
	DailyPluginCostPerServiceMetricName = "aws_ce_daily_cost_per_service"
)

var (
	dailyDefaultMetrics = []string{
		"AmortizedCost",
		"BlendedCost",
		"UnblendedCost",
	}
)

func init() {
	registry.AddSource(
		DailyPluginName,
		func(sess registry.Session) registry.Source {
			return &Daily{
				api: costexplorer.New(sess.(*session.Session)),
			}
		})
}

type Daily struct {
	Metrics []string `toml:"metrics"`
	api     costexploreriface.CostExplorerAPI
}

func (plugin *Daily) Init() error {
	if len(plugin.Metrics) == 0 {
		plugin.Metrics = dailyDefaultMetrics
	}

	return nil
}

func (plugin *Daily) Description() string {
	return "produce daily cost and usage statistics from cost explorer"
}

func (plugin *Daily) DefaultConfig() string {
	return `[[sources.aws_ce_daily]]
scopes = ["aws_global"]
metrics = ["AmortizedCost", "BlendedCost", "UnblendedCost"]`
}

func (plugin *Daily) Source(c context.Context, collector metric.Collector) error {
	now := util.ContextNowTime(c).UTC()
	since := now.Add(-48 * time.Hour)
	until := now.Add(-24 * time.Hour)

	input := costexplorer.GetCostAndUsageInput{
		Granularity: aws.String("DAILY"),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String("DIMENSION"),
				Key:  aws.String("SERVICE"),
			},
		},
		Metrics: aws.StringSlice(plugin.Metrics),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(since.Format("2006-01-02")),
			End:   aws.String(until.Format("2006-01-02")),
		},
	}

	out, err := plugin.api.GetCostAndUsageWithContext(c, &input)
	if err != nil {
		return err
	}

	for _, group := range out.ResultsByTime[0].Groups {
		datum, err := plugin.usageStats(c, group)
		if err != nil {
			return err
		}

		collector.Record(datum)
	}

	return nil
}

func (plugin *Daily) usageStats(c context.Context, group *costexplorer.Group) (metric.Datum, error) {
	now := util.ContextNowTime(c)
	d := metric.Datum{
		Name:   DailyPluginCostPerServiceMetricName,
		Time:   now.Add(-24 * time.Hour),
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	d.Tags["service"] = *group.Keys[0]

	for name, data := range group.Metrics {
		if *data.Unit != "USD" {
			return metric.Datum{}, errors.Errorf("expected USD, but got %s", *data.Unit)
		}

		amount, err := strconv.ParseFloat(*data.Amount, 64)
		if err != nil {
			return metric.Datum{}, errors.Wrapf(err, "parse amount: %s", *data.Amount)
		}

		d.Fields[name] = amount
	}

	return d, nil
}
