package cmd

import "github.com/mitchellh/cli"

type prSubmitCommand struct {
	UI cli.Ui
}

func (c *prSubmitCommand) Help() string {
	// go-fresh pr submit will submit a pr for a project to update a dependency
	return "help!"
}

func (c *prSubmitCommand) Synopsis() string {
	return "submits a pr for a project to update a dependency"
}

// PRSubmitCommandFactory creates the "pr submit" command
func PRSubmitCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &prSubmitCommand{
			UI: ui,
		}, nil
	}
}

func (c *prSubmitCommand) Run(args []string) int {
	return -1
}
