## Spot instance termination collector

Prometheus [collectors](https://prometheus.io/docs/instrumenting/writing_exporters/#collectors) are part of exporters - this is a collector to scrape for AWS spot price termination notice on the instance for Hollowtrees.

### Spot instance lifecycle

* User submits a bid to run a desired number of EC2 instances of a particular type. The bid includes the price that the user is willing to pay to use the instance for an hour.
* When your bid price exceeds the current Spot price (which varies based on supply and demand), the instances are run.
* When the current Spot price rises above the bid price, the Spot instance is reclaimed by AWS so that it can be given to another customer.

### Spot instance termination notice

The Termination Notice is accessible to code running on the instance via the instance’s metadata at `http://169.254.169.254/latest/meta-data/spot/termination-time`. This field becomes available when the instance has been marked for termination and will contain the time when a shutdown signal will be sent to the instance’s operating system. At that time, the Spot Instance Request’s bid status will be set to `marked-for-termination.` The bid status is accessible via the `DescribeSpotInstanceRequests` API for use by programs that manage Spot bids and instances.
