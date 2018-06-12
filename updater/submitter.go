package updater

import (
	"context"

	"github.com/paultyng/go-fresh/depmap"
)

type Submitter interface {
	SubmitPR(ctx context.Context, project depmap.Project, dependency, fromrev, toversion, torev string) error
}

func SubmitPR(ctx context.Context, project depmap.Project, dependency, fromrev, toversion, torev string) error {
	submitter, err := newNomadSubmitter()
	if err != nil {
		return err
	}
	return submitter.SubmitPR(ctx, project, dependency, fromrev, toversion, torev)
}
