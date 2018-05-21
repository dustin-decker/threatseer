# Configuring Threatseer

# agent

Nothing required except server endpoint (currently hardcoded to localhost)

# server

## Logging

Configure `beats.yml` per [the docs](https://www.elastic.co/guide/en/beats/filebeat/current/configuring-output.html) or the documentation in the file.

## Analysis Engines

Configure the `yaml` files in the `config` folder

### Dynamic Rules Engine syntax

Example queries tested [here](https://github.com/caibirdme/yql/blob/master/yql_test.go#L901)