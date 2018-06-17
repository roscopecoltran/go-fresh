package updater

import (
	"context"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/paultyng/go-fresh/depmap"
	"github.com/pkg/errors"
)

const (
	nomadJobIDGovendor = "go-fresh-pr-govendor"
)

type nomadSubmitter struct {
	client  *api.Client
	timeout time.Duration
}

func NewNomadSubmitter(address, region string) (Submitter, error) {
	conf := api.DefaultConfig()
	conf.Address = address
	conf.Region = region
	// conf.TLSConfig.CACert = d.Get("ca_file").(string)
	// conf.TLSConfig.ClientCert = d.Get("cert_file").(string)
	// conf.TLSConfig.ClientKey = d.Get("key_file").(string)
	// conf.SecretID = d.Get("secret_id").(string)

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to configure Nomad API")
	}

	return &nomadSubmitter{
		client:  client,
		timeout: 10 * time.Minute,
	}, nil
}

func (s *nomadSubmitter) JobComplete(ctx context.Context, id string) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			resp, _, err := s.client.Jobs().Summary(id, nil)
			if err != nil {
				return false, err
			}

			totalSummary := api.TaskGroupSummary{}

			for name, tg := range resp.Summary {
				if tg.Failed > 0 || tg.Lost > 0 {
					return false, errors.Errorf("unexpected task group status for %q, job %q", name, id)
				}
				totalSummary.Complete += tg.Complete
				totalSummary.Queued += tg.Queued
				totalSummary.Starting += tg.Starting
				totalSummary.Running += tg.Running
			}

			if totalSummary.Complete == 0 {
				// need at least 1 completion
				continue
			}

			if totalSummary.Queued != 0 || totalSummary.Starting != 0 || totalSummary.Running != 0 {
				// stuff still in progress
				continue
			}

			return true, nil
		}
	}
}

func (s *nomadSubmitter) SubmitPR(ctx context.Context, project depmap.Project, dependency, toversion, torev string) error {
	// QUESTION: does the nomad API not use context.Context?
	resp, _, err := s.client.Jobs().Dispatch(nomadJobIDGovendor, map[string]string{
		"PROJECT":    project.Name,
		"GIT_REMOTE": project.GitURL,
		"GIT_BRANCH": project.Branch,
		"DEPENDENCY": dependency,
		//"FROMREVISION": fromrev,
		"TOVERSION":  toversion,
		"TOREVISION": torev,
	}, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to dispatch nomad job")
	}

	complete, err := s.JobComplete(ctx, resp.DispatchedJobID)
	if err != nil {
		return errors.Wrapf(err, "unexpected error waiting for job completion")
	}
	if !complete {
		return errors.Errorf("PR submission did not complete")
	}

	return nil
}
