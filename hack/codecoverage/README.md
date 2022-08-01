# Go Coverage tool

The goal of the coverage tool is to measure the coverage of the code base for Golang.

## Usage
Execute the following command to get the coverage and store it in a file:
1. `go test  -coverprofile=coverage ./... |  run ./hack/codecoverage/main.go coverage.json 70`
2. The coverage.json contains the percentage of the code coverage that is required to pass for certain packages. This is usually because they don't match the desired coverage.
```json
{
  "github.com/slsa-framework/slsa-github-generator/github":70.4,
  "github.com/slsa-framework/slsa-github-generator/internal/builders/generic": 52.3,
  "github.com/slsa-framework/slsa-github-generator/internal/builders/go":17.1,
  "github.com/slsa-framework/slsa-github-generator/internal/errors":100.0,
  "github.com/slsa-framework/slsa-github-generator/internal/utils":72.1,
  "github.com/slsa-framework/slsa-github-generator/signing/envelope":82.4,
  "github.com/slsa-framework/slsa-github-generator/slsa":54.6
}
```
3. The `COVERAGE_PERCENTAGE` is the percentage of the code coverage that is required to pass for all the packages except the ones that are mentioned in the `THRESHOLD_FILE`.
4. The coverage tool will fail if the coverage is below the threshold for any package.
``` shell
2022/07/29 16:14:41 github.com/slsa-framework/slsa-github-generator/pkg/foo is below the threshold of 71.000000
exit status 1
```

### Design choices

1. The coverage tool should not depend on any other tools. It should work of the results from the `go test` command.
2. Coverage threshold should be configurable for each repository - for example `70%` within the repository.
3. A setting file should override the coverage threshold for a given package within the repository. `github.com/foo/bar/xyz : 61`
4. The coverage tool should use native `go` tools and shouldn't depend on external vendors.
5. The coverage tool should be configurable as part of the PR to fail if the desired threshold is not met.
6. Contributors should be able to run it locally if desired before doing a PR.
