ROOT_PATH := $(shell dirname `pwd`)

.PHONY: generate setup help
.DEFAULT_GOAL := help

generate: ## generate boiler template from schema documents
	@sed -i "s#output: \(.*\)\(internal.*\)#output: $(ROOT_PATH)\/\2#g" types.config.yaml
	@sed -i "s#output: \(.*\)\(internal.*\)#output: $(ROOT_PATH)\/\2#g" server.config.yaml
	oapi-codegen -config types.config.yaml main.yaml
	oapi-codegen -config server.config.yaml main.yaml

setup: ## set up generator
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

help: ## help in Makefile command
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
