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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/github"
)

type action struct {
	getenv func(string) string
	event  map[string]any
	client *github.OIDCClient
}

// TODO(github.com/slsa-framework/slsa-github-generator/issues/164): use the github context via the shared library

func newAction(getenv func(string) string, c *github.OIDCClient) (*action, error) {
	eventPath := getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, errors.New("GITHUB_EVENT_PATH not set")
	}

	payload, err := os.ReadFile(eventPath)
	if err != nil {
		return nil, err
	}

	var event map[string]any
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}

	return &action{
		getenv: getenv,
		event:  event,
		client: c,
	}, nil
}

// getEventValue returns a string value from the given Event map. Values are specified
// as dot-separated indexes into the map. e.g.
// "pull_request.head.repo.full_name".
func (a *action) getEventValue(key string) string {
	if key == "" {
		return ""
	}

	m := a.event
	parts := strings.Split(key, ".")

	// Traverse the first parts of the path.
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

	// Return the final part if it's a string.
	switch v := current.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func (a *action) getRepoRef(ctx context.Context) (string, string, error) {
	var repository, ref string

	// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove special logic for pull_requests.
	eventName := a.getenv("GITHUB_EVENT_NAME")
	if eventName == "pull_request" {
		// If a pull request get the repo from the pull request.
		repository = a.getEventValue("pull_request.head.repo.full_name")
		ref = a.getenv("GITHUB_HEAD_REF")
	} else {
		audience := a.getenv("GITHUB_REPOSITORY")
		if audience == "" {
			return "", "", errors.New("missing github repository context")
		}
		audience = path.Join(audience, "detect-workflow")

		t, err := a.client.Token(ctx, []string{audience})
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
	a, err := newAction(os.Getenv, c)
	if err != nil {
		log.Fatal(err)
	}

	repository, ref, err := a.getRepoRef(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`::set-output name=repository::%s`, repository))
	fmt.Println(fmt.Sprintf(`::set-output name=ref::%s`, ref))
}
