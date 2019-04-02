// util provides with simple utility types and functions so that
// our main application package is less cluttered
package util

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// A Postgres Repo inherits the behavior from
// a generic sql repo.
type PosgresRepo struct {
	SqlRepo
	uri string
}

// NewPostgresRepo creates a new Postgres backed repo
// from the given configuration
func NewPostgresRepo(config RepoConfig) (Repo, error) {

	// Create a new db struct that holds all the
	// configuration
	repo := &PosgresRepo{
		SqlRepo: SqlRepo{
			schema: config.Schema,
		},
		uri: config.Uri,
	}

	// create a shared db connection
	// The caller should defer the call to the Close function
	database, err := sql.Open("postgres", repo.uri)
	if err != nil {
		return repo, errors.Wrap(err, "Unable to connect to the database")
	}

	// maybe migrate the database
	if config.Migrations != "" {

		driver, err := postgres.WithInstance(database, &postgres.Config{})
		if err != nil {
			return repo, errors.Wrap(err, "Could not start migration")
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", config.Migrations),
			"postgres", driver)

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
func (repo *PosgresRepo) Description() string {
	return fmt.Sprintf("postgres (%s)", repo.uri)
}
