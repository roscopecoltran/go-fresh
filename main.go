package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/paultyng/go-fresh/depmap"
	"github.com/paultyng/go-fresh/updater"
)

func main() {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "go-fresh")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer os.RemoveAll(tmp)

	err = checkRepositories(ctx, tmp)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkRepositories(ctx context.Context, tmpDir string) error {
	repos := []depmap.Project{
		{
			Name:   "github.com/terraform-providers/terraform-provider-azurerm",
			GitURL: "https://github.com/terraform-providers/terraform-provider-azurerm.git",
			Branch: "master",
		},
	}

	for _, repo := range repos {
		deps, err := repo.Dependencies(ctx)
		if err != nil {
			return err
		}

		updates, err := updater.List(ctx, tmpDir, deps)
		if err != nil {
			return err
		}

		if len(updates) == 0 {
			continue
		}

		fmt.Printf("Updates found for %s\n\n", repo.GitURL)

		for project, u := range updates {
			fmt.Printf("%s\t%s => %s\n", project, u[0].From, u[0].To)
		}
	}

	return nil
}
