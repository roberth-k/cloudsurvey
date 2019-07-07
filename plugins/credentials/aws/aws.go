package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tetratom/cloudsurvey/pkg/registry"
)

func init() {
	registry.AddCredentials(
		"aws",
		func(cred registry.Session) registry.Credentials {
			x := AWS{}
			if cred != nil {
				x.from = cred.(*session.Session)
			}
			return &x
		})
}

type AWS struct {
	Profile         string `toml:"profile"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Token           string `toml:"token"`
	RoleARN         string `toml:"role_arn"`
	ExternalID      string `toml:"external_id"`
	Region          string `toml:"region"`
	RoleSessionName string `toml:"role_session_name"`
	SharedConfig    *bool  `toml:"shared_config"`

	from    *session.Session
	session *session.Session
}

func (plugin *AWS) Init() error {
	opts := session.Options{}

	if plugin.SharedConfig == nil || *plugin.SharedConfig {
		opts.SharedConfigState = session.SharedConfigEnable
	}

	if plugin.Profile != "" {
		opts.Profile = plugin.Profile
	} else if plugin.RoleARN != "" {
		sess := plugin.from
		if sess == nil {
			sess = session.Must(session.NewSessionWithOptions(opts))
		}

		opts.Config.Credentials = stscreds.NewCredentials(
			sess, plugin.RoleARN, func(provider *stscreds.AssumeRoleProvider) {
				if plugin.ExternalID != "" {
					provider.ExternalID = &plugin.ExternalID
				}

				if plugin.RoleSessionName != "" {
					provider.RoleSessionName = plugin.RoleSessionName
				}
			})
	} else if plugin.AccessKeyID != "" || plugin.SecretAccessKey != "" || plugin.Token != "" {
		opts.Config.Credentials = credentials.NewStaticCredentials(
			plugin.AccessKeyID,
			plugin.SecretAccessKey,
			plugin.Token)
	}

	if plugin.Region != "" {
		opts.Config.Region = &plugin.Region
	}

	plugin.session = session.Must(session.NewSessionWithOptions(opts))

	return nil
}

func (*AWS) Description() string {
	return "provides pointers to aws-sdk-go sessions"
}

func (*AWS) DefaultConfig() string {
	return `
[[credentials.aws]]
shared_config = true
scopes = ["aws_regional", "aws_global"]`
}

func (plugin *AWS) Credentials(context.Context) (registry.Session, error) {
	return plugin.session, nil
}
