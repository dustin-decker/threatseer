## Local development

### Local build

Fetch the deps:

``` bash
dep ensure
```

Build the binary:

``` bash
# agent
CGO_ENABLED=0 go build -o bin/agent agent/main.go

# server
CGO_ENABLED=0 go build -o bin/server server/main.go
```

Run the agent:

``` bash
sudo ./bin/agent
```

Run the server:

``` bash
./bin/server
```

### Docker build

Make the docker images:

``` bash
make build-agent
make build-server
```

Run the image:


`make run-agent`

`make run-server`

## Makefile targets

```
➜  threatseer git:(master) ✗ make help
help                           this help
protos                         generate protos for the agent API
build-agent                    build the threatseer agent
build-server                   build the threatseer agent
run-agent                      run the agent docker image
run-server                     run the server docker image
run-agent-local                run the agent using locally compiled binaries
run-server-local               run the agent using locally compiled binaries
build-local                    build agent and server locally, without docker
clean                          remove binaries
```
