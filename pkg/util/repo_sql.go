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
	"log"
	"strings"
)

// A Sqlite3 Repo
type Sqlite3Repo struct {
	debug         bool
	backend       string
	db            *sql.DB
	countStmt     string
	deleteAllStmt string
	deleteOneStmt string
	listStmt      string
	fetchStmt     string
	createStmt    string
	updateStmt    string
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
		debug:         config.Debug,
		backend:       backend,
		countStmt:     "SELECT COUNT(*) FROM payments WHERE deleted = 0",
		deleteAllStmt: "DELETE FROM payments",
		// We are not sorting payments in any particular order
		listStmt:      "SELECT id, version, organisation, attributes FROM payments WHERE deleted = 0 LIMIT ?, ?",
		fetchStmt:     "SELECT id, version, organisation, attributes FROM payments WHERE id = ? AND deleted = 0",
		createStmt:    "INSERT INTO payments (id, version, organisation, attributes) VALUES (?, ?, ?, ?)",
		updateStmt:    "UPDATE payments SET attributes=?, version=? WHERE id=? AND version=?",
		deleteOneStmt: "UPDATE payments SET deleted=1 WHERE id=? AND version=?",
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

		log.Printf("Database migrated")
	}

	repo.db = database
	return repo, nil
}

// Description returns this database storage type
func (repo *Sqlite3Repo) Description() string {
	return fmt.Sprintf("sqlite3 (%s)", repo.backend)
}

// Close the database
func (repo *Sqlite3Repo) Close() error {
	log.Printf("Closing: %v", repo.Description())
	return repo.Close()
}

// Check performs a simple check on the database
func (repo *Sqlite3Repo) Check() error {
	return repo.db.Ping()
}

// List Return a list of db items. Ignore items marked
// as deleted
func (repo *Sqlite3Repo) List(offset int, limit int) ([]*RepoItem, error) {
	items := []*RepoItem{}

	rows, err := repo.db.Query(repo.listStmt, offset, limit)
	if err != nil {
		return items, errors.Wrap(err, repo.listStmt)
	}

	defer rows.Close()

	for rows.Next() {
		item := &RepoItem{}
		err := rows.Scan(&item.Id, &item.Version, &item.Organisation, &item.Attributes)
		if err != nil {
			return items, errors.Wrap(err, "Error parsing database row")
		}
		items = append(items, item)

	}
	return items, nil
}

// Fetch tries to find a repo item by its id. Returns
// an error if not found
func (repo *Sqlite3Repo) Fetch(item *RepoItem) (*RepoItem, error) {
	found := &RepoItem{}

	rows, err := repo.db.Query(repo.fetchStmt, item.Id)
	if err != nil {
		return found, errors.Wrap(err, repo.fetchStmt)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&found.Id, &found.Version, &found.Organisation, &found.Attributes)
		if err != nil {
			return found, errors.Wrap(err, "Error parsing database row")
		}

		return found, nil

	}

	// if we are here, this means no database row
	// was found
	return found, fmt.Errorf("DB_NOT_FOUND")
}

// Create a new item in the database
func (repo *Sqlite3Repo) Create(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(repo.createStmt)
	if err != nil {
		return item, errors.Wrap(err, repo.createStmt)
	}

	defer stmt.Close()

	// We ignore the version number from the repo item
	// and we set it to 0
	_, err = stmt.Exec(item.Id, 0, item.Organisation, "")
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		// TODO: use error codes from sqlite3 instead
		// of naively inspecting the error string representation
		errorCode := "DB_ERROR"
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			errorCode = "DB_CONFLICT"
		}
		return item, errors.Wrap(err, errorCode)
	}

	// This is a new item, we force its version to be 1
	item.Version = 0

	// Everything went fine. We return the item as is
	// for now
	return item, nil
}

// Update an existing item in the database. Returns the updated
// db item, or an error
func (repo *Sqlite3Repo) Update(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(repo.updateStmt)
	if err != nil {
		return item, errors.Wrap(err, repo.updateStmt)
	}

	defer stmt.Close()

	// Try to update the repo item
	// Increment the existing version before updating. This gives
	// some protection against concurrent updates and better
	// feedback to the client
	newVersion := item.Version + 1

	log.Printf("version %v => %v", item.Version, newVersion)

	res, err := stmt.Exec(item.Attributes, newVersion, item.Id, item.Version)
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		errorCode := "DB_ERROR"
		return item, errors.Wrap(err, errorCode)
	}

	// No rows affected. We treat this a a conflict.
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		errorCode := "DB_ERROR"
		return item, errors.Wrap(err, errorCode)
	}

	switch rowsAffected {
	case 0:
		return item, errors.New("DB_CONFLICT")
	case 1:
		item.Version = newVersion
		return item, nil
	default:
		// This should not happen, but we treat the case
		// for completeness
		return item, fmt.Errorf("DB_ERROR: more than 1 row affected by update: %v", rowsAffected)
	}
}

// Delete deletes the item from the repo. In this implementation,
// We simply mark the item as deleted. This is to make sure
// it's id is not reused by future payments
func (repo *Sqlite3Repo) Delete(item *RepoItem) error {
	stmt, err := repo.db.Prepare(repo.deleteOneStmt)
	if err != nil {
		return errors.Wrap(err, repo.deleteOneStmt)
	}

	defer stmt.Close()

	// Make sure we are deleting the item with the right
	// version
	res, err := stmt.Exec(item.Id, item.Version)
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		errorCode := "DB_ERROR"
		return errors.Wrap(err, errorCode)
	}

	// No rows affected. We treat this a a conflict.
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		errorCode := "DB_ERROR"
		return errors.Wrap(err, errorCode)
	}

	switch rowsAffected {
	case 0:
		// No rows affected means no item with
		// that id and version was found.
		return errors.New("DB_NOT_FOUND")
	case 1:
		// Everything went fine
		return nil
	default:
		// This should not happen, as we should be hitting
		// the primary key, still  we treat the case
		// for completeness
		return fmt.Errorf("DB_ERROR: more than 1 row affected by delete: %v", rowsAffected)
	}
}

// IsConflict returns true, if the given error denotes
// a database conflict
func (repo *Sqlite3Repo) IsConflict(err error) bool {
	return strings.Contains(err.Error(), "DB_CONFLICT")
}

// IsNotFound Ireturns true, if the given error denotes
// an item that was not found
func (repo *Sqlite3Repo) IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "DB_NOT_FOUND")
}

// DeleteAll hard delete all items. This operation cannot
// be recovered, so use with care
func (repo *Sqlite3Repo) DeleteAll() error {
	stmt, err := repo.db.Prepare(repo.deleteAllStmt)
	if err != nil {
		return errors.Wrap(err, repo.deleteAllStmt)
	}

	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		errorCode := "DB_ERROR"
		// TODO: better translate errors
		return errors.Wrap(err, errorCode)
	}

	return nil
}

// Info returns basic information about the current
// status of the repo
func (repo *Sqlite3Repo) Info() (RepoInfo, error) {
	var count int
	var info RepoInfo

	rows, err := repo.db.Query(repo.countStmt)
	if err != nil {
		return info, errors.Wrap(err, repo.countStmt)
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return info, errors.Wrap(err, repo.countStmt)
		}
		break
	}

	return RepoInfo{Count: count}, nil
}
