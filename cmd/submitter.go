package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/paultyng/go-fresh/updater"
)

type submitterCommand struct {
}

func (c submitterCommand) Flags(m *meta) error {
	m.Flags.StringP("submitter", "s", "logonly", "method to use for PR submission")

	m.Flags.String("nomad-address", "http://127.0.0.1:4646", "address to Nomad API")
	m.Flags.String("nomad-region", "global", "Nomad region")

	return nil
}

func (c submitterCommand) Submitter(r *run) (updater.Submitter, error) {
	t, err := r.flags.GetString("submitter")
	if err != nil {
		return nil, err
	}

	r.ui.Info(fmt.Sprintf("using submitter type %q", t))

	switch t {
	case "logonly":
		return updater.NewLogOnlySubmitter(), nil
	case "nomad":
		address, err := r.flags.GetString("nomad-address")
		if err != nil {
			return nil, err
		}
		region, err := r.flags.GetString("nomad-region")
		if err != nil {
			return nil, err
		}
		return updater.NewNomadSubmitter(address, region)
	default:
		return nil, errors.Errorf("unexpected submitter type %q", t)
	}
}
