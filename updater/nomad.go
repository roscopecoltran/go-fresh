package updater

import (
	"context"
	"fmt"

	"github.com/hashicorp/nomad/api"
	"github.com/paultyng/go-fresh/depmap"
	"github.com/pkg/errors"
)

const (
	nomadJobIDGovendor = "go-fresh-pr-govendor"
)

type nomadSubmitter struct {
	Client *api.Client
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
		Client: client,
	}, nil
}

func (s *nomadSubmitter) SubmitPR(ctx context.Context, project depmap.Project, dependency, toversion, torev string) error {
	// QUESTION: does the nomad API not use context.Context?
	resp, _, err := s.Client.Jobs().Dispatch(nomadJobIDGovendor, map[string]string{
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

	fmt.Println(resp.DispatchedJobID)

	// TODO: wait for complete?
	panic("not implemented")

	// QUESTION: how to return data from nomad job? should I just use a queue or something?
}
