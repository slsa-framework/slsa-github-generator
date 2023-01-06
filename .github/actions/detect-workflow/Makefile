SHELL := /bin/bash
ACTION_NAME = $(shell basename "$$(pwd)")

.PHONY: help
help: ## Shows all targets and help from the Makefile (this message).
	@echo "$(ACTION_NAME) Makefile"
	@echo "Usage: make [COMMAND]"
	@echo ""
	@grep --no-filename -E '^([/a-z.A-Z0-9_%-]+:.*?|)##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = "(:.*?|)## ?"}; { \
			if (length($$1) > 0) { \
				printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2; \
			} else { \
				if (length($$2) > 0) { \
					printf "%s\n", $$2; \
				} \
			} \
		}'

## Testing
#####################################################################

.PHONY: unit-test
unit-test: ## Runs all unit tests.
	go mod vendor
	go test -mod=vendor -v ./...

## Build
#####################################################################

.PHONY: build
build: ## Builds the docker image.
	docker build .
