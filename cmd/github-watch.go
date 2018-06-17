package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/semver"
	throttle "github.com/boz/go-throttle"
	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/go-fresh/go-fresh/data"
	"github.com/go-fresh/go-fresh/updater"
)

type githubWatchCommand struct {
	githubCommand
	boltCommand
	submitterCommand

	db data.Client
}

// GithubWatchCommandFactory creates the "github watch" command
func GithubWatchCommandFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &githubWatchCommand{}
	return newCommandFactory(ui, "github watch", cmd, func(m *meta) error {
		m.Synopsis = "polls/processes github's public events stream"

		// TODO: help:
		// go-fresh github watch will poll/process github's public events stream for:
		// new releases: `ReleaseEvent`
		// code pushes in monitored repo/branches

		return m.Register(
			cmd.githubCommand,
			cmd.boltCommand,
			cmd.submitterCommand,
		)
	})
}

func (c *githubWatchCommand) Run(ctx context.Context, r *run) error {
	const (
		perPage         = 200 // i think max is 100?
		tickBacklog     = 100
		apiCallsPerHour = 4800 // max 5000
		sleep           = 1 * time.Hour / apiCallsPerHour
	)

	bdb, err := c.DB(r)
	if err != nil {
		return err
	}
	defer bdb.Close()
	c.db = data.NewBoltClient(bdb)

	submitter, err := c.Submitter(r)
	if err != nil {
		return err
	}

	client, err := c.GithubClient(ctx, r)
	if err != nil {
		return err
	}

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

	processingErrors := make(chan error)
	go func() {
		select {
		case err := <-processingErrors:
			if err != nil {
				log.Printf("error processing events: %s", err)
			}
		}
	}()

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

			go func() { processingErrors <- processEvents(ctx, c.db, submitter, newEvents) }()

			// record observed keys, this assumes successful processing which may not be the case
			// do not persist this variable as its not entirely accurate outside of the singleton
			// process
			for _, e := range newEvents {
				observedKeys[e.GetID()] = true
			}
		}
	}
}

func processEvents(ctx context.Context, db data.Client, submitter updater.Submitter, events []*github.Event) error {
	for _, e := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

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

			err = processReleaseEvent(ctx, db, submitter, re)
			if err != nil {
				return err
			}
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

func processReleaseEvent(ctx context.Context, db data.Client, submitter updater.Submitter, event *github.ReleaseEvent) error {
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("submitting PR for %s, bump %s to %s\n", k, repoName, v.String())
			project, _, err := db.Project(k)
			if err == data.ErrNotFound {
				continue
			}
			if err != nil {
				return err
			}

			err = submitter.SubmitPR(ctx, project, depName, v.String(), event.GetRelease().GetTargetCommitish())
			if err != nil {
				return err
			}

			log.Printf("PR submitted")
		}
	}

	return nil
}
