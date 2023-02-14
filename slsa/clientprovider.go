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

package slsa

import (
	"context"

	githubapi "github.com/google/go-github/v50/github"

	"github.com/slsa-framework/slsa-github-generator/github"
)

// ClientProvider creates Github API clients.
type ClientProvider interface {
	OIDCClient() (*github.OIDCClient, error)
	GithubClient(context.Context) (*githubapi.Client, error)
}

// DefaultClientProvider provides a default set of clients based on the Github
// Actions environment.
type DefaultClientProvider struct {
	oidcClient *github.OIDCClient
	ghClient   *githubapi.Client
}

// OIDCClient returns a default OIDC client.
func (p *DefaultClientProvider) OIDCClient() (*github.OIDCClient, error) {
	if p.oidcClient == nil {
		c, err := github.NewOIDCClient()
		if err != nil {
			return nil, err
		}
		p.oidcClient = c
	}
	return p.oidcClient, nil
}

// GithubClient returns a Github API client authenticated with the token
// provided in the github context.
func (p *DefaultClientProvider) GithubClient(ctx context.Context) (*githubapi.Client, error) {
	if p.ghClient == nil {
		c, err := github.NewGithubClient(ctx)
		if err != nil {
			return nil, err
		}
		p.ghClient = c
	}
	return p.ghClient, nil
}

// NilClientProvider does not provide clients. It is useful for testing where
// APIs are not available.
type NilClientProvider struct{}

// OIDCClient returns nil for the client.
func (p *NilClientProvider) OIDCClient() (*github.OIDCClient, error) {
	return nil, nil
}

// GithubClient returns nil for the client.
func (p *NilClientProvider) GithubClient(context.Context) (*githubapi.Client, error) {
	return nil, nil
}
