// util provides with simple utility types and functions so that
// our main application package is less cluttered
package util

import (
	"fmt"
)

// RepoItem represents a generic repo item record
// that can be persisted
type RepoItem struct {
	Id           string `db:"id"`
	Version      int    `db:"version"`
	Organisation string `db:"organisation"`
	Attributes   []byte `db:"attributes"`
}

// Basic repository live information
type RepoInfo struct {
	count int
}

// RepoConfig is a simple container for database configuration
type RepoConfig struct {
	Driver     string
	Uri        string
	Debug      bool
	Migrations string
}

// Repo is a small abstraction of a database
// so that we can easily switch between vendors or even
// storage technology
type Repo interface {
	// A simple description, for logging
	// purposes
	Description() string

	// Repository info
	Info() (RepoInfo, error)

	// A simple function that returns an error if
	// the database is not in a healthy state
	Check() error

	// Close the database
	Close() error

	// Return a finite list of db items
	List(offset int, limit int) ([]*RepoItem, error)

	// Create a new database item
	Create(item *RepoItem) (*RepoItem, error)

	// Update an existing database item
	// and return the new version
	Update(item *RepoItem) (*RepoItem, error)

	// Get all the information for the given
	// repo item. Note: the repo item passed as argument
	// does not need to hold every info about the
	// the item we are interested in, just basic
	// identification data (id, version,etc..)
	Fetch(item *RepoItem) (*RepoItem, error)

	// Delete a single repo item
	Delete(item *RepoItem) error

	// Delete all items from this repo
	DeleteAll() error

	// Defines an abstract way of determining
	// whether the given error represents a database
	// conflict
	IsConflict(err error) bool

	// Defines an abstract way of determining
	// whether the given error represents
	// a item that was not found
	IsNotFound(err error) bool
}

// NewRepo returns a new repo implementation for the given configuration
// or error if not supported. The caller is in charge of
// closing the repo when it is no longer needed.
func NewRepo(config RepoConfig) (Repo, error) {
	var db Repo
	switch config.Driver {
	case "sqlite3":
		return NewSqlite3Repo(config)
	default:
		return db, fmt.Errorf("repo driver not supported: %v", config.Driver)
	}
}