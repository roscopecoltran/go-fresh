package cmd

import (
	"context"

	"github.com/mitchellh/cli"
)

type prSubmitCommand struct {
}

// PRSubmitCommandFactory creates the "pr submit" command
func PRSubmitCommandFactory(ui cli.Ui) cli.CommandFactory {
	return newCommandFactory(ui, "pr submit", &prSubmitCommand{}, func(cmd *meta) error {
		cmd.Synopsis = "submits a pr for a project to update a dependency"
		// cmd.Flags
		return nil
	})
}

func (c *prSubmitCommand) Run(ctx context.Context, r *run) error {
	panic("not implemented")
}
