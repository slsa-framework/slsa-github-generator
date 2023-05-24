# Copyright 2023 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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

node_modules/.installed: package.json package-lock.json
	npm ci
	touch node_modules/.installed

.PHONY: action
action: node_modules/.installed ## Builds the action.
	npm run build

.PHONY: package
package: action ## Builds the distribution package.
	npm run package

.PHONY: clean
clean:
	rm -rf dist lib node_modules

## Testing
#####################################################################

.PHONY: unit-test
unit-test: node_modules/.installed ## Runs all unit tests.
	# NOTE: Make sure the package builds.
	npm run build
	npm run test

.PHONY: lint
lint: node_modules/.installed ## Runs eslint.
	npm run format-check
	npm run lint
