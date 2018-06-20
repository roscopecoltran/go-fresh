package cmd

import (
	"context"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/go-fresh/go-fresh/data"
)

type prSubmitCommand struct {
	boltCommand
	submitterCommand
}

// PRSubmitCommandFactory creates the "pr submit" command
func PRSubmitCommandFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &prSubmitCommand{}
	return newCommandFactory(ui, "pr submit", cmd, func(m *meta) error {
		m.Synopsis = "submits a pr for a project to update a dependency"

		m.Flags.StringP("project", "p", "", "project for which to submit PR")
		m.Flags.StringP("dependency", "d", "", "dependency to update")
		m.Flags.StringP("to-version", "t", "", "uptdate to version")

		return m.Register(
			cmd.boltCommand,
			cmd.submitterCommand,
		)
	})
}

func (c *prSubmitCommand) Run(ctx context.Context) error {
	projectName, err := flags(ctx).GetString("project")
	if err != nil {
		return err
	}
	if projectName == "" {
		return errors.Errorf("project is required")
	}
	dependency, err := flags(ctx).GetString("dependency")
	if err != nil {
		return err
	}
	if dependency == "" {
		return errors.Errorf("dependency is required")
	}
	toversion, err := flags(ctx).GetString("to-version")
	if err != nil {
		return err
	}
	if toversion == "" {
		return errors.Errorf("to-version is required")
	}
	// TODO: parse toversion to check valid semver?

	bdb, err := c.DB(ctx)
	if err != nil {
		return err
	}
	defer bdb.Close()
	db := data.NewBoltClient(bdb)

	project, _, err := db.Project(projectName)
	if err != nil {
		return err
	}

	submitter, err := c.Submitter(ctx)
	if err != nil {
		return err
	}

	return submitter.SubmitPR(ctx, project, dependency, toversion)
}
