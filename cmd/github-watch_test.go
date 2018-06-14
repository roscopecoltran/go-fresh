package cmd

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestShouldIgnoreReleaseEvent(t *testing.T) {
	var ptrString = func(s string) *string {
		return &s
	}
	var ptrBool = func(b bool) *bool {
		return &b
	}
	// skipping event, release event does not have necessary information
	// skipping %s, release event is flagged pre-release", repoName)
	// kipping %s, release event tag is not valid semver: %q", repoName, tagName)
	// not sure what would cause this, but should just skip
	// skipping %s, release event tag has pre-release information: %q", repoName, tagName)
	for i, c := range []struct {
		expected bool
		event    *github.ReleaseEvent
	}{
		{true, nil},
		{true, &github.ReleaseEvent{}},
		{true, &github.ReleaseEvent{
			Repo: &github.Repository{},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{},
			Release: &github.RepositoryRelease{},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("baz")},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("v1.2.3.4")},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("1.2.3.4")},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("v1.2.3-pre")},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("1.2.3-pre")},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("v1.2.3"), Prerelease: ptrBool(true)},
		}},
		{true, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("1.2.3"), Prerelease: ptrBool(true)},
		}},

		{false, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("v1.2.3")},
		}},
		{false, &github.ReleaseEvent{
			Repo:    &github.Repository{Name: ptrString("foo/bar")},
			Release: &github.RepositoryRelease{TagName: ptrString("1.2.3")},
		}},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := shouldIgnoreReleaseEvent(c.event)
			assert.Equal(t, c.expected, actual)
		})
	}
}
