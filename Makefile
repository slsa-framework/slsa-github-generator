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
unit-test: go-test ts-test ## Runs all unit tests.

.PHONY: go-test
go-test: ## Run Go unit tests.
	@ # NOTE: go test builds packages even if there are no tests.
	@set -e;\
		go mod vendor; \
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-v"; \
		fi; \
		go test -mod=vendor $$extraeargs ./...

.PHONY: ts-test
ts-test: ## Run TypeScript tests.
	@# Run unit tests for all TS actions where tests are found.
	@set -e;\
		PATHS=$$(find .github/actions/ actions/ -not -path '*/node_modules/*' -name __tests__ -type d | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path unit-test; \
		done

## Tools
#####################################################################

.PHONY: format
format: yaml-format md-format ts-format go-format ## Runs all code formatters.

.PHONY: yaml-format
yaml-format: node_modules/.installed ## Runs code formatter for YAML files.
	@set -e;\
		yml_files=$$( \
			find . -type f \
				\( \
					-name '*.yml' -o \
					-name '*.yaml' \
				\) \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
		); \
		for path in $$yml_files; do \
			./node_modules/.bin/prettier --write $$path; \
		done;

.PHONY: md-format
md-format: node_modules/.installed ## Runs code formatter for Markdown files.
	@set -e;\
		md_files=$$( \
			find . -type f -name '*.md' \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
		); \
		for path in $$md_files; do \
			./node_modules/.bin/prettier --write $$path; \
		done;

.PHONY: ts-format
ts-format: ## Runs code formatter for TypeScript files.
	@set -e;\
		actions_paths=$$(find .github/actions/ actions/ -not -path '*/node_modules/*' -name package.json -type f | xargs dirname); \
		for path in $$actions_paths; do \
			make -C $$path format; \
		done

.PHONY: go-format
go-format: ## Runs code formatter for Go files.
	@set -e;\
		go_files=$$( \
			find . -type f -name '*.go' \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
		); \
		for path in $$go_files; do \
			gofumpt -w $$path; \
		done;

COPYRIGHT ?= "SLSA Authors"
LICENSE ?= apache

.PHONY: autogen
autogen: ## Runs autogen on code files.
	@set -euo pipefail; \
		code_files=$$( \
			find . -type f \
				\( \
					-name '*.go' -o \
					-name '*.ts' -o \
					-name '*.sh' -o \
					-name '*.yaml' -o \
					-name '*.yml' -o \
					-name 'Makefile' \
				\) \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
		); \
		for filename in $${code_files}; do \
			if ! ( head "$${filename}" | grep -iL $(COPYRIGHT) > /dev/null ); then \
				echo $${filename}; \
				cd $$(dirname "$${filename}"); \
				autogen -i --no-code --no-tlc -c $(COPYRIGHT) -l $(LICENSE) $$(basename "$${filename}"); \
				cd - > /dev/null; \
			fi; \
		done

.PHONY: markdown-toc
markdown-toc: node_modules/.installed ## Runs markdown-toc on markdown files.
	@# NOTE: Do not include issue templates since they contain Front Matter.
	@# markdown-toc will update Front Matter even if there is no TOC in the file.
	@# See: https://github.com/jonschlinkert/markdown-toc/issues/151
	@set -euo pipefail; \
		md_files=$$( \
			find . -name '*.md' -type f \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
				-not -iwholename '*/.github/ISSUE_TEMPLATE/*' \
		); \
		for filename in $${md_files}; do \
			npm run markdown-toc "$${filename}"; \
		done;

## Linters
#####################################################################

.PHONY: lint
lint: markdownlint golangci-lint shellcheck eslint yamllint actionlint renovate-config-validator ## Run all linters.

.PHONY: actionlint
actionlint: ## Runs the actionlint linter.
	@# NOTE: We need to ignore config files used in tests.
	@set -e;\
		files=$$( \
			find .github/workflows/ -type f \
				\( \
					-name '*.yaml' -o \
					-name '*.yml' \
				\) \
				-not -iwholename '*/configs-*/*' \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			actionlint -format '{{range $$err := .}}::error file={{$$err.Filepath}},line={{$$err.Line}},col={{$$err.Column}}::{{$$err.Message}}%0A```%0A{{replace $$err.Snippet "\\n" "%0A"}}%0A```\n{{end}}' -ignore 'SC2016:' $${files}; \
		else \
			actionlint $${files}; \
		fi

.PHONY: markdownlint
markdownlint: node_modules/.installed ## Runs the markdownlint linter.
	@set -e;\
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				file=$$(echo "$$p" | jq -c -r '.fileName // empty'); \
				line=$$(echo "$$p" | jq -c -r '.lineNumber // empty'); \
				endline=$${line}; \
				message=$$(echo "$$p" | jq -c -r '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
				exit_code=1; \
				echo "::error file=$${file},line=$${line},endLine=$${endline}::$${message}"; \
			done <<< "$$(./node_modules/.bin/markdownlint --dot --json . 2>&1 | jq -c '.[]')"; \
			exit "$${exit_code}"; \
		else \
			npm run markdownlint; \
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
				level=$$(echo "$$p" | jq -c '.level // empty' | tr -d '"'); \
				file=$$(echo "$$p" | jq -c '.file // empty' | tr -d '"'); \
				line=$$(echo "$$p" | jq -c '.line // empty' | tr -d '"'); \
				endline=$$(echo "$$p" | jq -c '.endLine // empty' | tr -d '"'); \
				col=$$(echo "$$p" | jq -c '.column // empty' | tr -d '"'); \
				endcol=$$(echo "$$p" | jq -c '.endColumn // empty' | tr -d '"'); \
				message=$$(echo "$$p" | jq -c '.message // empty' | tr -d '"'); \
				exit_code=1; \
				case $$level in \
				"info") \
					echo "::notice file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
					;; \
				"warning") \
					echo "::warning file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
					;; \
				"error") \
					echo "::error file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
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

.PHONY: renovate-config-validator
renovate-config-validator: node_modules/.installed ## Runs renovate-config-validator
	@npm run renovate-config-validator

## Maintenance
#####################################################################

.PHONY: clean
clean: ## Delete temporary files.
	rm -rf node_modules
