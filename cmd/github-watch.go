package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/paultyng/go-fresh/db"
)

type githubWatchCommand struct {
	UI cli.Ui

	db db.Client
}

func (c *githubWatchCommand) Help() string {
	// go-fresh github watch will poll/process github's public events stream for:
	// new releases: `ReleaseEvent`
	// code pushes in monitored repo/branches
	return "help!"
}

func (c *githubWatchCommand) Synopsis() string {
	return "polls/processes github's public events stream"
}

// GithubWatchCommandFactory creates the "github watch" command
func GithubWatchCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &githubWatchCommand{
			UI: ui,
		}, nil
	}
}

func (c *githubWatchCommand) Run(args []string) int {
	// TODO: write a shared wrapper for this output
	err := c.run(context.Background(), args)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return 0
}

func (c *githubWatchCommand) run(ctx context.Context, args []string) error {
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

	const perPage = 100

	token := os.Getenv("GITHUB_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	latestKeys := map[string]bool{}
	sleep := 750 * time.Millisecond
	for {
		after := time.After(sleep)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-after:
			events, _, err := client.Activity.ListEvents(ctx, &github.ListOptions{
				PerPage: perPage,
			})
			if err != nil {
				return err
			}
			//fmt.Printf("rate limit %d, remaining %d, reset %v\n", resp.Rate.Limit, resp.Rate.Remaining, resp.Rate.Reset)

			newEvents := make([]*github.Event, 0, len(events))
			for _, e := range events {
				if latestKeys[e.GetID()] {
					//skipping as its a repeat
					continue
				}
				newEvents = append(newEvents, e)
			}

			//fmt.Printf("found %d events, new %d\n", len(events), len(newEvents))

			// if its not the first pass and the new events == page size, this method isn't fast enough
			if len(newEvents) == perPage && len(latestKeys) > 0 {
				log.Println("not fast enough!")
			}

			err = processEvents(ctx, c.db, newEvents)
			if err != nil {
				return err
			}

			for _, e := range newEvents {
				latestKeys[e.GetID()] = true
			}
		}
	}
}

func processEvents(ctx context.Context, db db.Client, events []*github.Event) error {
	for _, e := range events {
		if *e.Type != "ReleaseEvent" {
			continue
		}

		raw, err := e.ParsePayload()
		if err != nil {
			return err
		}
		re, ok := raw.(*github.ReleaseEvent)
		if !ok {
			return errors.Errorf("unable to convert event to ReleaseEvent, got %T", raw)
		}

		// promote repo from event to payload
		re.Repo = e.Repo

		err = processReleaseEvent(ctx, db, re)
		if err != nil {
			return err
		}
	}
	return nil
}

func processReleaseEvent(ctx context.Context, db db.Client, event *github.ReleaseEvent) error {
	if event == nil || event.Repo == nil || event.Repo.Name == nil {
		// skip event
		return nil
	}

	if event.Release.GetPrerelease() {
		// skip prerelease
		return nil
	}

	// tagName := event.Release.GetTagName()
	// TODO: ensure tagname is valid semver pattern v1.2.3, etc

	// need to append github.com/ for github Go projects
	keys, err := db.ProjectsForDependency(fmt.Sprintf("github.com/%s", event.Repo.GetName()))
	if err != nil {
		return err
	}

	for _, k := range keys {
		fmt.Printf("submitting PR for %s, bump %s to %s\n", k, event.Repo.GetName(), event.Release.GetTagName())
	}

	return nil
}
