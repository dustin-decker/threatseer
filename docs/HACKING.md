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

or

``` bash
docker run \
  --privileged \
  --name threatseer-agent \
  --rm \
  -it \
  --net=host \
  -v /proc:/var/run/capsule8/proc/:ro \
  -v /sys/kernel/debug:/sys/kernel/debug \
  -v /sys/fs/cgroup:/sys/fs/cgroup \
  -v /var/lib/docker:/var/lib/docker:ro \
  -v /var/run/docker:/var/run/docker:ro \
  dustindecker/threatseer-agent
```

and 

```bash
docker run \
  --name threatseer-server \
  --rm \
  -it \
  --net=host \
  dustindecker/threatseer-server
```