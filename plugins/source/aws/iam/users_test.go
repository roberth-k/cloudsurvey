package iam

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/stretchr/testify/require"
	"github.com/tetratom/cloudsurvey/internal/util"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	"testing"
	"time"
)

type mockUsersIAMAPI struct {
	iamiface.IAMAPI
	accessKeyMetadata []*iam.AccessKeyMetadata
	accessKeyLastUsed map[string]*iam.GetAccessKeyLastUsedOutput
}

func (api *mockUsersIAMAPI) ListAccessKeysWithContext(
	ctx aws.Context,
	input *iam.ListAccessKeysInput,
	options ...request.Option,
) (*iam.ListAccessKeysOutput, error) {
	return &iam.ListAccessKeysOutput{AccessKeyMetadata: api.accessKeyMetadata}, nil
}

func (api *mockUsersIAMAPI) GetAccessKeyLastUsedWithContext(
	ctx aws.Context,
	input *iam.GetAccessKeyLastUsedInput,
	options ...request.Option,
) (*iam.GetAccessKeyLastUsedOutput, error) {
	return api.accessKeyLastUsed[*input.AccessKeyId], nil
}

func TestUsers_userStats(t *testing.T) {
	tz := time.Date(2019, 1, 2, 3, 4, 0, 0, time.UTC)
	now := tz.Add(99 * time.Nanosecond)
	c := util.ContextWithNowTime(context.Background(), now)

	t.Run("with login and access keys", func(t *testing.T) {
		api := mockUsersIAMAPI{
			accessKeyMetadata: []*iam.AccessKeyMetadata{
				{
					CreateDate:  aws.Time(tz.Add(-20 * time.Nanosecond)),
					AccessKeyId: aws.String("A"),
					Status:      aws.String("inactive"),
				},
				{
					CreateDate:  aws.Time(tz.Add(20 * time.Nanosecond)),
					AccessKeyId: aws.String("B"),
					Status:      aws.String("active"),
				},
			},
			accessKeyLastUsed: map[string]*iam.GetAccessKeyLastUsedOutput{
				"A": {
					AccessKeyLastUsed: &iam.AccessKeyLastUsed{
						LastUsedDate: aws.Time(tz.Add(40 * time.Nanosecond)),
					},
				},
				"B": {
					AccessKeyLastUsed: &iam.AccessKeyLastUsed{
						LastUsedDate: aws.Time(tz.Add(50 * time.Nanosecond)),
					},
				},
			},
		}

		plugin := Users{
			api: &api,
		}

		user := iam.User{
			CreateDate:       &tz,
			UserName:         aws.String("bob"),
			Path:             aws.String("/friends/"),
			PasswordLastUsed: aws.Time(tz.Add(30)),
			Tags:             []*iam.Tag{{Key: aws.String("MyTag"), Value: aws.String("MyValue")}},
		}

		d, err := plugin.userStats(c, &user)
		require.NoError(t, err)
		require.Equal(t, metric.Datum{
			Name: "aws_iam_user",
			Time: now,
			Tags: map[string]string{
				"user_name": "bob",
				"user_path": "/friends/",
				"tag_MyTag": "MyValue",
			},
			Fields: map[string]interface{}{
				"age":                       99 * time.Nanosecond,
				"active_key_count":          1,
				"oldest_key_age":            119 * time.Nanosecond,
				"since_last_activity":       49 * time.Nanosecond,
				"since_last_login_activity": 69 * time.Nanosecond,
				"since_last_key_activity":   49 * time.Nanosecond,
			},
		}, d)
	})
}
