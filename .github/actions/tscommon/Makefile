SHELL := /bin/bash
ACTION_NAME = $(shell basename "$$(pwd)")
export UNIT_TESTS_WD := $(shell echo "$$(pwd)")
ifeq ($(origin GITHUB_EVENT_PATH),undefined)
export GITHUB_EVENT_PATH := $(shell readlink -m "$$(pwd)"/../event/path.json)
endif
ifeq ($(origin RUNNER_TEMP),undefined)
export RUNNER_TEMP := $(shell readlink -m "$$(pwd)"/../tmp)
endif


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

node_modules/.installed: package.json package-lock.json
	npm ci
	touch node_modules/.installed

.PHONY: action
action: node_modules/.installed ## Builds the action.
	npm run build

.PHONY: package
package: action ## Builds the distribution package.
# WARNING: nothing here because we don't want the pre-submit to overwrite
# the tarball. To build the tarbal, manually run `npm run package`.

.PHONY: clean
clean:
	rm -rf dist lib node_modules

## Testing
#####################################################################

.PHONY: unit-test
unit-test: node_modules/.installed ## Runs all unit tests.
	# NOTE: Make sure the package builds.
	#npm run build
	npm run test

.PHONY: lint
lint: node_modules/.installed ## Runs eslint.
	npm run format-check
	npm run lint
