package depmap

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var errDepSystemNotUsed = errors.New("dependency system not used")

type Project struct {
	Name   string
	GitURL string
	Branch string
}

type Dependency struct {
	// Required. The package path, not necessarily the project root.
	Name string

	// Required. Text representing a revision or tag.
	Revision string

	// Optional. Alternative source, or fork, for the project.
	Source string
}

var depManagers = []func(*git.Worktree) ([]Dependency, error){
	tryGovendor,
}

func (r *Project) Dependencies(ctx context.Context) ([]Dependency, error) {
	repo, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:           r.GitURL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", r.Branch)),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to clone repository %s", r.GitURL)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load work tree")
	}

	for _, dm := range depManagers {
		deps, err := dm(tree)
		if err == errDepSystemNotUsed {
			continue
		}
		if err != nil {
			return nil, errors.Wrapf(err, "error testing for dep manager")
		}
		return deps, nil
	}

	return nil, errors.New("no dependency management found")
}
