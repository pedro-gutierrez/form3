// util provides with simple utility types and functions so that
// our main application package is less cluttered
package util

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// A Sqlite3 Repo inherits the behavior from
// a generic sql repo.
type Sqlite3Repo struct {
	SqlRepo
	backend string
}

// NewSqlite3Repo creates a new Sqlite3 backed repo
// from the given configuration
func NewSqlite3Repo(config RepoConfig) (Repo, error) {
	// Define what kind of storage we should
	// use. Use whatever filename is specified, or fallback
	// to memory storage by default
	backend := config.Uri
	if backend == "" {
		backend = "file::memory:?cache=shared"
	}

	// Create a new db struct that holds all the
	// configuration
	repo := &Sqlite3Repo{
		backend: backend,
	}

	// create a shared db connection
	// The caller should defer the call to the Close function
	database, err := sql.Open("sqlite3", backend)
	if err != nil {
		return repo, errors.Wrap(err, "Unable to connect to the database")
	}

	// maybe migrate the database
	if config.Migrations != "" {
		driver, err := sqlite3.WithInstance(database, &sqlite3.Config{})
		if err != nil {
			return repo, errors.Wrap(err, "Could not start migration")
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", config.Migrations),
			"sqlite3", driver)

		if err != nil {
			return repo, errors.Wrap(err, "Migration failed")
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return repo, errors.Wrap(err, "Error while syncing")
		}
	}

	repo.db = database
	return repo, nil
}

// Description returns this database storage type
func (repo *Sqlite3Repo) Description() string {
	return fmt.Sprintf("sqlite3 (%s)", repo.backend)
}
