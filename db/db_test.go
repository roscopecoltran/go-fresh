package db

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/require"
)

func TestProjectsForDependency(t *testing.T) {
	for i, c := range []struct {
		expected []string
		key      string
		data     map[string][]string
	}{
		{
			[]string{"org2/proj1"},
			"org1/dep1",
			map[string][]string{
				"org1/dep1": []string{"org2/proj1"},
			},
		},
		{
			[]string{"org2/proj1"},
			"Org1/Dep1",
			map[string][]string{
				"org1/dep1": []string{"org2/proj1"},
			},
		},
		{
			[]string{"org2/proj1"},
			"org1/dep1",
			map[string][]string{
				"org1/dep1/subdep1": []string{"org2/proj1"},
			},
		},

		{
			[]string{},
			"org1/dep1",
			map[string][]string{
				"org1/dep1foo": []string{"org2/proj1"},
			},
		},
		{
			[]string{},
			"org1/dep1",
			map[string][]string{
				"fooorg1/dep1": []string{"org2/proj1"},
			},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert := require.New(t)

			tmp, err := ioutil.TempDir("", "")
			assert.NoError(err)

			path := filepath.Join(tmp, "bolt.db")

			bdb, err := bolt.Open(path, 0644, nil)
			assert.NoError(err)
			defer bdb.Close()

			assert.NoError(bdb.Update(func(tx *bolt.Tx) error {
				bucket, err := tx.CreateBucketIfNotExists(bucketDependencyProjects)
				if err != nil {
					return err
				}

				for dep, projects := range c.data {
					children, err := bucket.CreateBucketIfNotExists(projectKey(dep))
					if err != nil {
						return err
					}

					for _, p := range projects {
						err = children.Put(projectKey(p), nil)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}))

			client := NewBoltClient(bdb)
			actual, err := client.ProjectsForDependency(c.key)
			assert.NoError(err)

			assert.Equal(c.expected, actual)
		})
	}
}
