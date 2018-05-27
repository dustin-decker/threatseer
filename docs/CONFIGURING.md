# Configuring Threatseer

# agent

Use `-server` flag to specifiy a remote server (defaults to `127.0.0.1:8081`).

# server

## Daemon

Daemon-level configuration options are in [`/threatseer.yml`](threatseer.yaml).
The daemon config is self-documented.

### Logging
Configure [`threatseer.yml`](/threatseer.yml) per [the docs](https://www.elastic.co/guide/en/beats/filebeat/current/configuring-output.html) or the documentation in [the file](/threatseer.yml).

## Analysis Engines

Configure the `yaml` files in the `config` folder to your needs.

### Dynamic Rules Engine syntax

Example queries tested [here](https://github.com/caibirdme/yql/blob/master/yql_test.go#L901)