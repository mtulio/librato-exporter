# librato-exporter

Librato metrics Exporter for Prometheus


The Librato Exporter Exporter [consumes data from Librato API](https://www.librato.com/docs/api/) using the [Go lib](https://github.com/rcrowley/go-librato) exposing it for Prometheus.

For each Librato Metric, the Prometheus metric will be created:

`librato_<AccountAlias>_timestamp`: Current timetamp of the metric. Label `name` is added to each metric
`libratoe_<AccountAlias>_value`: Current value of the metric. Label `name` is added to each metric

## BUILD

`make build`

The binary will be created on `./bin` dir.

## ARGUMENTS

### REQUIRED

`-librato.email` : Librato's Account Email
`-librato.token` : Librato's Account Token
`-metrics.filter` : Librato's metrics to be gathered sepparated by comma

### OPTIONAL

* Exporter API context

`-web.listen-address` : HTTP port for the exporter API. Default: `9800`
`-web.telemetry-path` : HTTP path for the exporter API. Default: `/metrics`

* Librato context

`-librato.account` : the Alias name of Librato's Account. This Alias will be used to a Prometheus metric subsystem. In the example above, the value of `<AccountAlias>` will be overrided. Default: `app`
`-librato.interval` : the interval in seconds that the exporter will retrieve metrics from Librato API. Default: `300`
`-librato.offset` : the offset in seconds that the exporter will calculate the `time_start` timestamp. Default: `120`

* Metrics context

`-metrics.resolution` : Librato's metrics resolution. This is global value. Default: `30`

## USAGE

Show Librato metrics

```bash
./bin/librato-exporter -librato.email=my@email.com -librato.token=myToken \
    -metrics.filter=myapp.res_time.global,myapp.pageview.success,myapp.view.success \
    -librato.interval=60 -metrics.resolution=30 -metrics.offset=120 \
    -librato.account=myapp
```

> Sample output:

```log
# HELP librato_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which librato_exporter was built.
# TYPE librato_exporter_build_info gauge
librato_exporter_build_info{branch="",goversion="go1.11.2",revision="8c69428",version="v0.1.0"} 1
# HELP librato_myapp_timestamp Librato metric last timestamp
# TYPE librato_myapp_timestamp counter
librato_myapp_timestamp{name="myapp.view.success"} 1.54197078e+09
librato_myapp_timestamp{name="myapp.pageview.success"} 1.54197078e+09
librato_myapp_timestamp{name="myapp.res_time.global"} 1.54197078e+09
# HELP librato_myapp_up Librato metric status for the last collect
# TYPE librato_myapp_up gauge
librato_myapp_up{name="myap.view.success"} 1
librato_myapp_up{name="myapp.pageview.success"} 1
librato_myapp_up{name="myapp.res_time.global"} 1
# HELP librato_myapp_value Librato metric value
# TYPE librato_myapp_value gauge
librato_myapp_value{name="myapp.view.success"} 187
librato_myapp_value{name="myapp.pageview.success"} 790
librato_myapp_value{name="myapp.res_time.global"} 89.404956
# HELP librato_scrape_collector_duration_seconds node_exporter: Duration of a collector scrape.
# TYPE librato_scrape_collector_duration_seconds gauge
librato_scrape_collector_duration_seconds{collector="myapp"} 0.000108181
# HELP librato_scrape_collector_success master_exporter: Whether a collector succeeded.
# TYPE librato_scrape_collector_success gauge
librato_scrape_collector_success{collector="myapp"} 1
```

## USAGE DOCKER

Show librato metrics running in docker

```
docker run -p 9800:9800 -id mtulio/librato-exporter:latest \
    -librato.email=my@email.com -librato.token=myToken -metrics.filter=myapp.res_time.global,myapp.pageview.success,myapp.view.success \
    -librato.interval=60 -metrics.resolution=30 -metrics.offset=120 \
    -librato.account=myapp
```

## CONTRIBUTOR

* Fork me
* Open an PR with enhancements, bugfixes, etc
* Open an issue
* Write docs

[...]

You are welcome. =)
