package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/cli"
	flag "github.com/spf13/pflag"
)

type contextKey string

func (k contextKey) String() string {
	return "go-fresh/cmd contextKey: " + string(k)
}

var (
	contextKeyUI    = contextKey("ui")
	contextKeyFlags = contextKey("flags")
	contextKeyArgs  = contextKey("args")
)

func ui(ctx context.Context) cli.Ui {
	return ctx.Value(contextKeyUI).(cli.Ui)
}

func flags(ctx context.Context) *flag.FlagSet {
	return ctx.Value(contextKeyFlags).(*flag.FlagSet)
}

func args(ctx context.Context) []string {
	return ctx.Value(contextKeyArgs).([]string)
}

type runner interface {
	Run(context.Context) error
}

type meta struct {
	Name     string
	Flags    *flag.FlagSet
	Synopsis string
	Help     string
}

type command struct {
	ui     cli.Ui
	meta   *meta
	runner runner
}

type flagger interface {
	Flags(*meta) error
}

func (m *meta) Register(flaggers ...flagger) error {
	for _, f := range flaggers {
		err := f.Flags(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func newCommandFactory(ui cli.Ui, name string, c runner, setup func(*meta) error) cli.CommandFactory {
	return func() (cli.Command, error) {
		m := &meta{
			Name:  name,
			Flags: flag.NewFlagSet(name, flag.ContinueOnError),
		}

		err := setup(m)
		if err != nil {
			return nil, err
		}

		cmd := &command{
			ui:     ui,
			meta:   m,
			runner: c,
		}

		return cmd, nil
	}
}

func (c *command) Run(args []string) int {
	err := c.meta.Flags.Parse(args)
	if err != nil {
		c.ui.Error(err.Error())
		return -2
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyUI, c.ui)
	ctx = context.WithValue(ctx, contextKeyFlags, c.meta.Flags)
	ctx = context.WithValue(ctx, contextKeyArgs, args)

	err = c.runner.Run(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return -1
	}
	return 0
}

func (c *command) Synopsis() string {
	return c.meta.Synopsis
}

func (c *command) Help() string {
	help := c.meta.Help
	if help == "" {
		help = c.meta.Synopsis
	}

	out := &strings.Builder{}
	fmt.Fprintf(out, "Command: %s\n\n%s\n\n", c.meta.Name, help)

	if c.meta.Flags != nil {
		fmt.Fprint(out, "Options:\n\n")
		c.meta.Flags.VisitAll(func(f *flag.Flag) {
			printFlag(out, f)
		})
	}

	return strings.TrimRight(out.String(), "\n")
}

// printFlag prints a single flag to the given writer.
func printFlag(w io.Writer, f *flag.Flag) {
	example, _ := flag.UnquoteUsage(f)
	if example != "" {
		fmt.Fprintf(w, "  -%s=<%s>\n", f.Name, example)
	} else {
		fmt.Fprintf(w, "  -%s\n", f.Name)
	}

	indented := f.Usage
	fmt.Fprintf(w, "%s\n\n", indented)
}
