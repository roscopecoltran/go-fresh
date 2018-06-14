package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"

	"github.com/paultyng/go-fresh/db"
)

type githubListenCommand struct {
	UI cli.Ui

	bind      string
	secretKey []byte
	db        db.Client
}

func (c *githubListenCommand) Help() string {
	// go-fresh github listen will listen for Github webhooks for:
	// new releases: `ReleaseEvent`
	// code pushes in monitored repo/branches
	return "help!"
}

func (c *githubListenCommand) Synopsis() string {
	return "listens for github webhooks"
}

// GithubListenCommandFactory creates the "github watch" command
func GithubListenCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &githubListenCommand{
			UI: ui,
		}, nil
	}
}

func (c *githubListenCommand) Run(args []string) int {
	// TODO: write a shared wrapper for this output
	err := c.run(context.Background(), args)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return 0
}

func (c *githubListenCommand) run(ctx context.Context, args []string) error {
	// TODO: load these from flags
	c.bind = "127.0.0.1:4000"
	c.secretKey = []byte("secretkey")

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

	c.db = db.NewBoltClient(bdb)

	return http.ListenAndServe(c.bind, http.HandlerFunc(c.handleWebhook))
}

func (c *githubListenCommand) handlerError(w http.ResponseWriter, err error) {
	c.UI.Error(fmt.Sprintf("error in handler: %s", err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (c *githubListenCommand) handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, c.secretKey)
	if err != nil {
		c.handlerError(w, err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		c.handlerError(w, err)
		return
	}

	switch event := event.(type) {
	case *github.ReleaseEvent:
		err = processReleaseEvent(r.Context(), c.db, event)
		if err != nil {
			c.handlerError(w, err)
			return
		}
	}

}
