cloudsurvey [![CircleCI](https://circleci.com/gh/tetratom/cloudsurvey.svg?style=svg)](https://circleci.com/gh/tetratom/cloudsurvey)
===========

_cloudsurvey is a (multi-)cloud stats collector for telegraf._

It collects general information about your cloud infrastructure, such as the number of users or the age of your instances. It is currently designed to be compiled as an executable to be invoked by the telegraf exec plugin, but should eventually be available as an external plugin once support is available. You will find that the design of cloudsurvey takes a number of inspirations from telegraf directly.

Metrics are written to standard output according to the InfluxDB Wire Protocol.

## supported plugins

#### credentials

- [aws](./plugins/credentials/aws)

#### source

- [aws_ce_daily](./plugins/source/aws/costexplorer#aws_ce_daily)
- [aws_cloudwatch_log_groups](./plugins/source/aws/cloudwatch/logs#aws_cloudwatch_log_groups)
- [aws_codebuild_builds](./plugins/source/aws/codebuild#aws_codebuild_builds)
- [aws_ec2_clientvpn](./plugins/source/aws/ec2#aws_ec2_clientvpn)
- [aws_ec2_instances](./plugins/source/aws/ec2#aws_ec2_instances)
- [aws_iam_users](./plugins/source/aws/iam#aws_iam_users)
