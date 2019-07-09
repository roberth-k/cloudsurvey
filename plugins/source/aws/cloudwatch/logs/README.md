aws cloudwatch logs plugins
===========================

# aws_cloudwatch_log_groups

#### configuration

- `omit_retention_tag` (bool): when true, do not output the `retention` tag

#### access control

The following IAM actions are required:

- `logs:DescribeLogGroups` (resource: `*`)

#### output

Produce one datum for each CloudWatch log group found in the given session.

**name:** `aws_cloudwatch_log_group`
**tags:**

- `name`: the log group name
- `retention`: the log group retention (e.g. `7d`), or `infinite`

**fields:**

- `age` (duration): the age of the log group
- `retention_in_days` (int, optional): the log group's retention, if any
- `metric_filter_count` (int): the number of metric filters associated with the log group
- `stored_bytes` (int): the total storage required by the log group
 