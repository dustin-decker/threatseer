# threatseer

<p align="center">
  <img src="img/gopher.svg" width="200"/>
</p>

[![Build Status](https://travis-ci.org/dustin-decker/threatseer.svg?branch=master)](https://travis-ci.org/dustin-decker/threatseer)

## what is it?

Threatseer provides efficient behavioral telemetry and actions on important system events and Meltdown + Spectre exploitation attempts.

Auditd collects a firehose of data, and a lot of it you don't want, so you make it work even more to filter it out.  Threatseer is backed by [Capsule8](https://github.com/capsule8/capsule8), which makes use of kernel performance and tracing tools like perf, kprobe, and the Docker API to provide efficient, event driven behavioral montoring. Hook it up to action daemons and take control of the situation.

So far threatseer is a basic implementation of the examples included with [Capsule8](https://github.com/capsule8/capsule8).

## features

At a high level this project provides:

- event-driven structured data of important system events
  - container lifecycle
  - processes touching sensitive data
  - fork, exec, and other risky syscalls
- low resource cost: ~3% of one CPU core, ~20MiB RAM
- ~15mb statically compiled binary deployable

TODO:

- SOON: templates for actions under conditions
- SOON: a Kubernetes daemon to take action under conditions (bouncer)
- SOON: Kubernetes, Swarm, and local deployments
- SOON: Prometheus exporter integration

## getting telemetry

By default events are logged to stdout as JSON blobs. An example universal container logging pipeline described below works well with this.

Alternatively, you can use one of the dozens of [logging hooks](https://github.com/sirupsen/logrus#hooks), make your own logging hook, or use any [io.Writer](https://godoc.org/github.com/sirupsen/logrus#SetOutput).

## threatseer on Kubernetes


<p align="center">
  <img src="img/threatseer-arch.svg" width="500"/>
</p>

## container logging
Universal solution. Just log json blobs to stdout. Ending with producing to Kafka.

<p align="center">
  <img src="img/container-logging.svg" width="500"/>
</p>

## logging pipeline, continued
Enriched, interactive investigation experience with structured data. Starting from consuming from Kafka.

<p align="center">
  <img src="img/logging-pipeline.svg" width="500"/>
</p>


## example telemetry

### L3 cache timing attack (could be Meltdown, Spectre, Rowhammer or others)

``` json
{
   "LLCLoadMissRate":0.9945989,
   "PID":13933,
   "attack":"L3 cache miss timing",
   "hostname":"lol-victimbox1",
   "level":"error",
   "time":"2018-01-28T13:12:13-06:00"
}
```

### container exec

successful blind remote code execution callback

``` json
{
  "Event": {
    "Process": {
      "exec_command_line": [
        "sh",
        "-c",
        "dig +short ifjeow0234f90iwefo2odj.wat.lol"
      ],
      "exec_filename": "/bin/sh",
      "type": 2
    }
  },
  "container_id": "06cba6bc8583000803f75cd4ce88a9723497e716859eb820f35bef48582e9e3f",
  "container_name": "/dazzling_darwin",
  "credentials": {},
  "id": "7d59493a8d9d4ccbee584940628c8bad5ad6a9de7b3762b3138bcab988957e95",
  "image_id": "3fd9065eaf02feaf94d68376da52541925650b81698c53c6824d92ff63f98353",
  "image_name": "alpine",
  "process_pid": 3943,
  "sensor_id": "9a608f32bc59f6d1b5ba579170fff34401ffd1840f3695f9e18a45eef7103125",
  "sensor_monotime_nanos": 1517123007197660400,
  "sensor_sequence_number": 223,
  "time": "2018-01-28T18:04:04-06:00"
}

```