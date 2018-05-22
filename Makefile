# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-agent: ## build the threatseer agent
	docker build -t dustindecker/threatseer-agent . -f Dockerfile.agent

build-server: ## build the threatseer agent
	docker build -t dustindecker/threatseer-server . -f Dockerfile.server

run-agent: ## run the agent docker image
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

run-agent: ## run the server docker image
	docker run \
		--name threatseer-server \
		--rm \
		-it \
		--net=host \
		dustindecker/threatseer-server

build-local: ## build agent and server locally, without docker
	go build -o bin/agent agent/main.go
	go build -o bin/server server/main.go



clean: ## remove binaries
	rm -rf bin/*