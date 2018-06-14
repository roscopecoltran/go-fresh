package updater

import (
	"context"

	"github.com/paultyng/go-fresh/depmap"
)

// Submitter represents an implementation that can SubmitPR's
type Submitter interface {
	SubmitPR(ctx context.Context, project depmap.Project, dependency, fromrev, toversion, torev string) error
}

// SubmitPR submits a PR using the default configured Submitter implementation.
func SubmitPR(ctx context.Context, project depmap.Project, dependency, fromrev, toversion, torev string) error {
	submitter, err := newNomadSubmitter()
	if err != nil {
		return err
	}
	return submitter.SubmitPR(ctx, project, dependency, fromrev, toversion, torev)
}
