package cmd

import (
	"context"

	"github.com/mitchellh/cli"
	flag "github.com/spf13/pflag"
)

type run struct {
	ui    cli.Ui
	flags *flag.FlagSet
	args  []string
}

type runner interface {
	Run(context.Context, *run) error
}

type meta struct {
	Name     string
	Flags    *flag.FlagSet
	Synopsis string
}

type command struct {
	ui     cli.Ui
	meta   *meta
	runner runner
}

type flagger interface {
	Flags(*meta) error
}

func (m *meta) Register(flaggers ...flagger) error {
	for _, f := range flaggers {
		err := f.Flags(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func newCommandFactory(ui cli.Ui, name string, c runner, setup func(*meta) error) cli.CommandFactory {
	return func() (cli.Command, error) {
		m := &meta{
			Name:  name,
			Flags: flag.NewFlagSet(name, flag.ContinueOnError),
		}

		err := setup(m)
		if err != nil {
			return nil, err
		}

		cmd := &command{
			ui:     ui,
			meta:   m,
			runner: c,
		}

		return cmd, nil
	}
}

func (c *command) Run(args []string) int {
	err := c.meta.Flags.Parse(args)
	if err != nil {
		c.ui.Error(err.Error())
		return -2
	}
	err = c.runner.Run(context.Background(), &run{
		args:  args,
		flags: c.meta.Flags,
		ui:    c.ui,
	})
	if err != nil {
		c.ui.Error(err.Error())
		return -1
	}
	return 0
}

func (c *command) Synopsis() string {
	return c.meta.Synopsis
}

func (c *command) Help() string {
	// go-fresh github watch will poll/process github's public events stream for:
	// new releases: `ReleaseEvent`
	// code pushes in monitored repo/branches
	return "help!"
}
