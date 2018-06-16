package cmd

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/mitchellh/cli"

	"github.com/paultyng/go-fresh/data"
	"github.com/paultyng/go-fresh/depmap"
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

		return m.Register(
			cmd.boltCommand,
		)
	})
}

func (c *projectRegisterCommand) Run(ctx context.Context, r *run) error {
	projects := []depmap.Project{
		{
			Name:   "github.com/terraform-providers/terraform-provider-github",
			GitURL: "https://github.com/terraform-providers/terraform-provider-github.git",
			Branch: "master",
		},
	}

	// TODO: flag for tmp dir
	tmp, err := ioutil.TempDir("", "go-fresh")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	bdb, err := c.DB(r)
	if err != nil {
		return err
	}
	defer bdb.Close()
	c.db = data.NewBoltClient(bdb)

	for _, p := range projects {
		err = c.registerProject(ctx, tmp, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *projectRegisterCommand) registerProject(ctx context.Context, tmpDir string, project depmap.Project) error {
	deps, _, err := project.Dependencies(ctx)
	if err != nil {
		return err
	}

	return c.db.RegisterProject(project, deps)
}
