// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"

	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

// NewGithubClient returns a new GitHub API client authenticated using the
// token from the GitHub context.
func NewGithubClient(ctx context.Context) (*github.Client, error) {
	t, err := GetToken()
	if err != nil {
		return nil, err
	}
	return github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: t},
	))), nil
}
