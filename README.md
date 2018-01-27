# threatseer

efficient telemetry on important system events and Meltdown + Spectre exploitation attempts

<p align="center">
  <img src="gopher.png"/>
</p>

So far threatseer is a basic implementation of the examples included with Capsule8. It will quickly evolve to be more.

This project provides:
- structured data on important system events
- SOON: templates for actions under conditions
- SOON: a Kubernetes action daemon (bouncer)
- SOON: Kubernetes, Swarm, and local deployments

## threatseer on Kubernetes


<p align="center">
  <img src="img/threatseer-arch.png"/>
</p>

## container logging
Universal solution. Just log json blobs to stdout.

<p align="center">
  <img src="img/container-logging.png"/>
</p>

## logging pipeline
Enriched, interactive investigation experience with structured data.

<p align="center">
  <img src="img/logging-pipeline.png"/>
</p>
