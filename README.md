# Prometheus Mesos Exporter

[![Build Status](https://travis-ci.org/mesosphere/mesos_exporter.svg?branch=master)](https://travis-ci.org/mesosphere/mesos_exporter)

Exporter for Mesos master and agent metrics.

## Using
The Mesos Exporter can either expose cluster wide metrics from a master or task
metrics from an agent.

```sh
Usage of mesos_exporter:
  -addr string
        Address to listen on (default ":9105")
  -exportedSlaveAttributes string
        Comma-separated list of slave attributes to include in the corresponding metric
  -exportedTaskLabels string
        Comma-separated list of task labels to include in the corresponding metric
  -master string
        Expose metrics from master running on this URL
  -slave string
        Expose metrics from slave running on this URL
  -timeout duration
        Master polling timeout (default 5s)
  -username string
        Username to use for HTTP or strict mode authentication
  -password string
        Password to use for HTTP or strict mode authentication
  -loginURL
        URL for strict mode authentication (default https://leader.mesos/acs/api/v1/auth/login).
  -trustedCerts string
        Comma-separated list of certificates (.pem files) trusted for requests to Mesos endpoints
  -strictMode
        Enable strict mode API authentication
  -privateKey
        Private key used for strict mode authentication. This must be provided
        when using strict mode. However, it can be read from the environment if the secret store is used. 
  -skipSSLVerify
        Disable SSL certificate verification
```
When using HTTP or strict mode authentication, the following values are read from the environment, if they are not specified at run time:
- `MESOS_EXPORTER_USERNAME`
- `MESOS_EXPORTER_PASSWORD`
- `MESOS_EXPORTER_PRIVATE_KEY`

## Docker
```sh
Usage of docker:

For Master
docker run --net="host" -e MODE=-master -e URL=http://localhost:5050 -p 9015:9015 mesos_exporter:latest

For Slave
docker run --net="host" -e MODE=-slave -e URL=http://localhost:5051 -p 9015:9015 mesos_exporter:latest
```

## Prometheus Configuration

Usually you would run one exporter with `-master` for each master and one
exporter for each slave with `-slave`. Monitoring each master individually
ensures that the cluster can be monitored even if the underlying Mesos cluster
is in a degraded state.

- Master: `mesos_exporter -master http://localhost:5050`
- Agent: `mesos_exporter -slave http://localhost:5051`

The necessary Prometheus configuration could look like this:

```
- job_name: mesos-master
  scrape_interval: 15s
  scrape_timeout: 10s
  static_configs:
  - targets:
    - master1.mesos.example.org:9105
    - master2.mesos.example.org:9105
    - master3.mesos.example.org:9105

- job_name: mesos-slave
  scrape_interval: 15s
  scrape_timeout: 10s
  static_configs:
  - targets:
    - node1.mesos.example.org:9105
    - node2.mesos.example.org:9105
    - node3.mesos.example.org:9105
```


A minimal set of alerts to ensure your cluster is operational could then be defined
as follows:

```
ALERT MesosDown
  IF (up{job=~"mesos.*"} == 0) or (irate(mesos_collector_errors_total[5m]) > 0)
  FOR 5m
  LABELS { severity="warning" }
  ANNOTATIONS {
    description="Either the exporter or the associated Mesos component is down.",
    summary="The Mesos instance {{$labels.instance}} cannot be scraped."
  }

ALERT MesosMasterLeader
  IF sum(mesos_master_elected{job="mesos-master"}) != 1
  FOR 5m
  LABELS { severity="page" }
  ANNOTATIONS {
    description="Agents and frameworks require a unique leading Mesos master.",
    summary="Expected one leading Mesos master but there are {{ $value }}."
  }

ALERT MesosMasterTooManyRestarts
  IF resets(mesos_master_uptime_seconds{job="mesos-master"}[1h]) > 10
  FOR 5m
  LABELS { severity="page" }
  ANNOTATIONS {
    description="The number of seconds the process has been running is resetting regularly.",
    summary="The Mesos master {{$labels.instance}} has restarted {{ $value }} times in the last hour."
  }

ALERT MesosSlaveActive
  IF sum(mesos_master_slaves_state{state="active"}) < 0.9 * count(up{job="mesos-slave"})
  FOR 5m
  LABELS { severity="page" }
  ANNOTATIONS {
    description="Mesos agents must be registered with the master in order to receive tasks.",
    summary="More than 10% of all Mesos agents dropped out. Only {{ $value }} active agents remaining."
  }

ALERT MesosSlaveTooManyRestarts
  IF resets(mesos_slave_uptime_seconds{job="mesos-slave"}[1h]) > 10
  FOR 5m
  LABELS { severity="page" }
  ANNOTATIONS {
    description="The number of seconds the process has been running is resetting regularly.",
    summary="The Mesos agent {{$labels.instance}} has restarted {{ $value }} times in the last hour."
  }
```
