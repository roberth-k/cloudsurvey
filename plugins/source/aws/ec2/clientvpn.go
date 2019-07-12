package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"golang.org/x/sync/errgroup"
	"log"
	"strconv"
	"time"
)

const (
	ClientVpnPluginName                 = "aws_ec2_clientvpn"
	ClientVpnPluginConnectionMetricName = "aws_ec2_clientvpn_connection"
	clientVpnTimeFormat                 = "2006-01-02 15:04:05"
)

func init() {
	registry.AddSource(
		ClientVpnPluginName,
		func(sess registry.Session) registry.Source {
			return &ClientVPN{
				api: ec2.New(sess.(*session.Session)),
			}
		})
}

type ClientVPN struct {
	api ec2iface.EC2API
}

func (plugin *ClientVPN) Description() string {
	return ""
}

func (plugin *ClientVPN) DefaultConfig() string {
	return `[[sources.aws_ec2_clientvpn]]
scopes = ["aws_regional"]`
}

func (plugin *ClientVPN) Source(c context.Context, collector metric.Collector) error {
	eg, c := errgroup.WithContext(c)
	ch := make(chan *ec2.ClientVpnEndpoint, 2)

	eg.Go(func() error {
		defer close(ch)
		return plugin.getEndpointIds(c, ch)
	})

	for i := 0; i < 2; i++ {
		eg.Go(func() error {
			for endpoint := range ch {
				if err := plugin.sendEndpointStats(c, collector, endpoint); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return eg.Wait()
}

func (plugin *ClientVPN) getEndpointIds(c context.Context, ch chan<- *ec2.ClientVpnEndpoint) error {
	input := ec2.DescribeClientVpnEndpointsInput{}
	return plugin.api.DescribeClientVpnEndpointsPagesWithContext(
		c, &input, func(page *ec2.DescribeClientVpnEndpointsOutput, last bool) bool {
			for _, endpoint := range page.ClientVpnEndpoints {
				select {
				case <-c.Done():
					return false
				case ch <- endpoint:
				}
			}

			return true
		})
}

func (plugin *ClientVPN) sendEndpointStats(
	c context.Context,
	collector metric.Collector,
	endpoint *ec2.ClientVpnEndpoint,
) error {
	input := ec2.DescribeClientVpnConnectionsInput{
		ClientVpnEndpointId: endpoint.ClientVpnEndpointId,
	}
	return plugin.api.DescribeClientVpnConnectionsPagesWithContext(
		c, &input, func(page *ec2.DescribeClientVpnConnectionsOutput, last bool) bool {
			for _, connection := range page.Connections {
				datum, err := plugin.connectionStats(c, endpoint, connection)
				if err != nil {
					// TODO: Capturing this error?
					log.Printf("error: %+v", err)
					return false
				}

				collector.Record(datum)
			}

			return true
		})
}

func (plugin *ClientVPN) connectionStats(
	c context.Context,
	endpoint *ec2.ClientVpnEndpoint,
	connection *ec2.ClientVpnConnection,
) (metric.Datum, error) {
	now := util.ContextNowTime(c)
	d := metric.Datum{
		Time:   now,
		Name:   ClientVpnPluginConnectionMetricName,
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	d.Tags["endpoint_id"] = *connection.ClientVpnEndpointId

	if connection.CommonName != nil {
		d.Tags["common_name"] = *connection.CommonName
	}

	d.Tags["status"] = *connection.Status.Code

	if connection.Username != nil {
		d.Tags["username"] = *connection.Username
	}

	if t, err := time.Parse(clientVpnTimeFormat, *connection.ConnectionEstablishedTime); err == nil {
		d.Fields["age"] = now.Sub(t)
	}

	i64Fields := []struct {
		value *string
		name  string
	}{
		{connection.EgressBytes, "egress_bytes"},
		{connection.EgressPackets, "egress_packets"},
		{connection.IngressBytes, "ingress_bytes"},
		{connection.IngressPackets, "ingress_packets"},
	}

	for _, field := range i64Fields {
		if field.value == nil {
			continue
		}

		if b, err := strconv.ParseInt(*field.value, 10, 64); err == nil {
			d.Fields[field.name] = b
		}
	}

	return d, nil
}
