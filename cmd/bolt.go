package cmd

import (
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

func (c boltCommand) DB(r *run) (*bolt.DB, error) {
	dbfile, err := r.flags.GetString("db-file")
	if err != nil {
		return nil, err
	}
	dbfile, err = filepath.Abs(dbfile)
	if err != nil {
		return nil, err
	}

	r.ui.Info(fmt.Sprintf("using BoltDB file %q", dbfile))
	return bolt.Open(dbfile, 0644, nil)
}
