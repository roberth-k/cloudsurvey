aws ec2 plugins
===============

# aws_ec2_clientvpn

#### configuration

N/A

#### access control

The following IAM actions are required:

- `ec2:DescribeClientVpnConnections`
- `ec2:DescribeClientVpnEndpoints`

#### output

Produce one datum for each Client VPN connection.

**name:** `aws_ec2_clientvpn_connection`
**tags:**

- `endpoint_id`: the client endpoint id
- `common_name` (optional): the connection's common name
- `status`: the connection status
- `username` (optional): the connection's user name

**fields:**

- `age` (duration): the time since the connection was established
- `egress_bytes` (count): the total number of bytes sent from the connection
- `egress_packets` (count): the total number of packets sent from the connection
- `ingress_bytes` (count): the total number of bytes sent to the connection
- `ingress_packets` (count): the total number of packets sent from the connection

# aws_ec2_instances

#### configuration

- `ignore_image_details` (bool): when true, do not populate ami information
- `loose_instance_family` (bool): when true, fold specialized families such as r5a, r5ad, r5d into r5.

#### access control

The following IAM actions are required:

- `ec2:DescribeInstance` (resource: `*`)
- `ec2:DescribeImages`

#### output

Produce one datum for each EC2 instance found in the given session.

**name:** `aws_ec2_instance`
**tags:**

- `id`: the instance id
- `state`: the instance state (e.g. running)
- `platform`: the instance platform (e.g. linux)
- `type`: the instance type (e.g. t3.small)
- `family`: the instance family (e.g. t3)
- `lifecycle`: the instance lifecycle (e.g. spot)
- `image_id`: the ami id
- `image_name`: the ami name

**fields:**

- `age` (duration): the instance age
- `image_age` (duration): the age of the instance's ami
- `vcpus` (count): the number of vCPUs 