SHELL := /bin/bash
OUTPUT_FORMAT ?= $(shell if [ "${GITHUB_ACTIONS}" == "true" ]; then echo "github"; else echo ""; fi)

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

node_modules/.installed: package.json package-lock.json
	npm ci
	touch node_modules/.installed

## Testing
#####################################################################

.PHONY: unit-test
unit-test: go-test ts-test

.PHONY: go-test
go-test: ## Run Go unit tests.
	# Run unit tests for the detect-workflow action.
	make -C .github/actions/detect-workflow/ unit-test
	go mod vendor
	go test -mod=vendor -v ./...


.PHONY: ts-test
ts-test: ## Run TypeScript tests.
	# Run unit tests for the generate-attestations action.
	@set -e;\
		PATHS=$$(find .github/actions/ actions/ -not -path '*/node_modules/*' -name __tests__ -type d | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path unit-test; \
		done

## Linters
#####################################################################

.PHONY: lint
lint: markdownlint golangci-lint shellcheck eslint yamllint ## Run all linters.

.PHONY: markdownlint
markdownlint: node_modules/.installed ## Runs the markdownlint linter.
	@set -e;\
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				FILE=$$(echo "$$p" | jq -c -r '.fileName // empty'); \
				LINE=$$(echo "$$p" | jq -c -r '.lineNumber // empty'); \
				ENDLINE=$${LINE}; \
				MESSAGE=$$(echo "$$p" | jq -c -r '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
				exit_code=1; \
				echo "::error file=$${FILE},line=$${LINE},endLine=$${ENDLINE}::$${MESSAGE}"; \
			done <<< "$$(./node_modules/.bin/markdownlint --dot --json . 2>&1 | jq -c '.[]')"; \
			exit "$${exit_code}"; \
		else \
			npm run lint; \
		fi

.PHONY: golangci-lint
golangci-lint: ## Runs the golangci-lint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="--out-format github-actions"; \
		fi; \
		golangci-lint run -c .golangci.yml ./... $$extraargs

SHELLCHECK_ARGS = --severity=style --external-sources

.PHONY: shellcheck
shellcheck: ## Runs the shellcheck linter.
	@set -e;\
		files=$$(find . -type f -not -iwholename '*/.git/*' -not -iwholename '*/vendor/*' -not -iwholename '*/node_modules/*' -exec bash -c 'file "$$1" | cut -d':' -f2 | grep --quiet shell' _ {} \; -print); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				LEVEL=$$(echo "$$p" | jq -c '.level // empty' | tr -d '"'); \
				FILE=$$(echo "$$p" | jq -c '.file // empty' | tr -d '"'); \
				LINE=$$(echo "$$p" | jq -c '.line // empty' | tr -d '"'); \
				ENDLINE=$$(echo "$$p" | jq -c '.endLine // empty' | tr -d '"'); \
				COL=$$(echo "$$p" | jq -c '.column // empty' | tr -d '"'); \
				ENDCOL=$$(echo "$$p" | jq -c '.endColumn // empty' | tr -d '"'); \
				MESSAGE=$$(echo "$$p" | jq -c '.message // empty' | tr -d '"'); \
				exit_code=1; \
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
			done <<< "$$(echo -n "$$files" | xargs shellcheck -f json $(SHELLCHECK_ARGS) | jq -c '.[]')"; \
			exit "$${exit_code}"; \
		else \
			echo -n "$$files" | xargs shellcheck $(SHELLCHECK_ARGS); \
		fi

.PHONY: eslint
eslint: ## Runs the eslint linter.
	@set -e;\
		PATHS=$$(find .github/actions/ actions/ -not -path '*/node_modules/*' -name package.json -type f | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path lint; \
		done

.PHONY: yamllint
yamllint: ## Runs the yamllint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-f github"; \
		fi; \
		yamllint --strict -c .yamllint.yaml . $$extraargs
