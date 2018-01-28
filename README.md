# threatseer

efficient telemetry and actions on important system events and Meltdown + Spectre exploitation attempts

<p align="center">
  <img src="img/gopher.svg" width="200"/>
</p>

So far threatseer is a basic implementation of the examples included with Capsule8. It will quickly evolve to be more.

At a high level this project provides:
- event-driven structured data on important system events
  - container lifecycle
  - fork, exec, and other risky syscalls
- SOON: templates for actions under conditions
- SOON: a Kubernetes daemon to take action under conditions (bouncer)
- SOON: Kubernetes, Swarm, and local deployments

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
