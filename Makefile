SHELL := /bin/bash
OUTPUT_FORMAT = $(shell if [ "${GITHUB_ACTIONS}" == "true" ]; then echo "github"; else echo ""; fi)

.PHONY: help
help: ## Shows all targets and help from the Makefile (this message).
	@echo "slsa-github-generator Makefile"
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
unit-test: 
	# Run unit tests for the detect-workflow action.
	make -C .github/actions/detect-workflow/ unit-test
	go mod vendor
	go test -mod=vendor -v ./...

.PHONY: ts-test
ts-test: 
	# Run unit tests for the generate-attestations action.
	make -C .github/actions/generate-attestations/ unit-test


## Linters
#####################################################################

.PHONY: lint
lint: golangci-lint shellcheck eslint yamllint ## Run all linters.

.PHONY: golangci-lint
golangci-lint: ## Runs the golangci-lint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="--out-format github-actions"; \
		fi; \
		golangci-lint run -c .golangci.yml ./... $$extraargs

.PHONY: shellcheck
shellcheck: ## Runs the shellcheck linter.
	@set -e;\
		FILES=$$(find . -type f -not -iwholename '*/.git/*' -not -iwholename '*/vendor/*' -not -iwholename '*/node_modules/*' -exec bash -c 'file "$$1" | cut -d':' -f2 | grep --quiet shell' _ {} \; -print); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			echo -n $$FILES | xargs shellcheck -f json --external-sources | jq -c '.[]' | while IFS="" read -r p || [ -n "$$p" ]; do \
				LEVEL=$$(echo "$$p" | jq -c '.level // empty' | tr -d '"'); \
				FILE=$$(echo "$$p" | jq -c '.file // empty' | tr -d '"'); \
				LINE=$$(echo "$$p" | jq -c '.line // empty' | tr -d '"'); \
				ENDLINE=$$(echo "$$p" | jq -c '.endLine // empty' | tr -d '"'); \
				COL=$$(echo "$$p" | jq -c '.column // empty' | tr -d '"'); \
				ENDCOL=$$(echo "$$p" | jq -c '.endColumn // empty' | tr -d '"'); \
				MESSAGE=$$(echo "$$p" | jq -c '.message // empty' | tr -d '"'); \
				case $$LEVEL in \
				"info") \
					echo "::notice file=$${FILE},line=$${LINE},endLine=$${ENDLINE},col=$${COL},endColumn=$${ENDCOL}::$${MESSAGE}"; \
					;; \
				"warning") \
					echo "::warning file=$${FILE},line=$${LINE},endLine=$${ENDLINE},col=$${COL},endColumn=$${ENDCOL}::$${MESSAGE}"; \
					;; \
				"error") \
					echo "::error file=$${FILE},line=$${LINE},endLine=$${ENDLINE},col=$${COL},endColumn=$${ENDCOL}::$${MESSAGE}"; \
					;; \
				esac; \
			done; \
		else \
			echo -n $$FILES | xargs shellcheck --external-sources; \
		fi

.PHONY: eslint
eslint: ## Runs the eslint linter.
	make -C .github/actions/compute-sha256 lint
	make -C .github/actions/privacy-check lint
	make -C .github/actions/generate-attestations lint

.PHONY: yamllint
yamllint: ## Runs the yamllint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-f github"; \
		fi; \
		yamllint -c .yamllint.yaml . $$extraargs
