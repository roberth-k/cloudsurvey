cloudsurvey
===========

_cloudsurvey is a (multi-)cloud stats collector for telegraf._

It collects general information about your cloud infrastructure, such as the number of users or the age of your instances. It is currently designed to be compiled as an executable to be invoked by the telegraf exec plugin, but should eventually be available as an external plugin once support is available.

Metrics are written to standard output according to the InfluxDB Line Protocol.
