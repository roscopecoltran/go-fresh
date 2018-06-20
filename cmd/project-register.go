package cmd

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/mitchellh/cli"

	"github.com/go-fresh/go-fresh/data"
	"github.com/go-fresh/go-fresh/depmap"
)

type projectRegisterCommand struct {
	boltCommand

	db data.Client
}

// ProjectRegisterCommandFactory creates the "project register" command
func ProjectRegisterCommandFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &projectRegisterCommand{}
	return newCommandFactory(ui, "project register", cmd, func(m *meta) error {
		m.Synopsis = "ingests dependencies for a project for future watching"

		// m.Flags.StringP("type", "t", "github", "type of project (VCS host)")
		m.Flags.StringP("branch", "b", "master", "branch of project to PR into")

		return m.Register(
			cmd.boltCommand,
		)
	})
}

func (c *projectRegisterCommand) Run(ctx context.Context) error {
	branch, err := flags(ctx).GetString("branch")
	if err != nil {
		return err
	}

	// TODO: pass in from args/flags
	p := depmap.Project{
		Name:   "github.com/terraform-providers/terraform-provider-github",
		GitURL: "https://github.com/terraform-providers/terraform-provider-github.git",
		Branch: branch,
	}

	// TODO: flag for tmp dir
	tmp, err := ioutil.TempDir("", "go-fresh")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	bdb, err := c.DB(ctx)
	if err != nil {
		return err
	}
	defer bdb.Close()
	c.db = data.NewBoltClient(bdb)

	return c.registerProject(ctx, tmp, p)
}

func (c *projectRegisterCommand) registerProject(ctx context.Context, tmpDir string, project depmap.Project) error {
	deps, _, err := project.Dependencies(ctx)
	if err != nil {
		return err
	}

	return c.db.RegisterProject(project, deps)
}
