# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)
SHA=$(shell git describe --match=NeVeRmAtCh --always --abbrev=7 --dirty)

ifeq ($(TAG),)
	VERSION=$(SHA)
else
	VERSION=$(TAG)-$(SHA)
endif

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

protos:
	protoc -I api/ \                              
		-I${GOPATH}/src \
		--go_out=plugins=grpc:api \
		api/api.proto

build-agent: ## build the threatseer agent
	docker build -t dustindecker/threatseer-agent:${VERSION} . -f Dockerfile.agent

build-server-from-local:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o bin/server server/main.go
	docker build -t dustindecker/threatseer-server:${VERSION} . -f Dockerfile.server-from-local

build-server: ## build the threatseer agent
	docker build -t dustindecker/threatseer-server:${VERSION} . -f Dockerfile.server

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
		dustindecker/threatseer-agent::${VERSION}

run-server: ## run the server docker image
	docker run \
		--name threatseer-server \
		--rm \
		-it \
		--net=host \
		dustindecker/threatseer-server:${VERSION}

build-local: ## build agent and server locally, without docker
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o bin/agent agent/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o bin/server server/main.go

deploy-kubernetes:
	envsubst < k8s/deployment.yml | kubectl apply -f -

clean: ## remove binaries
	rm -rf bin/*