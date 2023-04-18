package github

import (
	"context"

	"github.com/google/go-github/v51/github"
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
