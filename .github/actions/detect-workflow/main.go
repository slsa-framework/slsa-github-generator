// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/github"
)

// getStr returns a value from the given Event map. Values are specified as "."
// separated indexes into the map. e.g. pull_request.head.repo.full_name.
func getStr(m map[string]interface{}, key string) string {
	if key == "" {
		return ""
	}

	parts := strings.Split(key, ".")

	current := m[parts[0]]
	for _, part := range parts[1:] {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[string]string:
			current = v[part]
		default:
			return ""
		}
	}

	switch v := current.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func getRepoRef(ctx context.Context, c *github.OIDCClient) (string, string, error) {
	var repository, ref string

	ghContext, err := github.GetWorkflowContext()
	if err != nil {
		return "", "", fmt.Errorf("getting github context: %w", err)
	}

	// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove special logic for pull_requests.
	if ghContext.EventName == "pull_request" {
		// If a pull request get the repo from the pull request.
		repository = getStr(ghContext.Event, "pull_request.head.repo.full_name")
		ref = ghContext.HeadRef
	} else {
		audience := ghContext.Repository
		if audience == "" {
			return "", "", errors.New("missing github repository context")
		}
		audience = path.Join(audience, "detect-workflow")

		t, err := c.Token(ctx, []string{audience})
		if err != nil {
			return "", "", fmt.Errorf("getting OIDC token: %w", err)
		}

		pathParts := strings.SplitN(t.JobWorkflowRef, "/", 3)
		if len(pathParts) < 3 {
			return "", "", errors.New("missing org/repository in job workflow ref")
		}
		repository = strings.Join(pathParts[:2], "/")

		refParts := strings.Split(t.JobWorkflowRef, "@")
		if len(refParts) < 2 {
			return "", "", errors.New("missing reference in job workflow ref")
		}
		ref = refParts[1]
	}

	return repository, ref, nil
}

func main() {
	c, err := github.NewOIDCClient()
	if err != nil {
		log.Fatal(err)
	}
	repository, ref, err := getRepoRef(context.Background(), c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`::set-output name=repository::%s`, repository))
	fmt.Println(fmt.Sprintf(`::set-output name=ref::%s`, ref))
}
