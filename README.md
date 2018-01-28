# threatseer

efficient behavioral telemetry and actions on important system events and Meltdown + Spectre exploitation attempts

<p align="center">
  <img src="img/gopher.svg" width="200"/>
</p>


## what is it?

Auditd is a firehose of data, and a lot of it you don't want. Threatseer is backed by [Capsule8](https://github.com/capsule8/capsule8), which makes efficient use of kernel performance and tracing tools like perf, kprobe, the Docker API, and eBPF to provide efficient, event driven behavioral montoring. Hook it up to action daemons and take control of the situation.

So far threatseer is a basic implementation of the examples included with [Capsule8](https://github.com/capsule8/capsule8).

## features

At a high level this project provides:

- event-driven structured data of important system events
  - container lifecycle
  - open() on sensitive data
  - fork, exec, and other risky syscalls
- low resource cost: ~3% of one CPU core, ~20MiB RAM

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

``` json
{
   "Event":{
      "Process":{
         "exec_command_line":[
            "cat",
            "/etc/passwd"
         ],
         "exec_filename":"/bin/cat",
         "type":2
      }
   },
   "cpu":2,
   "credentials":{
      "egid":0,
      "euid":0,
      "fsuid":0,
      "gid":0,
      "sgid":0,
      "suid":0,
      "uid":0
   },
   "id":"a6183871dd954f802087612d7ab5ee1c54ce08fbb4a1f6b8cd8fe7763feef8de",
   "level":"info",
   "msg":"",
   "process_id":"78486fd1d267dc08d348b62bf2a353bd0937695e5d92654f4e8e9a31180f889f",
   "process_pid":3338,
   "sensor_id":"5576a4e456d58e73649af776d07e7eba4b5c8e9d421902409a68c0d6baf48dcf",
   "sensor_monotime_nanos":1517122986914084400,
   "sensor_sequence_number":58,
   "time":"2018-01-28T12:50:46-06:00"
}
```