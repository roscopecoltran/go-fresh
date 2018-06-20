package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

type boltCommand struct {
}

func (c boltCommand) Flags(m *meta) error {
	m.Flags.StringP("db-file", "f", "gofresh.db", "path to database file")

	return nil
}

func (c boltCommand) DB(ctx context.Context) (*bolt.DB, error) {
	dbfile, err := flags(ctx).GetString("db-file")
	if err != nil {
		return nil, err
	}
	dbfile, err = filepath.Abs(dbfile)
	if err != nil {
		return nil, err
	}

	ui(ctx).Info(fmt.Sprintf("using BoltDB file %q", dbfile))
	return bolt.Open(dbfile, 0644, nil)
}
