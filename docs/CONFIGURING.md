# Configuring Threatseer

# agent

Important flags:

```
$ ./bin/agent -h
Usage of ./bin/agent:
  -server string
    	remote server to send telemetry to (default "127.0.0.1:8081")
  -tls
    	enable tls
  -ca string
    	custom certificate authority for the remote server to send telemetry to
  -cert string
    	certificate for agent
  -key string
    	key for agent
  -cn string
    	override the expected common name of the remote server
```

See [/docs/TLS.md](/docs/TLS.md) for information on generating certs.

# server

See [/docs/TLS.md](/docs/TLS.md) for information on generating certs.

## Daemon

Daemon-level configuration options are in [`/threatseer.yml`](threatseer.yaml).
The daemon config is self-documented.

### Logging
Configure [`threatseer.yml`](/threatseer.yml) per [the docs](https://www.elastic.co/guide/en/beats/filebeat/current/configuring-output.html) or the documentation in [the file](/threatseer.yml).

## Analysis Engines

Configure the `yaml` files in the `config` folder to your needs.

### Dynamic Rules Engine syntax

Example queries tested [here](https://github.com/caibirdme/yql/blob/master/yql_test.go#L901)