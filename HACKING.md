## Local development

### Local build

Fetch the deps:

``` bash
dep ensure
```

Build the binary:

``` bash
CGO_ENABLED=0 go build -o bin/agent cmd/agent/main.go
```

Run the binary (pretty printed with jq):

``` bash
sudo ./bin/agent 2>&1 | jq '.'
```

### Docker build

Make the docker image:

``` bash
make build-agent
```

Run the image:

``` bash
docker run \
  --privileged \
  --name threatseer \
  --rm \
  -it \
  -v /proc:/var/run/capsule8/proc/:ro \
  -v /sys/kernel/debug:/sys/kernel/debug \
  -v /sys/fs/cgroup:/sys/fs/cgroup \
  -v /var/lib/docker:/var/lib/docker:ro \
  -v /var/run/docker:/var/run/docker:ro \
  dustindecker/threatseer
```
