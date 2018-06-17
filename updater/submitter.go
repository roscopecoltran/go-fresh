package updater

import (
	"context"
	"log"

	"github.com/go-fresh/go-fresh/depmap"
)

// Submitter represents an implementation that can SubmitPR's
type Submitter interface {
	SubmitPR(ctx context.Context, project depmap.Project, dependency, toversion, torev string) error
}

type logOnlySubmitter struct{}

func NewLogOnlySubmitter() Submitter {
	return &logOnlySubmitter{}
}

func (s *logOnlySubmitter) SubmitPR(ctx context.Context, project depmap.Project, dependency, toversion, torev string) error {
	log.Printf("submit PR for %s, update %s to %s (%s)", project.Name, dependency, toversion, torev)
	return nil
}
