## Spot instance termination collector

Prometheus [collectors](https://prometheus.io/docs/instrumenting/writing_exporters/#collectors) are part of exporters - this is a collector to scrape for AWS spot price termination notice on the instance for [Hollowtrees](https://github.com/banzaicloud/hollowtrees).

### Spot instance lifecycle

* User submits a bid to run a desired number of EC2 instances of a particular type. The bid includes the price that the user is willing to pay to use the instance for an hour.
* When the bid price exceeds the current spot price (which varies based on supply and demand), the instances are run.
* When the current spot price rises above the bid price, the spot instance is reclaimed by AWS so that it can be given to another customer.

### Spot instance termination notice

The Termination Notice is accessible to code running on the instance via the instance’s metadata at `http://169.254.169.254/latest/meta-data/spot/termination-time`. This field becomes available when the instance has been marked for termination and will contain the time when a shutdown signal will be sent to the instance’s operating system. 
At that time, the Spot Instance Request’s bid status will be set to `marked-for-termination.`  
The bid status is accessible via the `DescribeSpotInstanceRequests` API for use by programs that manage Spot bids and instances.

### Quick start

The project uses the [promu](https://github.com/prometheus/promu) Prometheus utility tool. To build the exporter `promu` needs to be installed. To install promu and build the exporter:

```
go get github.com/prometheus/promu
promu build
```

The following options can be configured when starting the exporter:

```
./spot_expiry_exporter --help
Usage of ./spot_expiry_exporter:
  -bind-addr string
        bind address for the metrics server (default ":9189")
  -log-level string
        log level (default "info")
  -metadata-endpoint string
        metadata endpoint to query (default "http://169.254.169.254/latest/meta-data/")
  -metrics-path string
        path to metrics endpoint (default "/metrics")

```

### Test locally

The AWS instance metadata is available at `http://169.254.169.254/latest/meta-data/`. By default this is the endpoint that is being queried by the exporter but it is quite hard to reproduce a termination notice on an AWS instance for testing, so the meta-data endpoint can be changed in the configuration.
There is a test server in the `utils` directory that can be used to mock the behavior of the metadata endpoint. It listens on port 9092and provides dummy responses for `/instance-id` and `spot/instance-action`. It can be started with:
```
go run util/test_server.go
```
The exporter can be started with this configuration to query this endpoint locally:
```
./spot_expiry_exporter --metadata-endpoint http://localhost:9092/latest/meta-data/ --log-level debug
```

### Metrics

```
# HELP aws_instance_metadata_service_available Metadata service available
# TYPE aws_instance_metadata_service_available gauge
aws_instance_metadata_service_available 1
# HELP aws_instance_termination_imminent Instance is about to be terminated
# TYPE aws_instance_termination_imminent gauge
aws_instance_termination_imminent{instance_action="stop",instance_id="i-0d2aab13057917887"} 1
# HELP aws_instance_termination_in Instance will be terminated in
# TYPE aws_instance_termination_in gauge
aws_instance_termination_in 119.354569
```
