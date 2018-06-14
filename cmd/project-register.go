package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/mitchellh/cli"

	"github.com/paultyng/go-fresh/data"
	"github.com/paultyng/go-fresh/depmap"
)

type projectRegisterCommand struct {
	UI cli.Ui

	db data.Client
}

func (c *projectRegisterCommand) Help() string {
	// go-fresh project register will ingest dependencies for a project for future watching
	return "help!"
}

func (c *projectRegisterCommand) Synopsis() string {
	return "ingests dependencies for a project for future watching"
}

// ProjectRegisterCommandFactory creates the "project register" command
func ProjectRegisterCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &projectRegisterCommand{
			UI: ui,
		}, nil
	}
}

func (c *projectRegisterCommand) Run(args []string) int {
	// TODO: write a shared wrapper for this output
	err := c.run(context.Background(), args)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return 0
}

func (c *projectRegisterCommand) run(ctx context.Context, args []string) error {
	projects := []depmap.Project{
		{
			Name:   "github.com/terraform-providers/terraform-provider-azurerm",
			GitURL: "https://github.com/terraform-providers/terraform-provider-azurerm.git",
			Branch: "master",
		},
		{
			Name:   "github.com/terraform-providers/terraform-provider-newrelic",
			GitURL: "https://github.com/terraform-providers/terraform-provider-newrelic.git",
			Branch: "master",
		},
	}

	tmp, err := ioutil.TempDir("", "go-fresh")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// TODO: move this somewhere common
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	dbpath := filepath.Join(dir, "gofresh.db")

	bdb, err := bolt.Open(dbpath, 0644, nil)
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
	deps, err := project.Dependencies(ctx)
	if err != nil {
		return err
	}

	return c.db.RegisterProject(project, deps)
}
