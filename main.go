package main

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/paultyng/go-fresh/cmd"
)

func main() {
	ui := &cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		Ui: &cli.BasicUi{
			ErrorWriter: os.Stderr,
			Reader:      os.Stdin,
			Writer:      os.Stdout,
		},
	}

	c := cli.NewCLI("tfprovlint", "1.0.0")

	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"pr submit": cmd.PRSubmitCommandFactory(ui),

		"project register": cmd.ProjectRegisterCommandFactory(ui),

		"github listen": cmd.GithubListenCommandFactory(ui),
		"github watch":  cmd.GithubWatchCommandFactory(ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
