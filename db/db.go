package db

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/paultyng/go-fresh/depmap"
)

var (
	bucketProjects            = []byte("projects")
	bucketProjectDependencies = []byte("projectDependencies")

	bucketDependencyProjects = []byte("dependencyProjects")
)

// Client represents the common functions for a database client.
type Client interface {
	ProjectsForDependency(dep string) ([]string, error)
	RegisterProject(p depmap.Project, deps []depmap.Dependency) error
}

type boltClient struct {
	db *bolt.DB
}

func putStruct(b *bolt.Bucket, key []byte, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return b.Put(key, raw)
}

// NewBoltClient constructs a client for a Bolt database.
func NewBoltClient(db *bolt.DB) Client {
	return &boltClient{
		db: db,
	}
}

func depProjectKeyMatch(depKey []byte, test []byte) bool {
	if !bytes.HasPrefix(test, depKey) {
		return false
	}
	// if key equals dep
	if bytes.Equal(test, depKey) {
		return true
	}
	// if key starts with dep + "/"
	depKey = append(depKey, []byte("/")...)
	return bytes.HasPrefix(test, depKey)
}

func (c *boltClient) ProjectsForDependency(dep string) ([]string, error) {
	projectKeys := []string{}
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDependencyProjects)
		if b == nil {
			return nil
		}
		cursor := b.Cursor()
		depKey := projectKey(dep)
		k, _ := cursor.Seek(depKey)
		if k != nil && depProjectKeyMatch(depKey, k) {
			children := b.Bucket(k)
			if children == nil {
				return nil
			}
			return children.ForEach(func(k, v []byte) error {
				projectKeys = append(projectKeys, string(k))
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projectKeys, nil
}

func projectKey(name string) []byte {
	return []byte(strings.ToLower(name))
}

func (c *boltClient) RegisterProject(p depmap.Project, deps []depmap.Dependency) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		key := projectKey(p.Name)

		// project name to project
		{
			bucket, err := tx.CreateBucketIfNotExists(bucketProjects)
			if err != nil {
				return err
			}

			err = putStruct(bucket, key, p)
			if err != nil {
				return err
			}
		}

		// project name to dependency list
		{
			bucket, err := tx.CreateBucketIfNotExists(bucketProjectDependencies)
			if err != nil {
				return err
			}

			err = putStruct(bucket, key, deps)
			if err != nil {
				return err
			}
		}

		// dependency name to project name index
		{
			bucket, err := tx.CreateBucketIfNotExists(bucketDependencyProjects)
			if err != nil {
				return err
			}

			for _, d := range deps {
				children, err := bucket.CreateBucketIfNotExists(projectKey(d.Name))
				if err != nil {
					return err
				}

				err = children.Put(key, nil)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
