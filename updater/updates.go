package updater

import (
	"context"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/golang/dep/gps"
	"github.com/pkg/errors"

	"github.com/go-fresh/go-fresh/depmap"
)

// Update represents a dependency update to perform.
type Update struct {
	ProjectRoot string
	Name        string
	Revision    string

	From string
	To   string

	// CommitsBehind int
	// TimeBehind    time.Duration
}

// List returns all the dependency updates possible on a list of dependencies.
func List(ctx context.Context, tmpDir string, deps []depmap.Dependency) (map[string][]Update, error) {
	projects := map[gps.ProjectRoot][]depmap.Dependency{}
	updates := map[string][]Update{}

	smgr, err := gps.NewSourceManager(gps.SourceManagerConfig{
		DisableLocking: false,
		Cachedir:       tmpDir,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create dep/gps source manager")
	}

	for _, dep := range deps {
		pr, err := smgr.DeduceProjectRoot(dep.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to deduce project root for %s", dep.Name)
		}
		projects[pr] = append(projects[pr], dep)
	}

	for project, projectDeps := range projects {
		// urls, err := smgr.SourceURLsForPath(string(root))
		// if err != nil {
		// 	return nil, errors.Wrapf(err, "unable to determine source urls for %s", root)
		// }
		raw, err := smgr.ListVersions(gps.ProjectIdentifier{
			ProjectRoot: project,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to list versions for %s", project)
		}

		branches := []string{}
		vs := make([]semver.Version, 0, len(raw))
		pairs := map[semver.Version]gps.PairedVersion{}
		currentVersions := map[depmap.Dependency]string{}

		for _, r := range raw {
			rs := r.String()
			v, err := semver.NewVersion(rs)
			if err == semver.ErrInvalidSemVer {
				branches = append(branches, rs)
				continue
			}
			if err != nil {
				return nil, errors.Wrapf(err, "unable to parse semver for %s", rs)
			}

			pairs[v] = r
			for _, dep := range projectDeps {
				if dep.Revision != r.Revision().String() {
					continue
				}

				currentVersions[dep] = v.String()
			}

			if v.Prerelease() != "" {
				// skip prerelease
				continue
			}

			vs = append(vs, v)
		}

		if len(vs) == 0 {
			// no versions, skip project
			continue
		}

		sorted := semver.Collection(vs)
		sort.Sort(sorted)

		latest := sorted[len(sorted)-1]
		latestPair := pairs[latest]

		for _, dep := range projectDeps {
			if dep.Revision == latestPair.Revision().String() {
				// already on latest
				continue
			}

			updates[string(project)] = append(updates[string(project)], Update{
				Name:        dep.Name,
				ProjectRoot: string(project),
				Revision:    latestPair.Revision().String(),

				From: currentVersions[dep],
				To:   latest.String(),
			})
		}
	}

	return updates, nil
}
