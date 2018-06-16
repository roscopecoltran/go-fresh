package cmd

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type githubCommand struct {
}

func (c githubCommand) Flags(m *meta) error {
	m.Flags.StringP("github-token", "t", "", "GitHub access token")

	return nil
}

func (c githubCommand) GithubClient(ctx context.Context, r *run) (*github.Client, error) {
	token, err := r.flags.GetString("github-token")
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, nil
}
