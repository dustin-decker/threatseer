# Running Threatseer

The easiest way to get started is to run the `agent` and `server` Docker images.

## Running the agent

The agent must run as a priviledged container and with the mounts so it can collect telemetry from the kernel and filesystem.

```bash
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
    dustindecker/threatseer:agent-0.1.0
```

## Running the agent

```bash
docker run \
    --name threatseer-server \
    --rm \
    -it \
    --net=host \
    dustindecker/threatseer:server-0.1.0
```