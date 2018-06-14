package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/boltdb/bolt"
	throttle "github.com/boz/go-throttle"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/paultyng/go-fresh/data"
	"github.com/paultyng/go-fresh/updater"
)

type githubWatchCommand struct {
	UI cli.Ui

	db data.Client
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
	const perPage = 200
	//const sleep = 750 * time.Millisecond // 4800 per hour
	const sleep = 1 * time.Hour / 4950
	const tickBacklog = 100

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

	token := os.Getenv("GITHUB_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// TODO: i imagine this will eventually grow too large, should this eject?
	// should it be an inverse bloom or something?
	observedKeys := map[string]bool{}

	log.Printf("check every %v", sleep)

	// buffer 10 ticks just in case
	after := make(chan time.Time, tickBacklog)
	go func() {
		tick := time.Tick(sleep)
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-tick:
				after <- t
			}
		}
	}()

	_, resp, err := client.RateLimits(ctx)
	if err != nil {
		return err
	}
	rate := resp.Rate
	rateLimitThrottle := throttle.ThrottleFunc(5*time.Minute, false, func() {
		log.Printf("api calls reamining: %d, reset at %v", rate.Remaining, rate.Reset)
	})
	defer rateLimitThrottle.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-after:
			if len(after) > 1 {
				log.Printf("%d tick backlog", len(after))
			}

			events, resp, err := client.Activity.ListEvents(ctx, &github.ListOptions{
				PerPage: perPage,
			})
			if err != nil {
				if _, ok := err.(*github.RateLimitError); ok {
					// if its a rate limit error, just try next tick
					continue
				}
				return err
			}

			rate = resp.Rate
			rateLimitThrottle.Trigger()

			newEvents := make([]*github.Event, 0, len(events))
			for _, e := range events {
				if observedKeys[e.GetID()] {
					//skipping as its a repeat
					continue
				}
				newEvents = append(newEvents, e)
			}

			// if its not the first pass and the new events == page size, this method isn't fast enough
			if len(newEvents) == len(events) && len(observedKeys) > 0 {
				log.Println("not fast enough!")
			}

			err = processEvents(ctx, c.db, newEvents)
			if err != nil {
				return err
			}

			for _, e := range newEvents {
				observedKeys[e.GetID()] = true
			}
		}
	}
}

func processEvents(ctx context.Context, db data.Client, events []*github.Event) error {
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

func shouldIgnoreReleaseEvent(event *github.ReleaseEvent) bool {
	if event == nil || event.Repo == nil || event.Repo.Name == nil {
		// skipping event, release event does not have necessary information
		return true
	}

	if event.Release.GetPrerelease() {
		// skipping %s, release event is flagged pre-release", repoName)
		return true
	}

	tagName := event.Release.GetTagName()
	v, err := semver.NewVersion(tagName)
	if err == semver.ErrInvalidSemVer {
		// kipping %s, release event tag is not valid semver: %q", repoName, tagName)
		return true
	}
	if err != nil {
		// not sure what would cause this, but should just skip
		// TODO: log error?
		return true
	}

	if v.Prerelease() != "" {
		// skipping %s, release event tag has pre-release information: %q", repoName, tagName)
		return true
	}

	return false
}

func processReleaseEvent(ctx context.Context, db data.Client, event *github.ReleaseEvent) error {
	if shouldIgnoreReleaseEvent(event) {
		return nil
	}

	repoName := event.Repo.GetName()
	// need to append github.com/ for github Go projects
	depName := fmt.Sprintf("github.com/%s", repoName)

	tagName := event.Release.GetTagName()
	v, _ := semver.NewVersion(tagName)

	log.Printf("release for %s, %q", repoName, v.String())

	// TODO: need to fork this off to a new go routine so that tick listening doesn't backup too much

	keys, err := db.ProjectsForDependency(depName)
	if err != nil {
		return err
	}

	for _, k := range keys {
		log.Printf("submitting PR for %s, bump %s to %s\n", k, repoName, v.String())
		project, _, err := db.Project(k)
		if err == data.ErrNotFound {
			continue
		}
		if err != nil {
			return err
		}

		// TODO: lookup current revision in deps
		fromRev := ""

		err = updater.SubmitPR(ctx, project, depName, fromRev, v.String(), event.GetRelease().GetTargetCommitish())
		if err != nil {
			return err
		}
	}

	return nil
}
