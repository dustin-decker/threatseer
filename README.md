# threatseer

efficient behavioral telemetry and actions on important system events and Meltdown + Spectre exploitation attempts

<p align="center">
  <img src="img/gopher.svg" width="200"/>
</p>


## what is it?

Auditd is a firehose of data, and a lot of it you don't want. Threatseer is backed by [Capsule8](https://github.com/capsule8/capsule8), which makes efficient use of kernel performance and tracing tools like perf, kprobe, the docker API, and eBPF to provide efficient, event driven behavioral montoring. Hook it up to action daemons and take control of the situation.

So far threatseer is a basic implementation of the examples included with [Capsule8](https://github.com/capsule8/capsule8).

## features

At a high level this project provides:

- event-driven structured data of important system events
  - container lifecycle
  - open() on sensitive data
  - fork, exec, and other risky syscalls


- SOON: templates for actions under conditions
- SOON: a Kubernetes daemon to take action under conditions (bouncer)
- SOON: Kubernetes, Swarm, and local deployments
- SOON: Prometheus exporter integration

## getting telemeyry

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
