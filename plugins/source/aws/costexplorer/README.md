aws costexplorer plugins
========================

# aws_ce_daily

#### configuration

- `metrics` ([]string): default is `["AmortizedCost", "BlendedCost", "UnblendedCost"]`.

#### access control

The following IAM actions are required:

- `ce:Get*`

#### output

Produce one datum for each service group. **The data will be backdated to now - 24h.**

**name:** `aws_ce_daily_cost_per_service`
**tags:**

- `service`: the name of a service, as returned by Cost Explorer

**fields:**

See [aws documentation](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/ce-advanced.html) for more information about cost types. The plugin will output one field per metric configured. The default set is:

- `AmortizedCost` (money)
- `BlendedCost` (money)
- `UnblendedCost` (money)
