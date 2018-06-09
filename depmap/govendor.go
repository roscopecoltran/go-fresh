package depmap

import (
	"os"

	"github.com/kardianos/govendor/vendorfile"
	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
)

func tryGovendor(tree *git.Worktree) ([]Dependency, error) {
	const vendorJSON = "vendor/vendor.json"

	f, err := tree.Filesystem.Open(vendorJSON)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, errDepSystemNotUsed
		}
		return nil, errors.Wrapf(err, "unable to open file govendor vendor file %s", vendorJSON)
	}
	defer f.Close()

	vf := &vendorfile.File{}
	err = vf.Unmarshal(f)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal govendor vendorfile")
	}

	deps := make([]Dependency, 0, len(vf.Package))

	for _, pkg := range vf.Package {
		if pkg == nil {
			continue
		}

		deps = append(deps, Dependency{
			Name:     pkg.Path,
			Revision: pkg.Revision,
			Source:   pkg.Origin,
			// pkg.Version
			// pkg.VersionExact
		})
	}

	return deps, nil
}
