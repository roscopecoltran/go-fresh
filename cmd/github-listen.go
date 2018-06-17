package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"

	"github.com/go-fresh/go-fresh/data"
	"github.com/go-fresh/go-fresh/updater"
)

type githubListenCommand struct {
	boltCommand
	submitterCommand

	submitter updater.Submitter
	db        data.Client
	secretKey []byte
	ui        cli.Ui
}

// GithubListenCommandFactory creates the "github watch" command
func GithubListenCommandFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &githubListenCommand{}
	return newCommandFactory(ui, "github listen", cmd, func(m *meta) error {
		m.Synopsis = "listens for GitHub webhooks"

		m.Flags.StringP("bind", "b", "127.0.0.1:4000", "IP and port to bind listener to")
		m.Flags.StringP("secret-key", "k", "", "webhook secret key")

		return m.Register(
			cmd.boltCommand,
			cmd.submitterCommand,
		)
	})
}

func (c *githubListenCommand) Run(ctx context.Context, r *run) error {
	bind, err := r.flags.GetString("bind")
	if err != nil {
		return err
	}

	rawSecretKey, err := r.flags.GetString("secret-key")
	if err != nil {
		return err
	}
	c.secretKey = []byte(rawSecretKey)

	bdb, err := c.DB(r)
	if err != nil {
		return err
	}
	defer bdb.Close()
	c.db = data.NewBoltClient(bdb)

	c.submitter, err = c.Submitter(r)
	if err != nil {
		return err
	}

	return http.ListenAndServe(bind, http.HandlerFunc(c.handleWebhook))
}

func (c *githubListenCommand) handlerError(w http.ResponseWriter, err error) {
	c.ui.Error(fmt.Sprintf("error in handler: %s", err))
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
		err = processReleaseEvent(r.Context(), c.db, c.submitter, event)
		if err != nil {
			c.handlerError(w, err)
			return
		}
	}

}
