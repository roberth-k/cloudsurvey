package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"golang.org/x/sync/errgroup"
	"strings"
	"sync"
	"time"
)

const (
	InstancesPluginName        = "aws_ec2_instances"
	InstancesPluginMetricName  = "aws_ec2_instance"
	InstancesPluginConcurrency = 4
)

func init() {
	registry.AddSource(
		InstancesPluginName,
		func(sess registry.Session) registry.Source {
			return &Instances{
				api: ec2.New(sess.(*session.Session)),
			}
		})
}

type Instances struct {
	IgnoreImageDetails  bool `toml:"ignore_image_details"`
	LooseInstanceFamily bool `toml:"loose_instance_family"`

	api ec2iface.EC2API
}

var imagesById sync.Map // TODO: eviction

type imageCache struct {
	once  sync.Once
	image *ec2.Image
	err   error
}

func (plugin *Instances) Description() string {
	return "get stats about ec2 instances"
}

func (plugin *Instances) Source(c context.Context, collector metric.Collector) error {
	eg, c := errgroup.WithContext(c)
	c = util.ContextWithNowTime(c, time.Now())
	ch := make(chan *ec2.Instance, 10)

	eg.Go(func() error {
		defer close(ch)
		return plugin.describeInstances(c, ch)
	})

	for i := 0; i < InstancesPluginConcurrency; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-c.Done():
					return c.Err()
				case instance, more := <-ch:
					if !more {
						return nil
					}

					d, err := plugin.instanceStats(c, instance)
					if err != nil {
						return err
					}

					collector.Record(d)
				}
			}
		})
	}

	return eg.Wait()
}

func (plugin *Instances) describeInstances(c context.Context, ch chan<- *ec2.Instance) error {
	input := ec2.DescribeInstancesInput{}
	return plugin.api.DescribeInstancesPagesWithContext(
		c, &input, func(output *ec2.DescribeInstancesOutput, last bool) bool {
			for _, reservation := range output.Reservations {
				for _, instance := range reservation.Instances {
					select {
					case <-c.Done():
						return false
					case ch <- instance:
					}
				}
			}

			return true
		})
}

func (plugin *Instances) instanceStats(c context.Context, instance *ec2.Instance) (metric.Datum, error) {
	now := util.ContextNowTime(c)

	d := metric.Datum{
		Time:   util.ContextNowTime(c),
		Name:   InstancesPluginMetricName,
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	d.Tags["id"] = *instance.InstanceId
	d.Tags["state"] = *instance.State.Name
	d.Tags["platform"] = oneof("linux", instance.Platform)
	d.Tags["type"] = *instance.InstanceType
	d.Tags["family"] = instanceFamily(*instance.InstanceType, plugin.LooseInstanceFamily)
	d.Tags["lifecycle"] = oneof("normal", instance.InstanceLifecycle)
	d.Tags["image_id"] = *instance.ImageId
	d.Fields["age"] = now.Sub(*instance.LaunchTime)

	if !plugin.IgnoreImageDetails {
		image, err := plugin.describeImage(c, *instance.ImageId)
		if err == nil {
			// we'll ignore image retrieval errors, as they can become unavailable
			d.Tags["image_name"] = *image.Name

			if t, err := imageCreationDate(image.CreationDate); err == nil {
				d.Fields["image_age"] = now.Sub(t)
			}
		}
	}

	return d, nil
}

func (plugin *Instances) describeImage(c context.Context, id string) (*ec2.Image, error) {
	cache_, _ := imagesById.LoadOrStore(id, &imageCache{})
	cache := cache_.(*imageCache)

	cache.once.Do(func() {
		input := ec2.DescribeImagesInput{ImageIds: []*string{&id}}
		out, err := plugin.api.DescribeImagesWithContext(c, &input)
		if err != nil {
			cache.err = err
		} else if len(out.Images) == 0 {
			cache.err = errors.Errorf("image not found: %s", id)
		} else {
			cache.image = out.Images[0]
		}
	})

	return cache.image, cache.err
}

func oneof(either string, or *string) string {
	if or != nil && *or != "" {
		return *or
	}

	return either
}

func instanceFamily(instanceType string, loose bool) string {
	if loose {
		return instanceType[:2]
	}

	return strings.Split(instanceType, ".")[0]
}

func imageCreationDate(s *string) (time.Time, error) {
	if s == nil {
		return time.Time{}, errors.New("cannot parse nil time")
	}

	return time.Parse("2006-01-02T15:04:05.000Z", *s)
}
