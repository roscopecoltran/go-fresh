package updater

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
	"github.com/pkg/errors"
)

type nomadSubmitter struct {
	Client *api.Client
}

func newNomadSubmitter() (Submitter, error) {
	conf := api.DefaultConfig()
	conf.Address = "http://127.0.0.1:4646"
	conf.Region = "global"
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

func (s *nomadSubmitter) SubmitPR() error {
	leader, err := s.Client.Status().Leader()
	if err != nil {
		return err
	}
	fmt.Printf("Leader: %s\n", leader)
	panic("not implemented")
}
