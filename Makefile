# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-agent: ## build the threatseer agent
	docker build -t dustindecker/threatseer . -f build/package/Dockerfile.agent
	# docker run --name builder-copy-deployable dustindecker/threatseer /bin/true
	# docker cp builder-copy-deployable:/agent ./bin/agent
	# docker rm -f builder-copy-deployable

clean: ## remove binaries
	rm -rf bin/*
