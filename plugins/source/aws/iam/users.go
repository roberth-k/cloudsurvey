package iam

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"github.com/tetratom/cloudsurvey/pkg/registry"
	"golang.org/x/sync/errgroup"
	"time"
)

const (
	UsersPluginName        = "aws_iam_users"
	UsersPluginMetricName  = "aws_iam_user"
	UsersPluginConcurrency = 2
)

func init() {
	registry.AddSource(
		UsersPluginName,
		func(cred registry.Session) registry.Source {
			return &Users{
				api: iam.New(cred.(*session.Session)),
			}
		})
}

type Users struct {
	OmitUserTags bool `toml:"omit_user_tags"`
	OmitUserPath bool `toml:"omit_user_path"`

	api iamiface.IAMAPI
}

func (plugin *Users) Description() string {
	return "produces stats about iam users"
}

func (plugin *Users) Source(c context.Context, collector metric.Collector) error {
	eg, c := errgroup.WithContext(c)
	c = util.ContextWithNowTime(c, time.Now())
	ch := make(chan *iam.User, 10)

	eg.Go(func() error {
		defer close(ch)
		return plugin.listUsers(c, ch)
	})

	for i := 0; i < UsersPluginConcurrency; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-c.Done():
					return c.Err()
				case user, more := <-ch:
					if !more {
						return nil
					}

					d, err := plugin.userStats(c, user)
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

func (plugin *Users) userStats(c context.Context, user *iam.User) (metric.Datum, error) {
	keys, err := plugin.listAccessKeys(c, *user.UserName)
	if err != nil {
		return metric.Datum{}, err
	}

	d := metric.Datum{}
	d.Time = util.ContextNowTime(c)
	d.Name = UsersPluginMetricName
	d.Tags = map[string]string{
		"user_name": *user.UserName,
	}
	d.Fields = map[string]interface{}{
		"age": d.Time.Sub(*user.CreateDate),
	}

	if !plugin.OmitUserPath {
		d.Tags["user_path"] = *user.Path
	}

	var activeKeyCount int
	var lastActivity, lastKeyActivity, lastLoginActivity time.Time
	var oldestKeyCreatedDate time.Time

	if t := user.PasswordLastUsed; t != nil {
		lastLoginActivity = *t
		lastActivity = util.MaxTime(lastActivity, lastLoginActivity)
	}

	for _, key := range keys {
		if *key.Status == "active" {
			activeKeyCount++
		}

		if oldestKeyCreatedDate.IsZero() || key.CreateDate.Before(oldestKeyCreatedDate) {
			oldestKeyCreatedDate = *key.CreateDate
		}

		lastUsedTime, err := plugin.getAccessKeyLastUsed(c, *key.AccessKeyId)
		if err != nil {
			return metric.Datum{}, err
		}

		lastKeyActivity = util.MaxTime(lastKeyActivity, *lastUsedTime)
		lastActivity = util.MaxTime(lastActivity, lastKeyActivity)
	}

	if !oldestKeyCreatedDate.IsZero() {
		d.Fields["oldest_key_age"] = d.Time.Sub(oldestKeyCreatedDate)
	}

	d.Fields["active_key_count"] = activeKeyCount

	if !lastActivity.IsZero() {
		d.Fields["since_last_activity"] = d.Time.Sub(lastActivity)

		if !lastLoginActivity.IsZero() {
			d.Fields["since_last_login_activity"] = d.Time.Sub(lastLoginActivity)
		}

		if !lastKeyActivity.IsZero() {
			d.Fields["since_last_key_activity"] = d.Time.Sub(lastKeyActivity)
		}
	}

	if !plugin.OmitUserTags {
		for _, tag := range user.Tags {
			d.Tags["tag_"+*tag.Key] = *tag.Value
		}
	}

	return d, nil
}

func (plugin *Users) listUsers(c context.Context, ch chan<- *iam.User) error {
	input := iam.ListUsersInput{}
	return plugin.api.ListUsersPagesWithContext(
		c, &input, func(out *iam.ListUsersOutput, last bool) bool {
			for _, user := range out.Users {
				select {
				case <-c.Done():
					return false
				case ch <- user:
				}
			}
			return true
		})
}

func (plugin *Users) listAccessKeys(c context.Context, username string) ([]*iam.AccessKeyMetadata, error) {
	input := iam.ListAccessKeysInput{UserName: &username}
	out, err := plugin.api.ListAccessKeysWithContext(c, &input)
	if err != nil {
		return nil, err
	}
	return out.AccessKeyMetadata, nil
}

func (plugin *Users) getAccessKeyLastUsed(c context.Context, id string) (*time.Time, error) {
	input := iam.GetAccessKeyLastUsedInput{AccessKeyId: &id}
	out, err := plugin.api.GetAccessKeyLastUsedWithContext(c, &input)
	if err != nil {
		return nil, err
	}

	return out.AccessKeyLastUsed.LastUsedDate, nil
}
