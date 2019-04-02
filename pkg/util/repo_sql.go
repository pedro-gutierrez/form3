// util provides with simple utility types and functions so that
// our main application package is less cluttered
package util

import (
	"database/sql"
	"fmt"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strings"
)

var (
	countStmt     string
	deleteAllStmt string
	listStmt      string
	fetchStmt     string
	createStmt    string
	updateStmt    string
	deleteOneStmt string
)

func init() {
	countStmt = "SELECT COUNT(*) FROM payments WHERE deleted = 0"
	deleteAllStmt = "DELETE FROM payments"
	listStmt = "SELECT id, version, organisation, attributes FROM payments WHERE deleted = 0 LIMIT $1 OFFSET $2"
	fetchStmt = "SELECT id, version, organisation, attributes FROM payments WHERE id = $1 AND deleted = 0"
	createStmt = "INSERT INTO payments (id, version, organisation, attributes) VALUES ($1, $2, $3, $4)"
	updateStmt = "UPDATE payments SET attributes=$1, version=$2 WHERE id=$3 AND version=$4"
	deleteOneStmt = "UPDATE payments SET deleted=1 WHERE id=$1 AND version=$2"
}

// A Generic SQL repo
type SqlRepo struct {
	db *sql.DB
}

// Close the database
func (repo *SqlRepo) Close() error {
	return repo.db.Close()
}

// Check performs a simple check on the database
func (repo *SqlRepo) Check() error {
	return repo.db.Ping()
}

// List Return a list of db items. Ignore items marked
// as deleted
func (repo *SqlRepo) List(offset int, limit int) ([]*RepoItem, error) {
	items := []*RepoItem{}

	rows, err := repo.db.Query(listStmt, limit, offset)
	if err != nil {
		return items, errors.Wrap(err, listStmt)
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
func (repo *SqlRepo) Fetch(item *RepoItem) (*RepoItem, error) {
	found := &RepoItem{}

	rows, err := repo.db.Query(fetchStmt, item.Id)
	if err != nil {
		return found, errors.Wrap(err, fetchStmt)
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
func (repo *SqlRepo) Create(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(createStmt)
	if err != nil {
		return item, errors.Wrap(err, createStmt)
	}

	defer stmt.Close()

	// We ignore the version number from the repo item
	// and we set it to 0
	_, err = stmt.Exec(item.Id, 0, item.Organisation, item.Attributes)
	if err != nil {
		// inspect the underlying database error
		// and translate it into something higher level
		// TODO: use error codes from sqlite3/postgres instead
		// of naively inspecting the error string representation
		errorCode := "DB_ERROR"
		if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
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
func (repo *SqlRepo) Update(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(updateStmt)
	if err != nil {
		return item, errors.Wrap(err, updateStmt)
	}

	defer stmt.Close()

	// Try to update the repo item
	// Increment the existing version before updating. This gives
	// some protection against concurrent updates and better
	// feedback to the client
	newVersion := item.Version + 1

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
func (repo *SqlRepo) Delete(item *RepoItem) error {
	stmt, err := repo.db.Prepare(deleteOneStmt)
	if err != nil {
		return errors.Wrap(err, deleteOneStmt)
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
// a database conflict. Note: this kind of error is something we generate
// and send to the web layer (it is not coming from the underlying sql library)
func (repo *SqlRepo) IsConflict(err error) bool {
	return strings.Contains(err.Error(), "DB_CONFLICT")
}

// IsNotFound Ireturns true, if the given error denotes
// an item that was not found. Note: this kind of error is something we generate
// and send to the web layer (it is not coming from the underlying sql library)
func (repo *SqlRepo) IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "DB_NOT_FOUND")
}

// DeleteAll hard delete all items. This operation cannot
// be recovered, so use with care
func (repo *SqlRepo) DeleteAll() error {
	stmt, err := repo.db.Prepare(deleteAllStmt)
	if err != nil {
		return errors.Wrap(err, deleteAllStmt)
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
func (repo *SqlRepo) Info() (RepoInfo, error) {
	var count int
	var info RepoInfo

	rows, err := repo.db.Query(countStmt)
	if err != nil {
		return info, errors.Wrap(err, countStmt)
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return info, errors.Wrap(err, countStmt)
		}
		break
	}

	return RepoInfo{Count: count}, nil
}
